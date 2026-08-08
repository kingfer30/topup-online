package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	CardStatusDeleted  = -1 // 软删
	CardStatusBanned   = -2 // 封禁（账号已删除/停用）
	CardStatusNormal   = 1  // 正常
	CardStatusDisabled = 2  // 禁用
)

// AccountCard 账号卡密模型
type AccountCard struct {
	Id                      int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Account                 string         `json:"account" gorm:"type:varchar(100);not null;uniqueIndex:idx_account"`
	Password                string         `json:"password" gorm:"type:varchar(50)"`
	MailPassword            string         `json:"mail_password" gorm:"type:varchar(50)"`
	SubscriptionStatus      int            `json:"subscription_status" gorm:"type:tinyint(2);default:1;comment:订阅状态 1已订阅 2未订阅 -1掉订阅 -2已订阅需点击pro"`
	SubscriptionType        string         `json:"subscription_type" gorm:"type:varchar(30);comment:订阅类型;index:idx_subscription_type"`
	SubscriptionTime        *int64         `json:"subscription_time" gorm:"type:bigint(20);comment:订阅时间;index:idx_subscription_time"`
	SubscriptionExpiredTime *int64         `json:"subscription_expired_time" gorm:"type:bigint(20);comment:订阅过期时间"`
	SubscriptionCredits     *float64       `json:"subscription_credits" gorm:"type:decimal(10,2);comment:订阅额度"`
	IsCheck                 int            `json:"is_check" gorm:"type:tinyint(2);default:-1;comment:检查状态 -1未检查 1检查成功 2检查失败"`
	CheckTime               *int64         `json:"check_time" gorm:"type:bigint(20);comment:检查时间"`
	PurchaseDate            *int64         `json:"purchase_date" gorm:"type:bigint(20);comment:购买时间"`
	PurchasePrice           *float64       `json:"purchase_price" gorm:"type:decimal(10,2);comment:购买价格(成本)"`
	PurchaseFrom            string         `json:"purchase_from" gorm:"type:varchar(50);comment:购买平台"`
	PurchaseBy              string         `json:"purchase_by" gorm:"type:varchar(100);comment:卖家名称"`
	SellPrice               *float64       `json:"sell_price" gorm:"type:decimal(10,2);comment:出售价格"`
	SellDate                *int64         `json:"sell_date" gorm:"type:bigint(20);comment:出售时间"`
	SellTo                  string         `json:"sell_to" gorm:"type:varchar(50);comment:出售对方"`
	SellOrderNo             string         `json:"sell_order_no" gorm:"type:varchar(100);comment:出售订单号"`
	SellStatus              int            `json:"sell_status" gorm:"type:tinyint(2);default:1;comment:出售状态 1 未出售 2发货中 3已出售;index:idx_subscription_sell"`
	AccountType             int            `json:"account_type" gorm:"type:tinyint(2);default:1;comment:账号类型 1普号 2成品"`
	Status                  int            `json:"status" gorm:"type:tinyint(2);default:1;comment:状态 -1软删 -2封禁 1正常 2禁用"`
	ApiKey                  string         `json:"api_key" gorm:"type:varchar(300);comment:apikey"`
	Token                   string         `json:"token" gorm:"type:text;comment:token"`
	TwoFA                   string         `json:"2fa" gorm:"column:2fa;type:varchar(100);comment:2fa"`
	ClientId                string         `json:"client_id" gorm:"type:varchar(200);comment:client_id"`
	MailUrl                 string         `json:"mail_url" gorm:"type:varchar(50);comment:邮箱地址"`
	Remark                  string         `json:"remark" gorm:"type:varchar(100);comment:备注"`
	CodeLink                string         `json:"code_link" gorm:"type:varchar(300);comment:接码链接"`
	FreezeStatus            int            `json:"freeze_status" gorm:"type:tinyint(2);default:-1;comment:冻结状态 -1未冻结 1已冻结"`
	FreezeTime              *int64         `json:"freeze_time" gorm:"type:bigint(20);comment:冻结时间"`
	FreezeRemark            string         `json:"freeze_remark" gorm:"type:varchar(200);comment:冻结备注"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ParseAccountSearchList 解析账号搜索列表（换行/逗号分隔，去重）
func ParseAccountSearchList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	seen := make(map[string]struct{})
	list := make([]string, 0)
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ','
	}) {
		s := strings.TrimSpace(part)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		list = append(list, s)
	}
	return list
}

func applyAccountSearchFilter(query *gorm.DB, accounts []string) *gorm.DB {
	if len(accounts) == 0 {
		return query
	}
	if len(accounts) == 1 {
		kw := accounts[0]
		return query.Where("account LIKE ? OR mail_url LIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	return query.Where("account IN ?", accounts)
}

// GetAllCardsForExport 获取符合条件的全部卡密（不分页，用于导出）
func GetAllCardsForExport(tableName string, cardType string, accounts []string, subscriptionType string, subscriptionStatus int, isCheck int, purchaseDate int64, freezeStatus int, freezeTime int64, sellTo string, purchaseBy string) ([]AccountCard, error) {
	var cards []AccountCard

	query := DB.Table(tableName).Where("status != ?", CardStatusDeleted)

	switch cardType {
	case "all":
		query = query.Where("account_type = ?", 1)
	case "unsold":
		query = query.Where("sell_status IN ?", []int{1, 2}).Where("subscription_status IN ?", []int{1, -2})
	case "sold":
		query = query.Where("sell_status = ?", 3)
	}

	if subscriptionType != "" && (cardType == "unsold" || cardType == "sold") {
		query = query.Where("subscription_type = ?", subscriptionType)
	}

	// 订阅状态筛选（仅已售列表使用）
	if subscriptionStatus != 0 && cardType == "sold" {
		query = query.Where("subscription_status = ?", subscriptionStatus)
	}

	// 检查状态筛选（未售/已售列表均支持，0 表示不过滤）
	if isCheck != 0 && (cardType == "sold" || cardType == "unsold") {
		query = query.Where("is_check = ?", isCheck)
	}

	// 日期按范围（0 表示不过滤）：已售列表按 sell_date；未售列表按 purchase_date
	if purchaseDate > 0 && cardType == "sold" {
		query = query.Where("sell_date >= ? AND sell_date < ?", purchaseDate, purchaseDate+86400)
	}
	if purchaseDate > 0 && cardType == "unsold" {
		query = query.Where("purchase_date >= ? AND purchase_date < ?", purchaseDate, purchaseDate+86400)
	}

	// 冻结状态筛选（仅普号列表，0 表示不过滤）
	if freezeStatus != 0 && cardType == "all" {
		query = query.Where("freeze_status = ?", freezeStatus)
	}

	// 冻结时间筛选（仅普号列表，按当天范围）
	if freezeTime > 0 && cardType == "all" {
		query = query.Where("freeze_time >= ? AND freeze_time < ?", freezeTime, freezeTime+86400)
	}

	// 出售对方（仅已售列表，模糊匹配）
	if st := strings.TrimSpace(sellTo); st != "" && cardType == "sold" {
		query = query.Where("sell_to LIKE ?", "%"+st+"%")
	}

	// 卖家名称（仅已售列表，模糊匹配）
	if pb := strings.TrimSpace(purchaseBy); pb != "" && cardType == "sold" {
		query = query.Where("purchase_by LIKE ?", "%"+pb+"%")
	}

	query = applyAccountSearchFilter(query, accounts)

	orderClause := "freeze_status ASC, updated_at DESC, id DESC"
	if cardType == "unsold" {
		orderClause = "sell_status DESC, purchase_date ASC, id ASC"
	} else if cardType == "sold" {
		orderClause = "sell_date DESC, id DESC"
	}

	err := query.Order(orderClause).Find(&cards).Error
	return cards, err
}

// GetCardList 获取卡密列表
func GetCardList(tableName string, cardType string, page, pageSize int, accounts []string, subscriptionType string, subscriptionStatus int, isCheck int, purchaseDate int64, freezeStatus int, freezeTime int64, sellTo string, purchaseBy string) ([]AccountCard, int64, error) {
	var cards []AccountCard
	var total int64

	// 构建查询
	query := DB.Table(tableName).Where("status != ?", CardStatusDeleted)

	// 根据类型筛选
	switch cardType {
	case "all":
		query = query.Where("account_type = ?", 1) // 普号列表：只显示普号
	case "unsold":
		query = query.Where("sell_status IN ?", []int{1, 2}).Where("subscription_status IN ?", []int{1, -2}) // 未售：已订阅或已订阅需点击pro
	case "sold":
		query = query.Where("sell_status = ?", 3) // 已售列表：已出售
	}

	// 订阅类型筛选（仅未售/已售列表）
	if subscriptionType != "" && (cardType == "unsold" || cardType == "sold") {
		query = query.Where("subscription_type = ?", subscriptionType)
	}

	// 订阅状态筛选（仅已售列表使用，0 表示不过滤）
	if subscriptionStatus != 0 && cardType == "sold" {
		query = query.Where("subscription_status = ?", subscriptionStatus)
	}

	// 检查状态筛选（未售/已售列表均支持，0 表示不过滤）
	if isCheck != 0 && (cardType == "sold" || cardType == "unsold") {
		query = query.Where("is_check = ?", isCheck)
	}

	// 日期按范围（0 表示不过滤）：已售列表按 sell_date；未售列表按 purchase_date
	if purchaseDate > 0 && cardType == "sold" {
		query = query.Where("sell_date >= ? AND sell_date < ?", purchaseDate, purchaseDate+86400)
	}
	if purchaseDate > 0 && cardType == "unsold" {
		query = query.Where("purchase_date >= ? AND purchase_date < ?", purchaseDate, purchaseDate+86400)
	}

	// 冻结状态筛选（仅普号列表，0 表示不过滤）
	if freezeStatus != 0 && cardType == "all" {
		query = query.Where("freeze_status = ?", freezeStatus)
	}

	// 冻结时间筛选（仅普号列表，按当天范围）
	if freezeTime > 0 && cardType == "all" {
		query = query.Where("freeze_time >= ? AND freeze_time < ?", freezeTime, freezeTime+86400)
	}

	// 出售对方（仅已售列表，模糊匹配）
	if st := strings.TrimSpace(sellTo); st != "" && cardType == "sold" {
		query = query.Where("sell_to LIKE ?", "%"+st+"%")
	}

	// 卖家名称（仅已售列表，模糊匹配）
	if pb := strings.TrimSpace(purchaseBy); pb != "" && cardType == "sold" {
		query = query.Where("purchase_by LIKE ?", "%"+pb+"%")
	}

	// 账号搜索（单行模糊，多行精确 IN）
	query = applyAccountSearchFilter(query, accounts)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询：未售列表发货中优先，同组内按订阅到期时间先进先出（与取货逻辑一致）
	// 普号列表：未冻结（freeze_status=-1）排前，已冻结（freeze_status=1）排后，同组内按 updated_at DESC
	offset := (page - 1) * pageSize
	orderClause := "freeze_status ASC, updated_at DESC, id DESC"
	if cardType == "unsold" {
		orderClause = "sell_status DESC, subscription_expired_time ASC, id ASC"
	} else if cardType == "sold" {
		orderClause = "sell_date DESC, id DESC"
	}
	if err := query.Order(orderClause).Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

// GetCardsByIds 根据 ID 列表批量获取卡密（仅返回 token 不为空的）
func GetCardsByIds(tableName string, ids []int) ([]*AccountCard, error) {
	if len(ids) == 0 {
		return nil, errors.New("id 列表为空")
	}
	var cards []*AccountCard
	err := DB.Table(tableName).
		Where("id IN ? AND status != ? AND token IS NOT NULL AND token != ''", ids, CardStatusDeleted).
		Find(&cards).Error
	return cards, err
}

// GetCardById 根据ID获取卡密
func GetCardById(tableName string, id int) (*AccountCard, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	var card AccountCard
	err := DB.Table(tableName).Where("status != ?", CardStatusDeleted).First(&card, "id = ?", id).Error
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
	// 只更新编辑表单中实际可编辑的字段，避免把 purchase_date/subscription_time
	// 等表单未涉及的时间字段误写为 NULL（GORM map 更新会写入 nil 值）
	updates := map[string]interface{}{
		"account":             card.Account,
		"password":            card.Password,
		"mail_password":       card.MailPassword,
		"subscription_status": card.SubscriptionStatus,
		"subscription_type":   card.SubscriptionType,
		"purchase_price":      card.PurchasePrice,
		"purchase_from":       card.PurchaseFrom,
		"purchase_by":         card.PurchaseBy,
		"sell_price":          card.SellPrice,
		"sell_to":             card.SellTo,
		"sell_status":         card.SellStatus,
		"account_type":        card.AccountType,
		"status":              card.Status,
		"api_key":             card.ApiKey,
		"token":               card.Token,
		"2fa":                 card.TwoFA,
		"client_id":           card.ClientId,
		"mail_url":            card.MailUrl,
		"remark":              card.Remark,
		"code_link":           card.CodeLink,
	}
	return DB.Table(tableName).Where("id = ?", id).Updates(updates).Error
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

	// 使用事务批量导入：账号存在则更新，不存在则创建
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, card := range cards {
			var existed AccountCard
			err := tx.Table(tableName).Where("account = ?", card.Account).First(&existed).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("查询卡密失败: %s, %v", card.Account, err)
				}
				if err := tx.Table(tableName).Create(&card).Error; err != nil {
					return fmt.Errorf("创建卡密失败: %s, %v", card.Account, err)
				}
				continue
			}

			// 使用 struct 更新，仅覆盖非零值字段，避免空字符串/0误清空已有数据
			if err := tx.Table(tableName).Where("id = ?", existed.Id).Updates(&card).Error; err != nil {
				return fmt.Errorf("更新卡密失败: %s, %v", card.Account, err)
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

// GetUnsoldSubscriptionTypes 获取未售出的订阅类型列表（仅成品 account_type=2）
func GetUnsoldSubscriptionTypes(tableName string) ([]string, error) {
	var types []string
	err := DB.Table(tableName).
		Where("sell_status = ?", 1).
		Where("account_type = ?", 2).
		Where("status NOT IN ?", []int{CardStatusDeleted, CardStatusBanned}).
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
		// 查询一条未售出的成品卡密（account_type=2），按订阅过期时间升序（先进先出）
		query := tx.Table(tableName).
			Where("sell_status = ?", 1).
			Where("account_type = ?", 2).
			Where("status NOT IN ?", []int{CardStatusDeleted, CardStatusBanned}).
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

// BatchPickup 批量取货：将指定 ID 的未售卡密直接标记为已出售(3)，并记录售出信息
func BatchPickup(tableName string, ids []int, sellPrice *float64, sellTo string) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("未选择任何记录")
	}

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

	result := DB.Table(tableName).
		Where("id IN ? AND status NOT IN ? AND sell_status IN ?", ids, []int{CardStatusDeleted, CardStatusBanned}, []int{1, 2}).
		Updates(updates)
	return result.RowsAffected, result.Error
}

// RollbackSoldCard 回滚已售：将已出售(3)重置为未出售(1)，并清空售出相关字段
func RollbackSoldCard(tableName string, id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}

	result := DB.Table(tableName).Where("id = ? AND status != ? AND sell_status = ?", id, CardStatusDeleted, 3).Updates(map[string]interface{}{
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

// BatchDashboardGotoAfterResolve 提链返回 dashboard 时：仅当订阅状态为掉订阅(-1)时才回滚已售至未售，并标记订阅状态为已订阅需点击pro（-2）、检查成功
func BatchDashboardGotoAfterResolve(tableName string, ids []int) (success int, errMsgs []string) {
	for _, id := range ids {
		if id == 0 {
			continue
		}
		card, err := GetCardById(tableName, id)
		if err != nil || card == nil {
			errMsgs = append(errMsgs, fmt.Sprintf("ID %d: 记录不存在", id))
			continue
		}
		if card.SubscriptionStatus != -1 {
			errMsgs = append(errMsgs, fmt.Sprintf("ID %d (%s): 订阅状态非掉订阅，不可回滚", id, card.Account))
			continue
		}
		if err := RollbackSoldCard(tableName, id); err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("ID %d 回滚: %v", id, err))
			continue
		}
		if err := DB.Table(tableName).Where("id = ?", id).Updates(map[string]interface{}{
			"subscription_status": -2,
			"is_check":            1,
		}).Error; err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("ID %d 更新状态: %v", id, err))
			continue
		}
		success++
	}
	return success, errMsgs
}

// RollbackPickup 回滚取货：将发货中(2)重置为未出售(1)
func RollbackPickup(tableName string, id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}

	result := DB.Table(tableName).Where("id = ? AND status != ? AND sell_status = ?", id, CardStatusDeleted, 2).Update("sell_status", 1)
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
	Category           string                 `json:"category"`
	SoldCount          int64                  `json:"sold_count"`
	ProductStockCount  int64                  `json:"product_stock_count"`  // 成品未售库存合计（account_type=2）
	ProductStockByType []SubscriptionTypeStat `json:"product_stock_by_type"` // 成品按订阅类型细分
	RegularStockCount  int64                  `json:"regular_stock_count"`  // 普号库存合计（account_type=1，未冻结）
	RegularStockByType []SubscriptionTypeStat `json:"regular_stock_by_type"` // 普号按订阅类型细分
	RevenueUSD         float64                `json:"revenue_usd"`
	RevenueCNY         float64                `json:"revenue_cny"`
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
			"SELECT COUNT(*) AS sold_count, COALESCE(SUM(sell_price), 0) AS revenue_usd FROM `%s` WHERE status != -1 AND sell_status = 3 AND sell_date >= ? AND sell_date < ? AND deleted_at IS NULL",
			tableName,
		), startOfDay, endOfDay).Scan(&soldResult).Error
		if err != nil {
			continue
		}

		// 按订阅类型统计成品未售库存（account_type=2, sell_status=1）
		type subTypeRow struct {
			SubscriptionType string `gorm:"column:subscription_type"`
			Count            int64  `gorm:"column:cnt"`
		}
		var productRows []subTypeRow
		DB.Raw(fmt.Sprintf(
			"SELECT subscription_type, COUNT(*) AS cnt FROM `%s` WHERE status != -1 AND account_type = 2 AND sell_status = 1 AND deleted_at IS NULL GROUP BY subscription_type ORDER BY cnt DESC",
			tableName,
		)).Scan(&productRows)

		productStockByType := make([]SubscriptionTypeStat, 0, len(productRows))
		var productStock int64
		for _, row := range productRows {
			productStockByType = append(productStockByType, SubscriptionTypeStat{
				SubscriptionType: row.SubscriptionType,
				Count:            row.Count,
			})
			productStock += row.Count
		}

		// 按订阅类型统计普号库存（account_type=1，排除已冻结）
		var regularRows []subTypeRow
		DB.Raw(fmt.Sprintf(
			"SELECT subscription_type, COUNT(*) AS cnt FROM `%s` WHERE status != -1 AND account_type = 1 AND (freeze_status IS NULL OR freeze_status != 1) AND deleted_at IS NULL GROUP BY subscription_type ORDER BY cnt DESC",
			tableName,
		)).Scan(&regularRows)

		regularStockByType := make([]SubscriptionTypeStat, 0, len(regularRows))
		var regularStock int64
		for _, row := range regularRows {
			regularStockByType = append(regularStockByType, SubscriptionTypeStat{
				SubscriptionType: row.SubscriptionType,
				Count:            row.Count,
			})
			regularStock += row.Count
		}

		stats = append(stats, CardTypeStat{
			Category:           category,
			SoldCount:          soldResult.SoldCount,
			ProductStockCount:  productStock,
			ProductStockByType: productStockByType,
			RegularStockCount:  regularStock,
			RegularStockByType: regularStockByType,
			RevenueUSD:         soldResult.RevenueUSD,
			RevenueCNY:         soldResult.RevenueUSD * 7,
		})
	}

	return stats, nil
}

// BatchUpgradeRequest 批量升级为成品请求
type BatchUpgradeRequest struct {
	IDs                       []int
	SubscriptionType          string
	SubscriptionTime          *int64
	SubscriptionRemainingDays *int // 订阅剩余天数，用于计算 subscription_expired_time（同批量导入）
	PurchasePrice             *float64 // 追加到现有购买价格
	PurchaseFrom              string
	PurchaseDate              *int64
}

// BatchUpgradeToProduct 批量将普号升级为成品
// 购买价格为累加逻辑：新值 = COALESCE(原值, 0) + 追加值
func BatchUpgradeToProduct(tableName string, req BatchUpgradeRequest) (int64, error) {
	if len(req.IDs) == 0 {
		return 0, errors.New("未选择任何记录")
	}

	// 过滤掉已冻结的卡密
	var validIDs []int
	DB.Table(tableName).
		Where("id IN ? AND (freeze_status IS NULL OR freeze_status != 1)", req.IDs).
		Pluck("id", &validIDs)
	if len(validIDs) == 0 {
		return 0, errors.New("所选卡密均已冻结，无法更新为成品")
	}

	// 购买价格累加（单独执行，使用 SQL 表达式）
	if req.PurchasePrice != nil {
		err := DB.Table(tableName).Where("id IN ?", validIDs).
			UpdateColumn("purchase_price", gorm.Expr("COALESCE(purchase_price, 0) + ?", *req.PurchasePrice)).Error
		if err != nil {
			return 0, fmt.Errorf("更新购买价格失败: %v", err)
		}
	}

	// 其余字段批量更新
	updates := map[string]interface{}{
		"subscription_status": 1,
		"account_type":        2,
	}
	subscriptionType := strings.TrimSpace(req.SubscriptionType)
	if subscriptionType != "" {
		updates["subscription_type"] = subscriptionType
	}
	if req.SubscriptionTime != nil {
		updates["subscription_time"] = req.SubscriptionTime
	}
	if req.PurchaseFrom != "" {
		updates["purchase_from"] = req.PurchaseFrom
	}
	if req.PurchaseDate != nil {
		updates["purchase_date"] = req.PurchaseDate
	}
	// 订阅过期时间：当前时间 + 剩余天数（逻辑同批量导入，默认 30 天）
	remainingDays := 30
	if req.SubscriptionRemainingDays != nil {
		remainingDays = *req.SubscriptionRemainingDays
	}
	if remainingDays > 0 {
		expiredAt := time.Now().Unix() + int64(remainingDays)*24*60*60
		updates["subscription_expired_time"] = expiredAt
	}

	result := DB.Table(tableName).Where("id IN ?", validIDs).Updates(updates)
	return result.RowsAffected, result.Error
}

// BatchFreezeCards 批量冻结/解冻普号卡密
// freeze=1 冻结，freeze=-1 解冻
func BatchFreezeCards(tableName string, ids []int, freeze int, remark string) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("未选择任何记录")
	}
	updates := map[string]interface{}{
		"freeze_status": freeze,
		"freeze_remark": remark,
	}
	if freeze == 1 {
		now := time.Now().Unix()
		updates["freeze_time"] = now
	} else {
		updates["freeze_time"] = nil
	}
	result := DB.Table(tableName).Where("id IN ?", ids).Updates(updates)
	return result.RowsAffected, result.Error
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

	return DB.Table(tableName).Where("id = ? AND status != ?", id, CardStatusDeleted).Updates(updates).Error
}

// BatchDeleteByStatus 批量软删：将 status 置为 -1
func BatchDeleteByStatus(tableName string, ids []int) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("未选择任何记录")
	}
	result := DB.Table(tableName).Where("id IN ?", ids).Update("status", CardStatusDeleted)
	return result.RowsAffected, result.Error
}

// GetAllCardTableNames 获取所有 cards_* 表名
func GetAllCardTableNames() ([]string, error) {
	var tableNames []string
	err := DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name LIKE 'cards_%'").
		Pluck("table_name", &tableNames).Error
	return tableNames, err
}

// GetUnsoldCardsWithToken 获取指定表中 sell_status=1 且 token 不为空，且今天未检查过的成品卡密（不含 subscription_status=-2「已订阅需点击pro」，不参与每日掉订阅巡检）
func GetUnsoldCardsWithToken(tableName string, limit int) ([]*AccountCard, error) {
	var cards []*AccountCard
	// 今天零点的 Unix 时间戳
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	err := DB.Table(tableName).
		Where("sell_status = ? AND account_type=2 AND token IS NOT NULL AND token != ''", 1).
		Where("subscription_status != ?", -2).
		Where("check_time IS NULL OR check_time < ?", todayStart).
		Limit(limit).
		Find(&cards).Error
	if err != nil {
		return nil, err
	}
	return cards, nil
}

// UpdateCardCheckResult 更新卡密检查结果
func UpdateCardCheckResult(tableName string, id int, updates map[string]interface{}) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	return DB.Table(tableName).Where("id = ?", id).Updates(updates).Error
}

// MigrateCardTableColumns 对所有 cards_* 表执行新增列迁移（幂等，列已存在则跳过）
func MigrateCardTableColumns() error {
	tableNames, err := GetAllCardTableNames()
	if err != nil {
		return fmt.Errorf("获取 cards_* 表名失败: %w", err)
	}

	newColumns := []struct {
		name       string
		definition string
	}{
		{"subscription_credits", "decimal(10,2) NULL COMMENT '订阅额度'"},
		{"is_check", "tinyint(2) NOT NULL DEFAULT -1 COMMENT '检查状态 -1未检查 1检查成功 2检查失败'"},
		{"check_time", "bigint(20) NULL COMMENT '检查时间'"},
		{"code_link", "varchar(300) NULL COMMENT '接码链接'"},
		{"freeze_status", "tinyint(2) NOT NULL DEFAULT -1 COMMENT '冻结状态 -1未冻结 1已冻结'"},
		{"freeze_time", "bigint(20) NULL COMMENT '冻结时间'"},
		{"freeze_remark", "varchar(200) NULL COMMENT '冻结备注'"},
		{"client_id", "varchar(200) NULL COMMENT 'client_id'"},
	}

	for _, tableName := range tableNames {
		for _, col := range newColumns {
			// 查询列是否已存在
			var count int64
			DB.Raw(
				"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
				tableName, col.name,
			).Scan(&count)

			if count == 0 {
				sql := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", tableName, col.name, col.definition)
				if err := DB.Exec(sql).Error; err != nil {
					return fmt.Errorf("迁移表 %s 列 %s 失败: %w", tableName, col.name, err)
				}
			}
		}
	}
	return nil
}
