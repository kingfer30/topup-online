package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ChatShareLoginRequest ChatShare 登录请求结构
type ChatShareLoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"` // 加密后的密码
	Timestamp string `json:"timestamp"`
	Sign      string `json:"sign"`
}

// ChatShareLoginResponse ChatShare 登录响应结构
type ChatShareLoginResponse struct {
	IsSuccess bool   `json:"isSuccess"`
	Msg       string `json:"msg"`
	RespData  string `json:"respData"`
}

// BuildChatShareLoginRequest 构建 ChatShare 登录请求体
// 参数:
//   - username: 用户名
//   - password: 原始密码（会自动加密）
//
// 返回:
//   - *ChatShareLoginRequest: 构建好的请求体（密码已加密）
//   - error: 错误信息
func BuildChatShareLoginRequest(username, password string) (*ChatShareLoginRequest, error) {
	// 加密密码
	encryptedPassword, err := EncryptPasswordAESCFB(password)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败: %v", err)
	}

	// 构建请求体
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	return &ChatShareLoginRequest{
		Username:  username,
		Password:  encryptedPassword,
		Timestamp: timestamp,
		Sign:      "", // 如果需要签名，在这里添加签名逻辑
	}, nil
}

// CallChatShareLogin 调用 ChatShare 登录接口
// 参数:
//   - username: 用户名
//   - password: 原始密码
//   - baseURL: ChatShare API 基础URL（例如: "https://node3.chatshare.biz"）
//
// 返回:
//   - *ChatShareLoginResponse: 登录响应
//   - error: 错误信息
func CallChatShareLogin(username, password, baseURL string) (*ChatShareLoginResponse, error) {
	// 1. 构建请求体
	loginReq, err := BuildChatShareLoginRequest(username, password)
	if err != nil {
		return nil, err
	}

	// 2. 序列化为 JSON
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 3. 构建完整 URL
	url := baseURL + "/share-login/v1/user/auth/login"

	// 4. 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 5. 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	// 6. 执行请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 7. 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 8. 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 状态码错误: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 9. 解析响应
	var loginResp ChatShareLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}

	return &loginResp, nil
}
