package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ApiBillPending = 1
	ApiBillPaid    = 2
	ApiBillVoid    = 3
)

// ApiPlatform 上游 new-api 台子
type ApiPlatform struct {
	Id              int            `json:"id" gorm:"primaryKey;autoIncrement"`
	BaseURL         string         `json:"base_url" gorm:"type:varchar(255);not null;comment:台子地址"`
	APIKey          string         `json:"api_key" gorm:"type:varchar(512);not null;comment:API Key"`
	Account         string         `json:"account" gorm:"type:varchar(100);comment:账号"`
	Password        string         `json:"password" gorm:"type:varchar(200);comment:密码"`
	Remark          string         `json:"remark" gorm:"type:varchar(255);comment:备注"`
	UsedAmount      float64        `json:"used_amount" gorm:"type:decimal(18,6);default:0;comment:消耗总额美元"`
	SettledAmount   float64        `json:"settled_amount" gorm:"type:decimal(18,6);default:0;comment:已结算美元"`
	UnsettledAmount float64        `json:"unsettled_amount" gorm:"type:decimal(18,6);default:0;comment:未结算美元"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ApiPlatform) TableName() string { return "api_platform" }

// ApiPlatformChannel 台子下已保存的渠道
type ApiPlatformChannel struct {
	Id              int        `json:"id" gorm:"primaryKey;autoIncrement"`
	PlatformId      int        `json:"platform_id" gorm:"not null;uniqueIndex:idx_platform_upstream,priority:1;index"`
	ChannelId       int        `json:"channel_id" gorm:"not null;uniqueIndex:idx_platform_upstream,priority:2;comment:上游渠道ID"`
	Name            string     `json:"name" gorm:"type:varchar(255);comment:渠道名称"`
	FilterName      string     `json:"filter_name" gorm:"type:varchar(100);index;comment:筛选名称"`
	Models          string     `json:"models" gorm:"type:text;comment:可用模型"`
	Status          int        `json:"status" gorm:"type:tinyint;default:1;comment:1启用 2手动禁用 3自动禁用"`
	UsedQuota       int64      `json:"used_quota" gorm:"type:bigint;default:0;comment:原始已用额度"`
	UsedAmount      float64    `json:"used_amount" gorm:"type:decimal(18,6);default:0;comment:消耗总额美元"`
	SettledAmount   float64    `json:"settled_amount" gorm:"type:decimal(18,6);default:0;comment:已结算美元"`
	UnsettledAmount float64    `json:"unsettled_amount" gorm:"type:decimal(18,6);default:0;comment:未结算美元"`
	CostPrice       *float64   `json:"cost_price" gorm:"type:decimal(18,4);comment:成本价人民币"`
	SalePrice       *float64   `json:"sale_price" gorm:"type:decimal(18,4);comment:销售价人民币"`
	LastSyncAt      *time.Time `json:"last_sync_at"`
	LastTestResult  string     `json:"last_test_result" gorm:"type:varchar(2000);comment:最近测试结果"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (ApiPlatformChannel) TableName() string { return "api_platform_channels" }

// ApiPlatformBill 台子账单
type ApiPlatformBill struct {
	Id           int                   `json:"id" gorm:"primaryKey;autoIncrement"`
	PlatformId   int                   `json:"platform_id" gorm:"not null;index"`
	BillNo       string                `json:"bill_no" gorm:"type:varchar(32);not null;uniqueIndex;comment:账单编号"`
	Status       int                   `json:"status" gorm:"type:tinyint;default:1;index;comment:1待支付 2已支付 3已作废"`
	PayQuota     float64               `json:"pay_quota" gorm:"type:decimal(18,6);default:0;comment:待支付额度美元"`
	AmountCny    float64               `json:"amount_cny" gorm:"type:decimal(18,4);default:0;comment:金额人民币"`
	PaidAt       *time.Time            `json:"paid_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Items        []ApiPlatformBillItem `json:"items,omitempty" gorm:"foreignKey:BillId"`
	ChannelCount int                   `json:"channel_count" gorm:"-"`
}

func (ApiPlatformBill) TableName() string { return "api_platform_bills" }

// ApiPlatformBillItem 账单冻结明细
type ApiPlatformBillItem struct {
	Id            int     `json:"id" gorm:"primaryKey;autoIncrement"`
	BillId        int     `json:"bill_id" gorm:"not null;index"`
	ChannelRowId  int     `json:"channel_row_id" gorm:"not null;index;comment:本地渠道行ID"`
	ChannelId     int     `json:"channel_id" gorm:"not null;comment:上游渠道ID"`
	Name          string  `json:"name" gorm:"type:varchar(255)"`
	FilterName    string  `json:"filter_name" gorm:"type:varchar(100)"`
	UsedAmount    float64 `json:"used_amount" gorm:"type:decimal(18,6);default:0"`
	SettledAmount float64 `json:"settled_amount" gorm:"type:decimal(18,6);default:0"`
	PayQuota      float64 `json:"pay_quota" gorm:"type:decimal(18,6);default:0;comment:冻结的未结算美元"`
	CostPrice     float64 `json:"cost_price" gorm:"type:decimal(18,4);default:0"`
	SalePrice     float64 `json:"sale_price" gorm:"type:decimal(18,4);default:0"`
	AmountCny     float64 `json:"amount_cny" gorm:"type:decimal(18,4);default:0"`
}

func (ApiPlatformBillItem) TableName() string { return "api_platform_bill_items" }

func ListApiPlatforms() ([]ApiPlatform, error) {
	var rows []ApiPlatform
	err := DB.Order("id desc").Find(&rows).Error
	return rows, err
}

func GetApiPlatform(id int) (*ApiPlatform, error) {
	var row ApiPlatform
	if err := DB.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func CreateApiPlatform(row *ApiPlatform) error {
	return DB.Create(row).Error
}

func UpdateApiPlatform(row *ApiPlatform) error {
	return DB.Model(row).Select("base_url", "api_key", "account", "password", "remark").Updates(row).Error
}

func DeleteApiPlatform(id int) error {
	return DB.Delete(&ApiPlatform{}, id).Error
}

func ListApiPlatformChannels(platformID int) ([]ApiPlatformChannel, error) {
	var rows []ApiPlatformChannel
	err := DB.Where("platform_id = ?", platformID).Order("channel_id asc").Find(&rows).Error
	return rows, err
}

func GetApiPlatformChannel(platformID, rowID int) (*ApiPlatformChannel, error) {
	var row ApiPlatformChannel
	err := DB.Where("platform_id = ? AND id = ?", platformID, rowID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func PendingUpstreamChannelIDs(platformID int) (map[int]struct{}, error) {
	var ids []int
	err := DB.Table("api_platform_bill_items AS i").
		Joins("JOIN api_platform_bills AS b ON b.id = i.bill_id").
		Where("b.platform_id = ? AND b.status = ?", platformID, ApiBillPending).
		Pluck("i.channel_id", &ids).Error
	if err != nil {
		return nil, err
	}
	set := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set, nil
}

func ExistingUpstreamIDs(platformID int) (map[int]ApiPlatformChannel, error) {
	rows, err := ListApiPlatformChannels(platformID)
	if err != nil {
		return nil, err
	}
	out := make(map[int]ApiPlatformChannel, len(rows))
	for _, row := range rows {
		out[row.ChannelId] = row
	}
	return out, nil
}

// SavePulledChannel 写入拉取结果。已存在的行保留成本价、销售价和已结算。
func SavePulledChannel(platformID int, incoming ApiPlatformChannel) error {
	var existing ApiPlatformChannel
	err := DB.Where("platform_id = ? AND channel_id = ?", platformID, incoming.ChannelId).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		incoming.PlatformId = platformID
		incoming.SettledAmount = 0
		incoming.UnsettledAmount = incoming.UsedAmount
		if incoming.UnsettledAmount < 0 {
			incoming.UnsettledAmount = 0
		}
		now := time.Now()
		incoming.LastSyncAt = &now
		if err := DB.Create(&incoming).Error; err != nil {
			return err
		}
		return RecalcApiPlatform(platformID)
	}
	if err != nil {
		return err
	}
	unsettled := incoming.UsedAmount - existing.SettledAmount
	if unsettled < 0 {
		unsettled = 0
	}
	now := time.Now()
	updates := map[string]any{
		"name":             incoming.Name,
		"models":           incoming.Models,
		"status":           incoming.Status,
		"used_quota":       incoming.UsedQuota,
		"used_amount":      incoming.UsedAmount,
		"unsettled_amount": unsettled,
		"last_sync_at":     now,
	}
	if incoming.CostPrice != nil {
		updates["cost_price"] = *incoming.CostPrice
	}
	if incoming.SalePrice != nil {
		updates["sale_price"] = *incoming.SalePrice
	}
	if strings.TrimSpace(incoming.FilterName) != "" {
		updates["filter_name"] = strings.TrimSpace(incoming.FilterName)
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ApiPlatformChannel{}).Where("id = ?", existing.Id).Updates(updates).Error; err != nil {
			return err
		}
		return recalcApiPlatformTx(tx, platformID)
	})
}

func UpdateApiPlatformChannelMoney(platformID, rowID int, cost, sale, settled *float64) error {
	row, err := GetApiPlatformChannel(platformID, rowID)
	if err != nil {
		return err
	}
	updates := map[string]any{}
	if cost != nil {
		updates["cost_price"] = *cost
	}
	if sale != nil {
		updates["sale_price"] = *sale
	}
	if settled != nil {
		settledVal := *settled
		if settledVal < 0 {
			settledVal = 0
		}
		unsettled := row.UsedAmount - settledVal
		if unsettled < 0 {
			unsettled = 0
		}
		updates["settled_amount"] = settledVal
		updates["unsettled_amount"] = unsettled
	}
	if len(updates) == 0 {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ApiPlatformChannel{}).Where("id = ?", row.Id).Updates(updates).Error; err != nil {
			return err
		}
		return recalcApiPlatformTx(tx, platformID)
	})
}

// BatchUpdateApiPlatformChannelPrices 把同一组成本价、销售价写到勾选的渠道。
func BatchUpdateApiPlatformChannelPrices(platformID int, rowIDs []int, cost, sale *float64) (int, error) {
	ids := uniqueInts(rowIDs)
	if len(ids) == 0 {
		return 0, errors.New("请勾选渠道")
	}
	if cost == nil && sale == nil {
		return 0, errors.New("请填写成本价或销售价")
	}
	var count int64
	if err := DB.Model(&ApiPlatformChannel{}).Where("platform_id = ? AND id IN ?", platformID, ids).Count(&count).Error; err != nil {
		return 0, err
	}
	if int(count) != len(ids) {
		return 0, errors.New("包含不属于该台子的渠道")
	}
	updates := map[string]any{}
	if cost != nil {
		updates["cost_price"] = *cost
	}
	if sale != nil {
		updates["sale_price"] = *sale
	}
	err := DB.Model(&ApiPlatformChannel{}).Where("platform_id = ? AND id IN ?", platformID, ids).Updates(updates).Error
	return len(ids), err
}

func UpdateApiPlatformChannelStatus(platformID, rowID, status int) error {
	return DB.Model(&ApiPlatformChannel{}).
		Where("platform_id = ? AND id = ?", platformID, rowID).
		Update("status", status).Error
}

func UpdateApiPlatformChannelModels(platformID, rowID int, models string) error {
	return DB.Model(&ApiPlatformChannel{}).
		Where("platform_id = ? AND id = ?", platformID, rowID).
		Update("models", models).Error
}

func UpdateApiPlatformChannelTest(platformID, rowID int, result string) error {
	if len(result) > 2000 {
		result = result[:2000]
	}
	return DB.Model(&ApiPlatformChannel{}).
		Where("platform_id = ? AND id = ?", platformID, rowID).
		Update("last_test_result", result).Error
}

func ApplyApiPlatformChannelUsage(platformID, rowID int, name string, status int, usedQuota int64, usedAmount float64) error {
	row, err := GetApiPlatformChannel(platformID, rowID)
	if err != nil {
		return err
	}
	unsettled := usedAmount - row.SettledAmount
	if unsettled < 0 {
		unsettled = 0
	}
	now := time.Now()
	return DB.Model(&ApiPlatformChannel{}).Where("id = ?", row.Id).Updates(map[string]any{
		"name":             name,
		"status":           status,
		"used_quota":       usedQuota,
		"used_amount":      usedAmount,
		"unsettled_amount": unsettled,
		"last_sync_at":     now,
	}).Error
}

// ChannelUsageUpdate 一次同步里某个渠道的最新用量。
type ChannelUsageUpdate struct {
	RowID         int
	Name          string
	Status        int
	UsedQuota     int64
	UsedAmount    float64
	SettledAmount float64
}

// ApplyApiPlatformChannelUsages 批量回写消耗，已结算保持不变。
func ApplyApiPlatformChannelUsages(platformID int, items []ChannelUsageUpdate) error {
	if len(items) == 0 {
		return RecalcApiPlatform(platformID)
	}
	now := time.Now()
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			unsettled := item.UsedAmount - item.SettledAmount
			if unsettled < 0 {
				unsettled = 0
			}
			if err := tx.Model(&ApiPlatformChannel{}).
				Where("platform_id = ? AND id = ?", platformID, item.RowID).
				Updates(map[string]any{
					"name":             item.Name,
					"status":           item.Status,
					"used_quota":       item.UsedQuota,
					"used_amount":      item.UsedAmount,
					"unsettled_amount": unsettled,
					"last_sync_at":     now,
				}).Error; err != nil {
				return err
			}
		}
		return recalcApiPlatformTx(tx, platformID)
	})
}

func DistinctFilterNames(platformID int) ([]string, error) {
	var names []string
	err := DB.Model(&ApiPlatformChannel{}).
		Where("platform_id = ? AND filter_name <> ''", platformID).
		Distinct("filter_name").
		Pluck("filter_name", &names).Error
	return names, err
}

func RecalcApiPlatform(platformID int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return recalcApiPlatformTx(tx, platformID)
	})
}

func recalcApiPlatformTx(tx *gorm.DB, platformID int) error {
	var sums struct {
		Used      float64
		Settled   float64
		Unsettled float64
	}
	if err := tx.Model(&ApiPlatformChannel{}).
		Where("platform_id = ?", platformID).
		Select("COALESCE(SUM(used_amount),0) AS used, COALESCE(SUM(settled_amount),0) AS settled, COALESCE(SUM(unsettled_amount),0) AS unsettled").
		Scan(&sums).Error; err != nil {
		return err
	}
	return tx.Model(&ApiPlatform{}).Where("id = ?", platformID).Updates(map[string]any{
		"used_amount":      sums.Used,
		"settled_amount":   sums.Settled,
		"unsettled_amount": sums.Unsettled,
	}).Error
}

func ListApiPlatformBills(platformID int) ([]ApiPlatformBill, error) {
	var rows []ApiPlatformBill
	if err := DB.Where("platform_id = ?", platformID).Order("id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return rows, nil
	}
	ids := make([]int, len(rows))
	for i, row := range rows {
		ids[i] = row.Id
	}
	type countRow struct {
		BillId int
		Total  int
	}
	var counts []countRow
	if err := DB.Model(&ApiPlatformBillItem{}).
		Select("bill_id, COUNT(*) AS total").
		Where("bill_id IN ?", ids).
		Group("bill_id").
		Scan(&counts).Error; err != nil {
		return nil, err
	}
	countMap := make(map[int]int, len(counts))
	for _, c := range counts {
		countMap[c.BillId] = c.Total
	}
	for i := range rows {
		rows[i].ChannelCount = countMap[rows[i].Id]
	}
	return rows, nil
}

func GetApiPlatformBill(platformID, billID int) (*ApiPlatformBill, error) {
	var row ApiPlatformBill
	err := DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("channel_id asc")
	}).Where("platform_id = ? AND id = ?", platformID, billID).First(&row).Error
	if err != nil {
		return nil, err
	}
	row.ChannelCount = len(row.Items)
	return &row, nil
}

func CreateApiPlatformBill(platformID int, rowIDs []int) (*ApiPlatformBill, error) {
	if len(rowIDs) == 0 {
		return nil, errors.New("请勾选渠道")
	}
	var bill ApiPlatformBill
	err := DB.Transaction(func(tx *gorm.DB) error {
		var channels []ApiPlatformChannel
		if err := tx.Where("platform_id = ? AND id IN ?", platformID, rowIDs).Find(&channels).Error; err != nil {
			return err
		}
		if len(channels) != len(uniqueInts(rowIDs)) {
			return errors.New("包含不属于该台子的渠道")
		}
		pending, err := pendingUpstreamIDsTx(tx, platformID)
		if err != nil {
			return err
		}
		items := make([]ApiPlatformBillItem, 0, len(channels))
		var payQuota, amount float64
		for _, ch := range channels {
			if _, ok := pending[ch.ChannelId]; ok {
				return fmt.Errorf("渠道 %d 已在待支付账单中", ch.ChannelId)
			}
			cost, sale := 0.0, 0.0
			if ch.CostPrice != nil {
				cost = *ch.CostPrice
			}
			if ch.SalePrice != nil {
				sale = *ch.SalePrice
			}
			lineAmount := sale * ch.UnsettledAmount
			payQuota += ch.UnsettledAmount
			amount += lineAmount
			items = append(items, ApiPlatformBillItem{
				ChannelRowId:  ch.Id,
				ChannelId:     ch.ChannelId,
				Name:          ch.Name,
				FilterName:    ch.FilterName,
				UsedAmount:    ch.UsedAmount,
				SettledAmount: ch.SettledAmount,
				PayQuota:      ch.UnsettledAmount,
				CostPrice:     cost,
				SalePrice:     sale,
				AmountCny:     lineAmount,
			})
		}
		billNo, err := nextBillNoTx(tx)
		if err != nil {
			return err
		}
		bill = ApiPlatformBill{
			PlatformId: platformID,
			BillNo:     billNo,
			Status:     ApiBillPending,
			PayQuota:   payQuota,
			AmountCny:  amount,
		}
		if err := tx.Create(&bill).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].BillId = bill.Id
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		bill.Items = items
		bill.ChannelCount = len(items)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func PayApiPlatformBill(platformID, billID int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var bill ApiPlatformBill
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("platform_id = ? AND id = ?", platformID, billID).
			First(&bill).Error; err != nil {
			return err
		}
		if bill.Status != ApiBillPending {
			return errors.New("只有待支付账单可以标记已支付")
		}
		var items []ApiPlatformBillItem
		if err := tx.Where("bill_id = ?", bill.Id).Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			var ch ApiPlatformChannel
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("platform_id = ? AND id = ?", platformID, item.ChannelRowId).
				First(&ch).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("platform_id = ? AND channel_id = ?", platformID, item.ChannelId).
					First(&ch).Error
			}
			if err != nil {
				return fmt.Errorf("渠道 %d 不存在", item.ChannelId)
			}
			settled := ch.SettledAmount + item.PayQuota
			unsettled := ch.UsedAmount - settled
			if unsettled < 0 {
				unsettled = 0
			}
			if err := tx.Model(&ApiPlatformChannel{}).Where("id = ?", ch.Id).Updates(map[string]any{
				"settled_amount":   settled,
				"unsettled_amount": unsettled,
			}).Error; err != nil {
				return err
			}
		}
		now := time.Now()
		if err := tx.Model(&ApiPlatformBill{}).Where("id = ?", bill.Id).Updates(map[string]any{
			"status":  ApiBillPaid,
			"paid_at": now,
		}).Error; err != nil {
			return err
		}
		return recalcApiPlatformTx(tx, platformID)
	})
}

func VoidApiPlatformBill(platformID, billID int) error {
	var bill ApiPlatformBill
	if err := DB.Where("platform_id = ? AND id = ?", platformID, billID).First(&bill).Error; err != nil {
		return err
	}
	if bill.Status != ApiBillPending {
		return errors.New("只有待支付账单可以作废")
	}
	return DB.Model(&ApiPlatformBill{}).Where("id = ?", bill.Id).Update("status", ApiBillVoid).Error
}

func pendingUpstreamIDsTx(tx *gorm.DB, platformID int) (map[int]struct{}, error) {
	var ids []int
	err := tx.Table("api_platform_bill_items AS i").
		Joins("JOIN api_platform_bills AS b ON b.id = i.bill_id").
		Where("b.platform_id = ? AND b.status = ?", platformID, ApiBillPending).
		Pluck("i.channel_id", &ids).Error
	if err != nil {
		return nil, err
	}
	set := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set, nil
}

func nextBillNoTx(tx *gorm.DB) (string, error) {
	var last ApiPlatformBill
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id desc").First(&last).Error
	next := 1
	if err == nil {
		var n int
		if _, scanErr := fmt.Sscanf(strings.TrimPrefix(last.BillNo, "SNAP-"), "%d", &n); scanErr == nil && n >= next {
			next = n + 1
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	return fmt.Sprintf("SNAP-%04d", next), nil
}

func uniqueInts(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
