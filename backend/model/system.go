package model

import (
	"gorm.io/gorm"
)

// SystemConfig 系统配置表
type SystemConfig struct {
	gorm.Model
	Key   string `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (SystemConfig) TableName() string {
	return "system_config"
}

// Admin 管理员表
type Admin struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"-"` // 密码不返回给前端
	Email    string `gorm:"type:varchar(100)" json:"email"`
	Status   int    `gorm:"default:1" json:"status"` // 1:正常 0:禁用
}

func (Admin) TableName() string {
	return "admins"
}

// Order 订单表
type Order struct {
	gorm.Model
	OrderNo     string  `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_no"`
	UserEmail   string  `gorm:"type:varchar(100)" json:"user_email"`
	Amount      float64 `gorm:"type:decimal(10,2)" json:"amount"`
	Status      string  `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending,processing,completed,cancelled,refunded
	CardCode    string  `gorm:"type:varchar(100)" json:"card_code"`
	CompletedAt *int64  `json:"completed_at"`
	Note        string  `gorm:"type:text" json:"note"`
}

func (Order) TableName() string {
	return "orders"
}
