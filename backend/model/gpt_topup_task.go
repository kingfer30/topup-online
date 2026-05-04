package model

import (
	"time"

	"gorm.io/gorm"
)

// GptTopupTask 充值任务
type GptTopupTask struct {
	Id             int            `json:"id" gorm:"primaryKey;autoIncrement"`
	CdkId          int            `json:"cdk_id" gorm:"not null;comment:关联gpt_cdk.id;index:idx_cdk_id"`
	CdkKey         string         `json:"cdk_key" gorm:"type:varchar(200);comment:CDK密钥快照"`
	CardId         *int           `json:"card_id" gorm:"comment:关联gpt_cards.id"`
	SupplierTaskId string         `json:"supplier_task_id" gorm:"type:varchar(100);comment:供应商任务ID"`
	UserEmail      string         `json:"user_email" gorm:"type:varchar(100);comment:充值目标邮箱"`
	Status         int            `json:"status" gorm:"type:tinyint(2);default:0;comment:状态 0待处理 1处理中 2成功 3失败;index:idx_status"`
	Message        string         `json:"message" gorm:"type:varchar(500);comment:结果消息"`
	CompletedAt    *time.Time     `json:"completed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (GptTopupTask) TableName() string {
	return "gpt_topup_task"
}

// CreateTopupTask 创建充值任务
func CreateTopupTask(task *GptTopupTask) error {
	return DB.Create(task).Error
}

// UpdateTopupTask 更新任务
func UpdateTopupTask(id int, updates map[string]interface{}) error {
	return DB.Model(&GptTopupTask{}).Where("id = ?", id).Updates(updates).Error
}

// GetTopupTaskById 按ID查询任务
func GetTopupTaskById(id int) (*GptTopupTask, error) {
	var task GptTopupTask
	err := DB.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}
