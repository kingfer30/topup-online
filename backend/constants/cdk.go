package constants

// 验证卡密请求
type CDKVerifyRequest struct {
	CardInfo string `json:"cardInfo" binding:"required"`
}

// 充值请求
type CDKTopupRequest struct {
	CardInfo     string `json:"cardInfo" binding:"required"`
	UserEmail    string `json:"userEmail" binding:"required"`
	UserGptToken string `json:"userGptToken" binding:"required"`
	FullAuthData string `json:"fullAuthData" binding:"required"`
	ProductId    *int   `json:"productId,omitempty"`
}

// 充值响应 data
type CDKTopupData struct {
	TaskId       string      `json:"taskId"`
	Processing   bool        `json:"processing"`
	NeedsPolling bool        `json:"needsPolling"`
	UsageId      interface{} `json:"usageId"`
	CardId       int         `json:"cardId"`
	Success      bool        `json:"success"`
	Message      string      `json:"message"`
}

// 充值响应
type CDKTopupResponse struct {
	Code    int          `json:"code"`
	Success bool         `json:"success"`
	Data    CDKTopupData `json:"data"`
}

// 查询任务状态请求
type CDKQueryTaskRequest struct {
	TaskId    string `json:"taskId" binding:"required"`
	ProductId int    `json:"productId"`
	CardInfo  string `json:"cardInfo,omitempty"`
}

// 查询任务状态响应 data
type CDKQueryTaskData struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// 查询任务状态响应
type CDKQueryTaskResponse struct {
	Code    int              `json:"code"`
	Success bool             `json:"success"`
	Data    CDKQueryTaskData `json:"data"`
}
