// Package sanchuan 三川供应商对接实现。
// 导入此包（_ import）即可将三川注册到供应商注册表。
package sanchuan

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kingfer30/topup-online/constants"
	"github.com/kingfer30/topup-online/supplier"
	"github.com/kingfer30/topup-online/utils/logger"
	"github.com/kingfer30/topup-online/utils/request"
)

const baseURL = "https://kkk.ow800.com"

// Driver 三川供应商驱动
type Driver struct{}

func init() {
	supplier.Register(&Driver{})
}

func (d *Driver) Name() string {
	return "三川"
}

// VerifyCard 验证卡密是否有效
func (d *Driver) VerifyCard(cardInfo string) error {
	url := fmt.Sprintf("%s/api/cards/verify", baseURL)
	param := constants.CDKVerifyRequest{CardInfo: cardInfo}

	err, resp := request.Curl(url, "POST", param)
	if err != nil {
		return fmt.Errorf("request fail: %w", err)
	}
	defer resp.Body.Close()

	bodyByte, _ := io.ReadAll(resp.Body)
	logger.SysLogf("[sanchuan] VerifyCard body: %s", string(bodyByte))

	var result map[string]interface{}
	if err = json.Unmarshal(bodyByte, &result); err != nil {
		return fmt.Errorf("response unmarshal fail: %w", err)
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return fmt.Errorf("card verify failed: %s", msg)
	}
	return nil
}

// TopUp 发起充值
func (d *Driver) TopUp(param supplier.TopupParam) (*supplier.TopupResult, error) {
	url := fmt.Sprintf("%s/api/cards/verify-gpt", baseURL)
	body := constants.CDKTopupRequest{
		CardInfo:     param.CardInfo,
		UserEmail:    param.UserEmail,
		UserGptToken: param.UserGptToken,
		FullAuthData: param.FullAuthData,
		ProductId:    param.ProductId,
	}

	err, resp := request.Curl(url, "POST", body)
	if err != nil {
		return nil, fmt.Errorf("request fail: %w", err)
	}
	defer resp.Body.Close()

	bodyByte, _ := io.ReadAll(resp.Body)
	logger.SysLogf("[sanchuan] TopUp body: %s", string(bodyByte))

	var resData constants.CDKTopupResponse
	if err = json.Unmarshal(bodyByte, &resData); err != nil {
		return nil, fmt.Errorf("response unmarshal fail: %w", err)
	}

	if !resData.Success {
		return nil, fmt.Errorf("topup failed: %s", resData.Data.Message)
	}
	if resData.Data.TaskId == "" {
		return nil, fmt.Errorf("topup failed: task not created")
	}

	return &supplier.TopupResult{
		TaskId:       resData.Data.TaskId,
		Processing:   resData.Data.Processing,
		NeedsPolling: resData.Data.NeedsPolling,
		Message:      resData.Data.Message,
	}, nil
}

// QueryTaskStatus 查询充值任务状态
func (d *Driver) QueryTaskStatus(taskId string, productId int, cardInfo string) (*supplier.TaskStatusResult, error) {
	url := fmt.Sprintf("%s/api/recharge/query-task-status", baseURL)
	body := constants.CDKQueryTaskRequest{
		TaskId:    taskId,
		ProductId: productId,
		CardInfo:  cardInfo,
	}

	err, resp := request.Curl(url, "POST", body)
	if err != nil {
		return nil, fmt.Errorf("request fail: %w", err)
	}
	defer resp.Body.Close()

	bodyByte, _ := io.ReadAll(resp.Body)
	logger.SysLogf("[sanchuan] QueryTaskStatus body: %s", string(bodyByte))

	var resData constants.CDKQueryTaskResponse
	if err = json.Unmarshal(bodyByte, &resData); err != nil {
		return nil, fmt.Errorf("response unmarshal fail: %w", err)
	}

	return &supplier.TaskStatusResult{
		Status:  resData.Data.Status,
		Message: resData.Data.Message,
	}, nil
}
