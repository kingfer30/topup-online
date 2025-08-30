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
	"github.com/kingfer30/topup-online/utils/random"
	"github.com/kingfer30/topup-online/utils/request"
)

var baseUrl = "https://api.ow520.com"

// 获取卡密详情
func GetCardDetail(c *gin.Context) {
	number := c.Param("number")
	if number == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "please set the parameter of number ",
			"data":    "",
		})
		return
	}
	url := fmt.Sprintf("%s/api/card-keys/%s", baseUrl, number)
	err, resp := request.Curl(url, "GET", nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "request fail: " + err.Error(),
			"data":    "",
		})
		return
	}

	defer resp.Body.Close()
	var cdkData constants.CDKStatusResponse
	bodyByte, _ := io.ReadAll(resp.Body)

	logger.SysLogf("url: %s. body: %s", url, string(bodyByte))
	err = json.Unmarshal(bodyByte, &cdkData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "response unmarshal fail: " + err.Error(),
			"data":    "",
		})
		return
	}
	if cdkData.Available == false {
		errMsg := "this cdks status is abnormal"
		switch cdkData.Error {
		case "卡密不存在":
			errMsg = "this cdk is not exists."
		case "卡密已使用":
			errMsg = "this cdk has been used."
		case "卡密已停用":
			errMsg = "this cdk has been suspended."
		}
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": errMsg,
			"data":    "",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "success",
		"data":    cdkData,
	})
	return
}

// 用户信息校验
func Check(c *gin.Context) {
	param := constants.CDKCheckRequest{}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}
	url := fmt.Sprintf("%s/api/parse-token", baseUrl)
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
	var accData constants.CDKCheckResponse
	bodyByte, _ := io.ReadAll(resp.Body)

	logger.SysLogf("url: %s. body: %s", url, string(bodyByte))
	err = json.Unmarshal(bodyByte, &accData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "response unmarshal fail: " + err.Error(),
			"data":    "",
		})
		return
	}
	if accData.Success == false {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to get this account information: " + accData.Message,
			"data":    "",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "success",
		"data": gin.H{
			"user_name": accData.Message,
		},
	})
	return
}

// 充值任务
func TopUp(c *gin.Context) {
	param := constants.CDKTopupRequest{}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parameter error: " + err.Error(),
		})
		return
	}
	if param.Account == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "please set the parameter of account",
		})
		return
	}
	if count, serr := redis.Exists(fmt.Sprintf("Thread:%s", param.Account)); serr != nil || count == 0 {
		if ok, err := redis.SetNx(fmt.Sprintf("Thread:%s", param.Account), "1", time.Duration(constants.CacheFrequency)*time.Second); ok || err == nil {
			status, data := createThread(c, param)
			c.JSON(status, data)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "Please do not submit repeated requests",
	})
	return
}

// 创建任务
func createThread(c *gin.Context, param constants.CDKTopupRequest) (int, gin.H) {
	url := fmt.Sprintf("%s/api/tasks", baseUrl)
	err, resp := request.Curl(url, "POST", param)
	if err != nil {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "request fail: " + err.Error(),
			"data":    "",
		}

	}

	defer resp.Body.Close()
	var resData constants.CDKTopupResponse
	bodyByte, _ := io.ReadAll(resp.Body)

	logger.SysLogf("url: %s. body: %s", url, string(bodyByte))
	err = json.Unmarshal(bodyByte, &resData)
	if err != nil {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "response unmarshal fail: " + err.Error(),
			"data":    "",
		}
	}
	if resData.Success == false {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "failed to top up: " + resData.Error,
			"data":    "",
		}
	}
	if resData.TaskId == "" {
		return http.StatusOK, gin.H{
			"success": false,
			"message": "failed to top up:  failed to create task",
			"data":    "",
		}
	}
	var threadId = random.GetRandomString(32)
	redis.SetNx(threadId, resData.TaskId, time.Duration(constants.CacheFrequency)*time.Second)
	return http.StatusOK, gin.H{
		"success": true,
		"message": "success",
		"data": gin.H{
			"thread_id": threadId,
		},
	}
}

// 获取任务详情
func GetThreadDetail(c *gin.Context) {
	threadId := c.Param("id")
	if threadId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "please set the parameter of thread-id ",
			"data":    "",
		})
		return
	}
	taskId, err := redis.Get(threadId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to get this thread: failed to find thread-id",
			"data":    "",
		})
		return
	}
	if taskId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to get this thread: thread-id not found",
			"data":    "",
		})
		return
	}
	url := fmt.Sprintf("%s/api/tasks/%s", baseUrl, taskId)
	err, resp := request.Curl(url, "GET", nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "request fail: " + err.Error(),
			"data":    "",
		})
		return
	}

	defer resp.Body.Close()
	var resData constants.CDKTaskResponse
	bodyByte, _ := io.ReadAll(resp.Body)

	logger.SysLogf("url: %s. body: %s", url, string(bodyByte))
	err = json.Unmarshal(bodyByte, &resData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "response unmarshal fail: " + err.Error(),
			"data":    "",
		})
		return
	}
	var data gin.H
	switch resData.Status {
	case "failed":
		data = gin.H{
			"success": false,
			"message": "failed to top-up: " + resData.Error,
			"data":    "",
		}
	case "pending":
		data = gin.H{
			"success": true,
			"message": "current thread is pending",
			"data": gin.H{
				"status": "pending",
			},
		}
	case "processing":
		data = gin.H{
			"success": false,
			"message": "current thread is processing",
			"data": gin.H{
				"status": "processing",
			},
		}
	case "completed":
		data = gin.H{
			"success": true,
			"message": "success",
			"data": gin.H{
				"status": "completed",
			},
		}
	}

	c.JSON(http.StatusOK, data)
	return
}
