package model

import (
	"time"
)

// DigisellerPrice Digiseller订阅类型今日售价配置
type DigisellerPrice struct {
	Id               int       `json:"id" gorm:"primaryKey;autoIncrement"`
	SubscriptionType string    `json:"subscription_type" gorm:"type:varchar(30);uniqueIndex;not null;comment:订阅类型"`
	Price            float64   `json:"price" gorm:"type:decimal(10,2);default:0;comment:今日售价"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (DigisellerPrice) TableName() string {
	return "digiseller_prices"
}

// GetAllDigisellerPrices 获取所有Digiseller价格配置
func GetAllDigisellerPrices() ([]DigisellerPrice, error) {
	var prices []DigisellerPrice
	if err := DB.Order("id asc").Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

// UpsertDigisellerPrice 新增或更新某订阅类型的售价
func UpsertDigisellerPrice(subscriptionType string, price float64) error {
	var existing DigisellerPrice
	result := DB.Where("subscription_type = ?", subscriptionType).First(&existing)
	if result.Error != nil {
		// 不存在则创建
		record := DigisellerPrice{
			SubscriptionType: subscriptionType,
			Price:            price,
		}
		return DB.Create(&record).Error
	}
	// 已存在则更新
	return DB.Model(&existing).Update("price", price).Error
}
