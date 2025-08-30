package constants

type CDKStatusResponse struct {
	Available bool   `json:"available"`
	Error     string `json:"error"`
}

type CDKCheckRequest struct {
	AccessToken string `json:"access_token"`
}

type CDKCheckResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CDKTopupRequest struct {
	CardKey     string `json:"card_key"`
	Account     string `json:"account"`
	AccessToken string `json:"access_token"`
}

type CDKTopupResponse struct {
	Success bool   `json:"success"`
	TaskId  string `json:"task_id"`
	Error   string `json:"error"`
}

type CDKTaskResponse struct {
	Status string `json:"status"`
	Result string `json:"result"`
	Error  string `json:"error"`
}
