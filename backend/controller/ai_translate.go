package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
)

type aiTranslateRequest struct {
	Text       string `json:"text" binding:"required"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang" binding:"required"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string                  `json:"model"`
	Messages    []chatCompletionMessage `json:"messages"`
	Temperature float64                 `json:"temperature"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func chatCompletionsURL(base string) string {
	b := strings.TrimRight(strings.TrimSpace(base), "/")
	if b == "" {
		return ""
	}
	if strings.HasSuffix(b, "/chat/completions") {
		return b
	}
	if strings.HasSuffix(b, "/v1") {
		return b + "/chat/completions"
	}
	return b + "/v1/chat/completions"
}

func langLabel(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "zh", "zh-cn":
		return "中文（简体）"
	case "en":
		return "英语"
	case "ja":
		return "日语"
	case "ru":
		return "俄语"
	case "ko":
		return "韩语"
	case "fr":
		return "法语"
	case "de":
		return "德语"
	case "es":
		return "西班牙语"
	default:
		return code
	}
}

func buildTranslatePrompt(sourceLang, targetLang, text string) (systemPrompt, userContent string) {
	src := strings.TrimSpace(strings.ToLower(sourceLang))
	tgt := strings.TrimSpace(strings.ToLower(targetLang))
	if src == "" {
		src = "zh"
	}
	srcLabel := langLabel(src)
	tgtLabel := langLabel(tgt)

	systemPrompt = "你是专业翻译助手，只需要将我发送的内容直接翻译并输出译文本身，不要解释、不要前缀、不要用引号包裹整段译文。"
	userContent = fmt.Sprintf("将以下文本从「%s」翻译成「%s」。\n\n%s", srcLabel, tgtLabel, text)
	return systemPrompt, userContent
}

// AdminAITranslate 使用系统设置中的大模型配置进行翻译
func AdminAITranslate(c *gin.Context) {
	var req aiTranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请输入要翻译的文本"})
		return
	}

	modelName, baseURL, apiKey, err := model.GetAISettingsForTranslate()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "读取模型配置失败: " + err.Error()})
		return
	}
	if strings.TrimSpace(modelName) == "" || strings.TrimSpace(baseURL) == "" || strings.TrimSpace(apiKey) == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请先在系统设置-AI模型设置中配置模型名称、Base URL 和 API Key"})
		return
	}

	url := chatCompletionsURL(baseURL)
	if url == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "Base URL 无效"})
		return
	}

	systemPrompt, userContent := buildTranslatePrompt(req.SourceLang, req.TargetLang, text)

	body := openAIChatRequest{
		Model: modelName,
		Messages: []chatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		Temperature: 0.2,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "构造请求失败"})
		return
	}

	ctx := c.Request.Context()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建请求失败: " + err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "调用模型接口失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "读取响应失败: " + err.Error()})
		return
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "解析模型响应失败",
			"data":    gin.H{"raw": string(respBytes)},
		})
		return
	}

	if parsed.Error != nil && parsed.Error.Message != "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "模型返回错误: " + parsed.Error.Message,
		})
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": fmt.Sprintf("模型接口 HTTP %d", resp.StatusCode),
			"data":    gin.H{"body": string(respBytes)},
		})
		return
	}

	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "模型未返回有效译文",
			"data":    gin.H{"body": string(respBytes)},
		})
		return
	}

	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data": gin.H{
			"translation": out,
		},
	})
}
