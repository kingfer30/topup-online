package controller

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/middleware"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/logger"
	"gorm.io/gorm"
)

// RoomProxy 房间反向代理
func RoomProxy(c *gin.Context) {
	if db == nil {
		c.String(http.StatusInternalServerError, "数据库未初始化")
		return
	}

	roomID := c.Param("id")
	logger.SysLog(fmt.Sprintf("房间ID: %s", roomID))

	// 获取当前登录用户
	user, exists := middleware.GetUser(c)
	if !exists {
		c.String(http.StatusUnauthorized, "未授权访问，请先登录")
		return
	}

	// 检查用户是否绑定了镜像卡密
	if user.MirrorCardId == 0 {
		c.String(http.StatusForbidden, "用户未绑定镜像卡密")
		return
	}

	// 查询镜像卡密
	var mirrorCard model.MirrorCard
	if err := db.First(&mirrorCard, user.MirrorCardId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusForbidden, "镜像卡密不存在")
			return
		}
		c.String(http.StatusInternalServerError, "查询镜像卡密失败")
		return
	}

	// 检查卡密状态
	if mirrorCard.Status != model.MirrorCardStatusEnabled || mirrorCard.Token == "" {
		c.String(http.StatusForbidden, "镜像卡密未启用或Token不存在")
		return
	}

	// 首次进入房间，需要调用 enter 接口获取真实的目标URL
	// 检查是否是第一次访问（通过检查浏览器是否有目标站点的 Cookie）
	targetCookie := ""
	cookieHeader := c.GetHeader("Cookie")

	// 从浏览器的 Cookie 中提取目标站点的 Cookie（以 _mirror_ 前缀标识）
	if cookieHeader != "" {
		cookies := strings.Split(cookieHeader, "; ")
		var targetCookies []string
		for _, cookie := range cookies {
			if strings.HasPrefix(cookie, "_mirror_") {
				// 去除前缀，恢复原始 Cookie
				targetCookies = append(targetCookies, strings.TrimPrefix(cookie, "_mirror_"))
			}
		}
		if len(targetCookies) > 0 {
			targetCookie = strings.Join(targetCookies, "; ")
		}
	}

	var targetURL string
	needEnterAPI := targetCookie == ""

	if needEnterAPI {
		logger.SysLog("首次访问，调用 enter 接口")
		// 调用 enter 接口
		enterURL := fmt.Sprintf("%s/share-login/v1/user/home/enter", mirrorCard.NodeURL)

		timestamp := fmt.Sprintf("%d", getTimestampMillis())
		requestBody := fmt.Sprintf(`{"channel":"xy","car_id":"%s","timestamp":"%s","sign":""}`, roomID, timestamp)

		req, err := http.NewRequest("POST", enterURL, strings.NewReader(requestBody))
		if err != nil {
			c.String(http.StatusInternalServerError, "创建请求失败")
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", "token="+mirrorCard.Token)

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // 不自动跟随重定向
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			c.String(http.StatusInternalServerError, "请求失败")
			return
		}
		defer resp.Body.Close()

		// 收集响应中的所有 Cookie，并发送给浏览器
		var cookieBuilder strings.Builder
		cookieBuilder.WriteString("token=" + mirrorCard.Token)

		for _, cookie := range resp.Cookies() {
			cookieBuilder.WriteString("; ")
			cookieBuilder.WriteString(cookie.Name)
			cookieBuilder.WriteString("=")
			cookieBuilder.WriteString(cookie.Value)

			// 将目标站点的 Cookie 通过 Set-Cookie 发送给浏览器
			// 使用前缀 _mirror_ 来标识这是代理的 Cookie
			c.SetCookie(
				"_mirror_"+cookie.Name, // name
				cookie.Value,           // value
				cookie.MaxAge,          // maxAge
				"/",                    // path
				"",                     // domain
				false,                  // secure
				false,                  // httpOnly
			)
		}
		targetCookie = cookieBuilder.String()
		logger.SysLog(fmt.Sprintf("收集的所有 Cookie: %s", targetCookie))

		bodyBytes, _ := io.ReadAll(resp.Body)

		var apiResp struct {
			IsSuccess bool   `json:"isSuccess"`
			Msg       string `json:"msg"`
			RespData  string `json:"respData"`
		}

		if err := json.Unmarshal(bodyBytes, &apiResp); err != nil || !apiResp.IsSuccess {
			c.String(http.StatusInternalServerError, "获取房间URL失败")
			return
		}

		targetURL = apiResp.RespData
	} else {
		logger.SysLog("后续访问，使用已有 Cookie")
		// 后续访问，直接使用 NodeURL 作为基础URL
		targetURL = mirrorCard.NodeURL
	}

	logger.SysLog(fmt.Sprintf("目标URL: %s", targetURL))

	// 解析目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "无效的目标URL")
		return
	}

	logger.SysLog(fmt.Sprintf("目标URL: %s", targetURL))
	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director，修改请求
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 设置目标Host
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// 去除路径前缀 /api/rooms/:id
		// 例如: /api/rooms/PLUS-8/some/path -> /some/path
		//      /api/rooms/PLUS-8 -> /
		proxyPath := c.Param("proxyPath")
		if proxyPath != "" {
			// 有子路径，使用子路径
			req.URL.Path = proxyPath
		} else {
			// 没有子路径，访问根路径
			req.URL.Path = "/"
		}

		// 添加所有 Cookie（包括从 enter 接口获取的）
		if targetCookie == "" {
			// 如果没有收集到 Cookie，至少使用 token
			targetCookie = "token=" + mirrorCard.Token
		}
		req.Header.Set("Cookie", targetCookie)
		logger.SysLog(fmt.Sprintf("转发 Cookie: %s", targetCookie))

		// 设置其他必要的请求头
		req.Header.Set("Referer", targetURL)
		req.Header.Set("Origin", fmt.Sprintf("%s://%s", target.Scheme, target.Host))

		// 修改 Accept-Encoding，只接受 gzip，不接受 br (Brotli)
		// 因为 Go 标准库不支持 Brotli 解压
		req.Header.Set("Accept-Encoding", "gzip, deflate")

		logger.SysLog(fmt.Sprintf("代理请求: %s %s", req.Method, req.URL.String()))
	}

	// 自定义ModifyResponse，修改响应内容
	proxy.ModifyResponse = func(resp *http.Response) error {
		// 处理重定向：重写 Location 头
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			location := resp.Header.Get("Location")
			if location != "" {
				logger.SysLog(fmt.Sprintf("原始重定向: %s", location))

				// 解析 Location URL
				locationURL, err := url.Parse(location)
				if err == nil {
					// 如果是相对路径或者是目标域名的URL，需要重写
					if locationURL.Host == "" || locationURL.Host == target.Host {
						// 构建新的代理路径
						newPath := locationURL.Path
						if locationURL.RawQuery != "" {
							newPath += "?" + locationURL.RawQuery
						}

						// 重写为代理URL
						proxyLocation := fmt.Sprintf("/api/rooms/%s%s", roomID, newPath)
						resp.Header.Set("Location", proxyLocation)
						logger.SysLog(fmt.Sprintf("重写重定向: %s", proxyLocation))
					}
				}
			}
		}

		contentType := resp.Header.Get("Content-Type")

		// 只处理HTML、JS、CSS内容
		if strings.Contains(contentType, "text/html") ||
			strings.Contains(contentType, "text/javascript") ||
			strings.Contains(contentType, "application/javascript") ||
			strings.Contains(contentType, "text/css") {

			logger.SysLog(fmt.Sprintf("处理响应: Content-Type=%s, Content-Encoding=%s",
				contentType, resp.Header.Get("Content-Encoding")))

			// 先读取原始字节
			rawBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body.Close()

			logger.SysLog(fmt.Sprintf("原始字节长度: %d, 前100字节: %v",
				len(rawBytes), rawBytes[:min(100, len(rawBytes))]))

			var bodyBytes []byte

			// 检查压缩类型
			contentEncoding := resp.Header.Get("Content-Encoding")
			switch contentEncoding {
			case "gzip":
				logger.SysLog("检测到 gzip 压缩，开始解压")
				gzipReader, err := gzip.NewReader(bytes.NewReader(rawBytes))
				if err != nil {
					logger.SysLog(fmt.Sprintf("gzip 解压失败: %v", err))
					return err
				}
				defer gzipReader.Close()

				bodyBytes, err = io.ReadAll(gzipReader)
				if err != nil {
					logger.SysLog(fmt.Sprintf("读取解压内容失败: %v", err))
					return err
				}
				logger.SysLog(fmt.Sprintf("解压后长度: %d", len(bodyBytes)))

			case "br":
				logger.SysLog("检测到 Brotli 压缩，但 Go 标准库不支持，返回错误")
				return fmt.Errorf("不支持 Brotli (br) 压缩格式")

			case "deflate":
				logger.SysLog("检测到 deflate 压缩（暂不支持）")
				return fmt.Errorf("不支持 deflate 压缩格式")

			case "":
				logger.SysLog("未压缩内容，直接使用")
				bodyBytes = rawBytes

			default:
				logger.SysLog(fmt.Sprintf("未知压缩格式: %s", contentEncoding))
				bodyBytes = rawBytes
			}

			// 打印内容前500字符
			content := string(bodyBytes)
			logger.SysLog(fmt.Sprintf("内容前500字符: %s", content[:min(500, len(content))]))

			// 替换内容中的URL
			content = rewriteURLs(content, target.String(), c.Request.Host, roomID)

			// 更新响应
			newBody := []byte(content)
			resp.Body = io.NopCloser(bytes.NewReader(newBody))
			resp.ContentLength = int64(len(newBody))
			resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))

			// 移除压缩相关的响应头，因为我们返回的是未压缩的内容
			resp.Header.Del("Content-Encoding")
			resp.Header.Del("Transfer-Encoding")
		}

		return nil
	}

	// 执行代理
	proxy.ServeHTTP(c.Writer, c.Request)
}

