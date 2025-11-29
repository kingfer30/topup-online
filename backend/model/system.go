package model

import (
	"gorm.io/gorm"
)

// SystemConfig 系统配置表
type SystemConfig struct {
	ID    int    `gorm:"primaryKey" json:"id"`
	Key   string `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
	gorm.Model
}

func (SystemConfig) TableName() string {
	return "system_config"
}

// Admin 管理员表
type Admin struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	Username string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"-"` // 密码不返回给前端
	Email    string `gorm:"type:varchar(100)" json:"email"`
	Status   int    `gorm:"default:1" json:"status"` // 1:正常 0:禁用
	gorm.Model
}

func (Admin) TableName() string {
	return "admins"
}

// Order 订单表
type Order struct {
	ID          int     `gorm:"primaryKey" json:"id"`
	OrderNo     string  `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_no"`
	UserEmail   string  `gorm:"type:varchar(100)" json:"user_email"`
	Amount      float64 `gorm:"type:decimal(10,2)" json:"amount"`
	Status      string  `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending,processing,completed,cancelled,refunded
	CardCode    string  `gorm:"type:varchar(100)" json:"card_code"`
	CompletedAt *int64  `json:"completed_at"`
	Note        string  `gorm:"type:text" json:"note"`
	gorm.Model
}

func (Order) TableName() string {
	return "orders"
}

// Card 卡密表
type Card struct {
	ID        int    `gorm:"primaryKey" json:"id"`
	Code      string `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Type      string `gorm:"type:varchar(50)" json:"type"`     // 卡密类型
	Value     int    `gorm:"default:0" json:"value"`           // 面值（天数）
	Status    int    `gorm:"default:0" json:"status"`          // 0:未使用 1:已使用 2:已过期
	UsedBy    string `gorm:"type:varchar(100)" json:"used_by"` // 使用者邮箱
	UsedAt    *int64 `json:"used_at"`                          // 使用时间
	ExpiredAt *int64 `json:"expired_at"`                       // 过期时间
	gorm.Model
}

func (Card) TableName() string {
	return "cards"
}
