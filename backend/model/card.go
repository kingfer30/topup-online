package model

import (
	"errors"
	"fmt"
	"strings"
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
func GetCardList(tableName string, cardType string, page, pageSize int, keyword string, subscriptionType string) ([]AccountCard, int64, error) {
	var cards []AccountCard
	var total int64

	// 构建查询
	query := DB.Table(tableName)

	// 根据类型筛选
	switch cardType {
	case "all":
		query = query.Where("account_type = ?", 1) // 普号列表：只显示普号
	case "unsold":
		query = query.Where("sell_status IN ?", []int{1, 2}) // 未售列表：未出售 + 发货中
	case "sold":
		query = query.Where("sell_status = ?", 3) // 已售列表：已出售
	}

	// 订阅类型筛选（仅未售/已售列表）
	if subscriptionType != "" && (cardType == "unsold" || cardType == "sold") {
		query = query.Where("subscription_type = ?", subscriptionType)
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("account LIKE ? OR mail_url LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询：未售列表按 sell_status DESC（发货中=2 排前面），已售列表按 sell_date DESC，其余按 id DESC
	offset := (page - 1) * pageSize
	orderClause := "id DESC"
	if cardType == "unsold" {
		orderClause = "sell_status DESC, purchase_date ASC, id ASC"
	} else if cardType == "sold" {
		orderClause = "sell_date DESC, id DESC"
	}
	if err := query.Order(orderClause).Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
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
// format 为 "reverse" 时额外要求 token 不为空
func PickupCard(tableName string, subscriptionType string, format string) (*AccountCard, error) {
	var card AccountCard

	// 开启事务
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 查询一条未售出的卡密，按订阅过期时间升序（先进先出）
		query := tx.Table(tableName).
			Where("sell_status = ?", 1).
			Where("subscription_type = ?", subscriptionType).
			Order("subscription_expired_time ASC, id ASC")

		// 逆向格式要求 token 不为空
		if format == "reverse" {
			query = query.Where("token != ?", "")
		}

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

// RollbackSoldCard 回滚已售：将已出售(3)重置为未出售(1)，并清空售出相关字段
func RollbackSoldCard(tableName string, id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}

	result := DB.Table(tableName).Where("id = ? AND sell_status = ?", id, 3).Updates(map[string]interface{}{
		"sell_status":   1,
		"sell_price":    nil,
		"sell_date":     nil,
		"sell_to":       "",
		"sell_order_no": "",
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("卡密不存在或状态不是已出售")
	}
	return nil
}

// RollbackPickup 回滚取货：将发货中(2)重置为未出售(1)
func RollbackPickup(tableName string, id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}

	result := DB.Table(tableName).Where("id = ? AND sell_status = ?", id, 2).Update("sell_status", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("卡密不存在或状态不是发货中")
	}
	return nil
}

// SubscriptionTypeStat 按订阅类型的库存统计
type SubscriptionTypeStat struct {
	SubscriptionType string `json:"subscription_type"`
	Count            int64  `json:"count"`
}

// CardTypeStat 卡密类型统计
type CardTypeStat struct {
	Category    string                 `json:"category"`
	SoldCount   int64                  `json:"sold_count"`
	StockCount  int64                  `json:"stock_count"`   // 剩余未售库存合计
	StockByType []SubscriptionTypeStat `json:"stock_by_type"` // 按订阅类型细分库存
	RevenueUSD  float64                `json:"revenue_usd"`
	RevenueCNY  float64                `json:"revenue_cny"`
}

// GetDashboardStats 获取控制台按卡密类型统计销售数据
// dateStr 格式为 "2006-01-02"，空字符串表示今天
func GetDashboardStats(dateStr string) ([]CardTypeStat, error) {
	// 获取所有 cards_ 开头的表
	var tableNames []string
	err := DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name LIKE 'cards_%'").
		Pluck("table_name", &tableNames).Error
	if err != nil {
		return nil, err
	}

	// 固定使用 Asia/Shanghai (UTC+8) 避免服务器时区差异
	cst, _ := time.LoadLocation("Asia/Shanghai")
	var startOfDay int64
	if dateStr == "" {
		// 未指定日期则使用今天
		now := time.Now().In(cst)
		startOfDay = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, cst).Unix()
	} else {
		// 解析传入日期
		t, err := time.ParseInLocation("2006-01-02", dateStr, cst)
		if err != nil {
			return nil, fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD: %v", err)
		}
		startOfDay = t.Unix()
	}
	endOfDay := startOfDay + 86400

	stats := make([]CardTypeStat, 0, len(tableNames))
	for _, tableName := range tableNames {
		// 提取类别名（去掉 "cards_" 前缀）
		category := strings.TrimPrefix(tableName, "cards_")

		// 统计今日售出数量和总收入
		var soldResult struct {
			SoldCount  int64   `gorm:"column:sold_count"`
			RevenueUSD float64 `gorm:"column:revenue_usd"`
		}
		err := DB.Raw(fmt.Sprintf(
			"SELECT COUNT(*) AS sold_count, COALESCE(SUM(sell_price), 0) AS revenue_usd FROM `%s` WHERE sell_status = 3 AND sell_date >= ? AND sell_date < ? AND deleted_at IS NULL",
			tableName,
		), startOfDay, endOfDay).Scan(&soldResult).Error
		if err != nil {
			continue
		}

		// 按订阅类型统计剩余未售库存（sell_status = 1 未出售）
		type subTypeRow struct {
			SubscriptionType string `gorm:"column:subscription_type"`
			Count            int64  `gorm:"column:cnt"`
		}
		var subTypeRows []subTypeRow
		DB.Raw(fmt.Sprintf(
			"SELECT subscription_type, COUNT(*) AS cnt FROM `%s` WHERE sell_status = 1 AND deleted_at IS NULL GROUP BY subscription_type ORDER BY cnt DESC",
			tableName,
		)).Scan(&subTypeRows)

		stockByType := make([]SubscriptionTypeStat, 0, len(subTypeRows))
		var totalStock int64
		for _, row := range subTypeRows {
			stockByType = append(stockByType, SubscriptionTypeStat{
				SubscriptionType: row.SubscriptionType,
				Count:            row.Count,
			})
			totalStock += row.Count
		}

		stats = append(stats, CardTypeStat{
			Category:    category,
			SoldCount:   soldResult.SoldCount,
			StockCount:  totalStock,
			StockByType: stockByType,
			RevenueUSD:  soldResult.RevenueUSD,
			RevenueCNY:  soldResult.RevenueUSD * 7,
		})
	}

	return stats, nil
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
