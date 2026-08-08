package controller

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/model"
)

const (
	adImageWidth  = 160
	adImageHeight = 100
)

// GetAdConfigList 广告配置列表
func GetAdConfigList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	position := c.Query("position")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := model.GetAdConfigList(page, pageSize, keyword, position)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"list":      list,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAdConfigById 广告配置详情
func GetAdConfigById(c *gin.Context) {
	id, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	item, err := model.GetAdConfigById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "广告不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    item,
	})
}

type adConfigRequest struct {
	Title     string `json:"title"`
	Image     string `json:"image"`
	Link      string `json:"link"`
	Position  string `json:"position"`
	Buyer     string `json:"buyer"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func parseAdTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("时间为空")
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	// 毫秒时间戳
	if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
		if ms > 1e12 {
			return time.UnixMilli(ms), nil
		}
		return time.Unix(ms, 0), nil
	}
	return time.Time{}, fmt.Errorf("时间格式无效")
}

func bindAdConfig(req *adConfigRequest) (*model.AdConfig, error) {
	start, err := parseAdTime(req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("生效时间: %w", err)
	}
	end, err := parseAdTime(req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("失效时间: %w", err)
	}
	return &model.AdConfig{
		Title:     strings.TrimSpace(req.Title),
		Image:     strings.TrimSpace(req.Image),
		Link:      strings.TrimSpace(req.Link),
		Position:  strings.TrimSpace(req.Position),
		Buyer:     strings.TrimSpace(req.Buyer),
		Sort:      req.Sort,
		Status:    req.Status,
		StartTime: start,
		EndTime:   end,
	}, nil
}

// CreateAdConfig 创建广告配置
func CreateAdConfig(c *gin.Context) {
	var req adConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	item, err := bindAdConfig(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if err := item.Create(); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": item})
}

// UpdateAdConfig 更新广告配置
func UpdateAdConfig(c *gin.Context) {
	id, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	var req adConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	item, err := bindAdConfig(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
		return
	}
	item.Id = id
	if err := item.Update(); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": item})
}

// DeleteAdConfig 删除广告配置
func DeleteAdConfig(c *gin.Context) {
	id, err := parsePositiveInt64(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	if err := model.DeleteAdConfigById(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// UploadAdImage 上传广告图片（必须 160x100）
func UploadAdImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择图片文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "打开文件失败"})
		return
	}
	defer src.Close()

	img, format, err := image.Decode(src)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无法解析图片，请上传 PNG/JPG/GIF"})
		return
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w != adImageWidth || h != adImageHeight {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": fmt.Sprintf("图片必须为 %dx%d，当前为 %dx%d", adImageWidth, adImageHeight, w, h),
		})
		return
	}

	ext := "." + strings.ToLower(format)
	if ext == ".jpeg" {
		ext = ".jpg"
	}
	switch ext {
	case ".png", ".jpg", ".gif":
	default:
		ext = filepath.Ext(file.Filename)
		if ext == "" {
			ext = ".png"
		}
	}

	dir := filepath.Join(constants.GetDataDir(), "uploads", "ads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建目录失败"})
		return
	}

	filename := uuid.New().String() + ext
	dst := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "保存文件失败"})
		return
	}

	relPath := "/uploads/ads/" + filename
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传成功",
		"data": gin.H{
			"url":    relPath,
			"width":  adImageWidth,
			"height": adImageHeight,
		},
	})
}

func absoluteURL(c *gin.Context, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	host := c.Request.Host
	if fwd := c.GetHeader("X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	return scheme + "://" + host + path
}

func mapActiveAdItems(c *gin.Context, list []model.AdConfig) []gin.H {
	items := make([]gin.H, 0, len(list))
	for _, ad := range list {
		items = append(items, gin.H{
			"id":       ad.Id,
			"title":    ad.Title,
			"image":    absoluteURL(c, ad.Image),
			"link":     ad.Link,
			"position": ad.Position,
		})
	}
	return items
}

// GetActiveAds 公开：获取当前生效左右广告 + 顶部通知（最新一条）
func GetActiveAds(c *gin.Context) {
	now := time.Now()
	left, _ := model.ListActiveAdsByPosition("left", now)
	right, _ := model.ListActiveAdsByPosition("right", now)

	data := gin.H{
		"left":  mapActiveAdItems(c, left),
		"right": mapActiveAdItems(c, right),
	}
	if top, err := model.GetLatestActiveAdByPosition("top", now); err == nil && top != nil {
		data["top"] = gin.H{
			"id":       top.Id,
			"title":    top.Title,
			"link":     top.Link,
			"position": top.Position,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    data,
	})
}

// ClickAd 公开：广告点击统计 +1
func ClickAd(c *gin.Context) {
	var req struct {
		Id int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Id <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的广告ID"})
		return
	}
	if err := model.IncrementAdClickCount(req.Id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "成功"})
}

func parsePositiveInt64(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}
