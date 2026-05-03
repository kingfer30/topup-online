package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	redis "github.com/kingfer30/topup-online/config/cache"
	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/request"
)

var cdkBaseUrl = "https://kkk.ow800.com"

// 验证卡密
func VerifyCard(c *gin.Context) {
	param := constants.CDKVerifyRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}

	url := fmt.Sprintf("%s/api/cards/verify", cdkBaseUrl)
	err, resp := request.Curl(url, "POST", param)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "request fail: " + err.Error(),
			"data":    "",
		})
		return
	}
	defer resp.Body.Close()

	bodyByte, _ := io.ReadAll(resp.Body)
	logger.SysLogf("url: %s. body: %s", url, string(bodyByte))

	var result map[string]interface{}
	if err = json.Unmarshal(bodyByte, &result); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "response unmarshal fail: " + err.Error(),
			"data":    "",
		})
		return
	}

	success, _ := result["success"].(bool)
	if !success {
		msg := ""
		if m, ok := result["message"].(string); ok {
			msg = m
		}
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "card verify failed: " + msg,
			"data":    "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "success",
		"data":    result["data"],
	})
}

// 验证卡密并充值
func TopUp(c *gin.Context) {
	param := constants.CDKTopupRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}

	// 防重复提交：以 userEmail 为维度限流
	lockKey := fmt.Sprintf("Thread:%s", param.UserEmail)
	if count, serr := redis.Exists(lockKey); serr != nil || count == 0 {
		if ok, err := redis.SetNx(lockKey, "1", time.Duration(constants.GetCacheFrequency())*time.Second); ok || err == nil {
			status, data := doTopUp(param)
			c.JSON(status, data)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "Please do not submit repeated requests",
	})
}

// 执行充值请求
func doTopUp(param constants.CDKTopupRequest) (int, gin.H) {
	url := fmt.Sprintf("%s/api/cards/verify-gpt", cdkBaseUrl)
	err, resp := request.Curl(url, "POST", param)
	if err != nil {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "request fail: " + err.Error(),
			"data":    "",
		}
	}
	defer resp.Body.Close()

	bodyByte, _ := io.ReadAll(resp.Body)
	logger.SysLogf("url: %s. body: %s", url, string(bodyByte))

	var resData constants.CDKTopupResponse
	if err = json.Unmarshal(bodyByte, &resData); err != nil {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "response unmarshal fail: " + err.Error(),
			"data":    "",
		}
	}

	if !resData.Success {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "topup failed: " + resData.Data.Message,
			"data":    "",
		}
	}

	if resData.Data.TaskId == "" {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "topup failed: task not created",
			"data":    "",
		}
	}

	return http.StatusOK, gin.H{
		"success": true,
		"message": "success",
		"data": gin.H{
			"taskId":       resData.Data.TaskId,
			"processing":   resData.Data.Processing,
			"needsPolling": resData.Data.NeedsPolling,
			"message":      resData.Data.Message,
		},
	}
}

// 查询充值任务状态
func QueryTaskStatus(c *gin.Context) {
	param := constants.CDKQueryTaskRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}

	url := fmt.Sprintf("%s/api/recharge/query-task-status", cdkBaseUrl)
	err, resp := request.Curl(url, "POST", param)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "request fail: " + err.Error(),
			"data":    "",
		})
		return
	}
	defer resp.Body.Close()

	bodyByte, _ := io.ReadAll(resp.Body)
	logger.SysLogf("url: %s. body: %s", url, string(bodyByte))

	var resData constants.CDKQueryTaskResponse
	if err = json.Unmarshal(bodyByte, &resData); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "response unmarshal fail: " + err.Error(),
			"data":    "",
		})
		return
	}

	switch resData.Data.Status {
	case "success":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": resData.Data.Message,
			"data": gin.H{
				"status": "success",
			},
		})
	case "processing":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": resData.Data.Message,
			"data": gin.H{
				"status": "processing",
			},
		})
	case "failed":
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": resData.Data.Message,
			"data": gin.H{
				"status": "failed",
			},
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "unknown task status: " + resData.Data.Status,
			"data":    "",
		})
	}
}
