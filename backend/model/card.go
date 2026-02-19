package model

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AccountCard 账号卡密模型
type AccountCard struct {
	Id                      int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Account                 string         `json:"account" gorm:"type:varchar(100);not null;uniqueIndex:idx_account"`
	Password                string         `json:"password" gorm:"type:varchar(50)"`
	MailPassword            string         `json:"mail_password" gorm:"type:varchar(50)"`
	SubscriptionStatus      int            `json:"subscription_status" gorm:"type:tinyint(2);default:1;comment:订阅状态 1已订阅 2未订阅"`
	SubscriptionType        string         `json:"subscription_type" gorm:"type:varchar(30);comment:订阅类型;index:idx_subscription_type"`
	SubscriptionTime        *int64         `json:"subscription_time" gorm:"type:bigint(20);comment:订阅时间;index:idx_subscription_time"`
	SubscriptionExpiredTime *int64         `json:"subscription_expired_time" gorm:"type:bigint(20);comment:订阅过期时间"`
	PurchaseDate            *int64         `json:"purchase_date" gorm:"type:bigint(20);comment:购买时间"`
	PurchasePrice           *float64       `json:"purchase_price" gorm:"type:decimal(10,2);comment:购买价格(成本)"`
	PurchaseFrom            string         `json:"purchase_from" gorm:"type:varchar(50);comment:购买平台"`
	PurchaseOrderNo         string         `json:"purchase_order_no" gorm:"type:varchar(100);comment:购买订单号"`
	SellPrice               *float64       `json:"sell_price" gorm:"type:decimal(10,2);comment:出售价格"`
	SellDate                *int64         `json:"sell_date" gorm:"type:bigint(20);comment:出售时间"`
	SellTo                  string         `json:"sell_to" gorm:"type:varchar(50);comment:出售对方"`
	SellOrderNo             string         `json:"sell_order_no" gorm:"type:varchar(100);comment:出售订单号"`
	SellStatus              int            `json:"sell_status" gorm:"type:tinyint(2);default:1;comment:出售状态 1 未出售 2发货中 3已出售;index:idx_subscription_sell"`
	AccountType             int            `json:"account_type" gorm:"type:tinyint(2);default:1;comment:账号类型 1普号 2成品"`
	Status                  int            `json:"status" gorm:"type:tinyint(2);default:1;comment:状态 1 正常 2 禁用"`
	ApiKey                  string         `json:"api_key" gorm:"type:varchar(300);comment:apikey"`
	Token                   string         `json:"token" gorm:"type:text;comment:token"`
	TwoFA                   string         `json:"2fa" gorm:"column:2fa;type:varchar(100);comment:2fa"`
	MailUrl                 string         `json:"mail_url" gorm:"type:varchar(50);comment:邮箱地址"`
	Remark                  string         `json:"remark" gorm:"type:varchar(100);comment:备注"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// GetCardList 获取卡密列表
func GetCardList(tableName string, cardType string, page, pageSize int, keyword string) ([]AccountCard, int64, error) {
	var cards []AccountCard
	var total int64

	// 构建查询
	query := DB.Table(tableName)

	// 根据类型筛选
	switch cardType {
	case "all":
		query = query.Where("account_type = ?", 1) // 普号列表：只显示普号
	case "unsold":
		query = query.Where("sell_status = ?", 1) // 未售列表：未出售
	case "sold":
		query = query.Where("sell_status = ?", 3) // 已售列表：已出售
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("account LIKE ? OR mail_url LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

// GetCardById 根据ID获取卡密
func GetCardById(tableName string, id int) (*AccountCard, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	var card AccountCard
	err := DB.Table(tableName).First(&card, "id = ?", id).Error
	return &card, err
}

// CreateCard 创建卡密
func CreateCard(tableName string, card *AccountCard) error {
	if card.Account == "" {
		return errors.New("账号不能为空")
	}

	// 检查账号是否已存在
	var count int64
	DB.Table(tableName).Where("account = ?", card.Account).Count(&count)
	if count > 0 {
		return errors.New("账号已存在")
	}

	return DB.Table(tableName).Create(card).Error
}

// UpdateCard 更新卡密
func UpdateCard(tableName string, id int, card *AccountCard) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	if card.Account == "" {
		return errors.New("账号不能为空")
	}

	// 检查账号是否与其他卡密冲突
	var count int64
	DB.Table(tableName).Where("account = ? AND id != ?", card.Account, id).Count(&count)
	if count > 0 {
		return errors.New("账号已存在")
	}

	card.Id = id
	return DB.Table(tableName).Where("id = ?", id).Updates(card).Error
}

// DeleteCard 删除卡密（软删除）
func DeleteCard(tableName string, id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}

	return DB.Table(tableName).Where("id = ?", id).Delete(&AccountCard{}).Error
}

// BatchCreateCards 批量创建卡密
func BatchCreateCards(tableName string, cards []AccountCard) error {
	if len(cards) == 0 {
		return errors.New("没有要导入的数据")
	}

	// 使用事务批量插入
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, card := range cards {
			// 检查账号是否已存在
			var count int64
			tx.Table(tableName).Where("account = ?", card.Account).Count(&count)
			if count > 0 {
				// 跳过已存在的账号
				continue
			}

			if err := tx.Table(tableName).Create(&card).Error; err != nil {
				return fmt.Errorf("创建卡密失败: %s, %v", card.Account, err)
			}
		}
		return nil
	})
}

// GetTableNameByCategory 根据类别获取表名
func GetTableNameByCategory(category string) string {
	return "cards_" + category
}

// CheckTableExists 检查表是否存在
func CheckTableExists(tableName string) bool {
	var count int64
	DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count)
	return count > 0
}

// GetUnsoldSubscriptionTypes 获取未售出的订阅类型列表
func GetUnsoldSubscriptionTypes(tableName string) ([]string, error) {
	var types []string
	err := DB.Table(tableName).
		Where("sell_status = ?", 1).
		Where("subscription_type != ?", "").
		Distinct("subscription_type").
		Pluck("subscription_type", &types).Error
	return types, err
}

// PickupCard 取货：按订阅过期时间先进先出选出一条，更新为售出中
func PickupCard(tableName string, subscriptionType string) (*AccountCard, error) {
	var card AccountCard

	// 开启事务
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 查询一条未售出的卡密，按订阅过期时间升序（先进先出）
		query := tx.Table(tableName).
			Where("sell_status = ?", 1).
			Where("subscription_type = ?", subscriptionType).
			Order("subscription_expired_time ASC, id ASC")

		if err := query.First(&card).Error; err != nil {
			return errors.New("没有可用的卡密")
		}

		// 更新为售出中
		if err := tx.Table(tableName).Where("id = ?", card.Id).Update("sell_status", 2).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &card, nil
}

// CompletePickup 完成取货：更新为已售出
func CompletePickup(tableName string, id int, sellPrice *float64, sellTo string) error {
	if id == 0 {
		return errors.New("id 为空")
	}

	// 获取当前时间戳
	now := time.Now().Unix()

	updates := map[string]interface{}{
		"sell_status": 3,
		"sell_date":   now,
	}

	if sellPrice != nil {
		updates["sell_price"] = sellPrice
	}

	if sellTo != "" {
		updates["sell_to"] = sellTo
	}

	return DB.Table(tableName).Where("id = ?", id).Updates(updates).Error
}