// rewriteURLs 重写内容中的URL
func rewriteURLs(content, targetBase, proxyHost, roomID string) string {
	// 解析目标URL
	target, err := url.Parse(targetBase)
	if err != nil {
		return content
	}

	targetScheme := target.Scheme
	targetHost := target.Host

	// 构建代理前缀
	proxyPrefix := fmt.Sprintf("/api/rooms/%s", roomID)

	// 替换绝对URL (https://target.com/path -> /api/rooms/:id/path)
	content = regexp.MustCompile(fmt.Sprintf(`%s://%s(/[^"'\s)>]*)`, targetScheme, regexp.QuoteMeta(targetHost))).
		ReplaceAllString(content, proxyPrefix+"$1")

	// 替换协议相对URL (//target.com/path -> /api/rooms/:id/path)
	content = regexp.MustCompile(fmt.Sprintf(`//%s(/[^"'\s)>]*)`, regexp.QuoteMeta(targetHost))).
		ReplaceAllString(content, proxyPrefix+"$1")

	// 替换根路径相对URL，但要小心不要替换已经被替换过的
	// 例如: "/some/path" -> "/api/rooms/:id/some/path"
	// 但不要替换 "/api/rooms/:id/xxx"
	content = regexp.MustCompile(`(href|src|action|url)\s*=\s*["']/([^"'\s>]*)`).
		ReplaceAllStringFunc(content, func(match string) string {
			// 提取路径部分
			re := regexp.MustCompile(`(href|src|action|url)\s*=\s*["']/([^"'\s>]*)`)
			if submatches := re.FindStringSubmatch(match); len(submatches) > 2 {
				attr := submatches[1]
				path := submatches[2]

				// 如果路径已经是 /api/rooms/ 开头，不再替换
				if strings.HasPrefix(path, "api/rooms/") {
					return match
				}

				return fmt.Sprintf(`%s="%s/%s`, attr, proxyPrefix, path)
			}
			return match
		})

	return content
}

// getTimestampMillis 获取当前时间戳（毫秒）
func getTimestampMillis() int64 {
	return time.Now().UnixMilli()
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
