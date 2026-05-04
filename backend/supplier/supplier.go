package supplier

import "sort"

// TopupParam 充值参数
type TopupParam struct {
	CardInfo     string `json:"cardInfo"`
	UserEmail    string `json:"userEmail"`
	UserGptToken string `json:"userGptToken"`
	FullAuthData string `json:"fullAuthData"`
	ProductId    *int   `json:"productId,omitempty"`
}

// TopupResult 充值结果
type TopupResult struct {
	TaskId       string `json:"taskId"`
	Processing   bool   `json:"processing"`
	NeedsPolling bool   `json:"needsPolling"`
	Message      string `json:"message"`
}

// TaskStatusResult 任务状态结果
type TaskStatusResult struct {
	Status  string `json:"status"` // success | processing | failed
	Message string `json:"message"`
}

// Driver 供应商接口。新增供应商需实现此接口，在其 init() 中调用 Register 注册，
// 并在 main.go 中以 _ import 该子包触发注册。
type Driver interface {
	// Name 返回供应商唯一名称（用于下拉显示和 gpt_cards.supplier 字段存储）
	Name() string
	// VerifyCard 验证卡密是否有效；返回 nil 表示有效，error 表示已使用/无效
	VerifyCard(cardInfo string) error
	// TopUp 发起充值，返回任务信息
	TopUp(param TopupParam) (*TopupResult, error)
	// QueryTaskStatus 查询充值任务状态
	QueryTaskStatus(taskId string, productId int, cardInfo string) (*TaskStatusResult, error)
}

var registry = map[string]Driver{}

// Register 注册供应商，在各供应商子包的 init() 中调用
func Register(d Driver) {
	registry[d.Name()] = d
}

// Names 返回已注册的供应商名称列表（有序）
func Names() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// Get 根据名称获取供应商驱动
func Get(name string) (Driver, bool) {
	d, ok := registry[name]
	return d, ok
}
