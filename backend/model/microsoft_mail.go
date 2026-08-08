package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MicrosoftMail 微软邮箱账号库存
type MicrosoftMail struct {
	Id               int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Account          string         `json:"account" gorm:"type:varchar(100);not null;uniqueIndex:idx_ms_mail_account"`
	Password         string         `json:"password" gorm:"type:varchar(100);comment:邮箱密码"`
	IsCheck          int            `json:"is_check" gorm:"type:tinyint(2);default:-1;comment:检查状态 -1未检查 1检查成功 2检查失败;index:idx_ms_mail_is_check"`
	CheckTime        *int64         `json:"check_time" gorm:"type:bigint(20);comment:检查时间"`
	NextCheckTime    *int64         `json:"next_check_time" gorm:"type:bigint(20);comment:下次检查时间;index:idx_ms_mail_next_check"`
	PurchaseDate     *int64         `json:"purchase_date" gorm:"type:bigint(20);comment:购买时间"`
	PurchasePrice    *float64       `json:"purchase_price" gorm:"type:decimal(10,2);comment:购买价格(成本)"`
	PurchaseFrom     string         `json:"purchase_from" gorm:"type:varchar(50);comment:购买平台"`
	PurchaseBy       string         `json:"purchase_by" gorm:"type:varchar(100);comment:卖家名称"`
	SellPrice        *float64       `json:"sell_price" gorm:"type:decimal(10,2);comment:出售价格"`
	SellDate         *int64         `json:"sell_date" gorm:"type:bigint(20);comment:出售时间"`
	SellTo           string         `json:"sell_to" gorm:"type:varchar(50);comment:出售对方"`
	SellOrderNo      string         `json:"sell_order_no" gorm:"type:varchar(100);comment:出售订单号"`
	SellStatus       int            `json:"sell_status" gorm:"type:tinyint(2);default:1;comment:出售状态 1未出售 2发货中 3已出售;index:idx_ms_mail_sell"`
	Status           int            `json:"status" gorm:"type:tinyint(2);default:1;comment:状态 -1软删 -2封禁 1正常 2禁用"`
	Token            string         `json:"token" gorm:"type:text;comment:microsoft refresh_token"`
	TwoFA            string         `json:"2fa" gorm:"column:2fa;type:varchar(100);comment:2fa"`
	ClientId         string         `json:"client_id" gorm:"type:varchar(200);comment:client_id"`
	MailUrl          string         `json:"mail_url" gorm:"type:varchar(50);comment:邮箱地址"`
	Remark           string         `json:"remark" gorm:"type:varchar(100);comment:备注"`
	FreezeStatus     int            `json:"freeze_status" gorm:"type:tinyint(2);default:-1;comment:冻结状态 -1未冻结 1已冻结"`
	FreezeTime       *int64         `json:"freeze_time" gorm:"type:bigint(20);comment:冻结时间"`
	FreezeRemark     string         `json:"freeze_remark" gorm:"type:varchar(200);comment:冻结备注"`
	AccountCardId    *int           `json:"account_card_id" gorm:"type:int;comment:所属账号卡密id;index:idx_ms_mail_card_id"`
	AccountCardTable string         `json:"account_card_table" gorm:"type:varchar(100);comment:所属账号卡密表名"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (MicrosoftMail) TableName() string {
	return "microsoft_mails"
}

func applyMsMailAccountSearchFilter(query *gorm.DB, accounts []string) *gorm.DB {
	if len(accounts) == 0 {
		return query
	}
	if len(accounts) == 1 {
		kw := accounts[0]
		return query.Where("account LIKE ? OR mail_url LIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	return query.Where("account IN ?", accounts)
}

// GetMicrosoftMailList 分页列表
func GetMicrosoftMailList(listType string, page, pageSize int, accounts []string, isCheck int, purchaseDate int64, sellTo, purchaseBy string) ([]MicrosoftMail, int64, error) {
	var mails []MicrosoftMail
	var total int64

	query := DB.Model(&MicrosoftMail{}).Where("status != ?", CardStatusDeleted)

	switch listType {
	case "unsold":
		query = query.Where("sell_status IN ?", []int{1, 2})
	case "sold":
		query = query.Where("sell_status = ?", 3)
	default:
		query = query.Where("sell_status IN ?", []int{1, 2})
	}

	if isCheck != 0 {
		query = query.Where("is_check = ?", isCheck)
	}
	if purchaseDate > 0 && listType == "sold" {
		query = query.Where("sell_date >= ? AND sell_date < ?", purchaseDate, purchaseDate+86400)
	}
	if purchaseDate > 0 && listType == "unsold" {
		query = query.Where("purchase_date >= ? AND purchase_date < ?", purchaseDate, purchaseDate+86400)
	}
	if st := strings.TrimSpace(sellTo); st != "" && listType == "sold" {
		query = query.Where("sell_to LIKE ?", "%"+st+"%")
	}
	if pb := strings.TrimSpace(purchaseBy); pb != "" && listType == "sold" {
		query = query.Where("purchase_by LIKE ?", "%"+pb+"%")
	}

	query = applyMsMailAccountSearchFilter(query, accounts)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	orderClause := "sell_status DESC, id ASC"
	if listType == "sold" {
		orderClause = "sell_date DESC, id DESC"
	}
	if err := query.Order(orderClause).Offset(offset).Limit(pageSize).Find(&mails).Error; err != nil {
		return nil, 0, err
	}
	return mails, total, nil
}

// GetAllMicrosoftMailsForExport 导出全部符合条件记录
func GetAllMicrosoftMailsForExport(listType string, accounts []string, isCheck int, purchaseDate int64, sellTo, purchaseBy string) ([]MicrosoftMail, error) {
	var mails []MicrosoftMail
	query := DB.Model(&MicrosoftMail{}).Where("status != ?", CardStatusDeleted)

	switch listType {
	case "unsold":
		query = query.Where("sell_status IN ?", []int{1, 2})
	case "sold":
		query = query.Where("sell_status = ?", 3)
	default:
		query = query.Where("sell_status IN ?", []int{1, 2})
	}

	if isCheck != 0 {
		query = query.Where("is_check = ?", isCheck)
	}
	if purchaseDate > 0 && listType == "sold" {
		query = query.Where("sell_date >= ? AND sell_date < ?", purchaseDate, purchaseDate+86400)
	}
	if purchaseDate > 0 && listType == "unsold" {
		query = query.Where("purchase_date >= ? AND purchase_date < ?", purchaseDate, purchaseDate+86400)
	}
	if st := strings.TrimSpace(sellTo); st != "" && listType == "sold" {
		query = query.Where("sell_to LIKE ?", "%"+st+"%")
	}
	if pb := strings.TrimSpace(purchaseBy); pb != "" && listType == "sold" {
		query = query.Where("purchase_by LIKE ?", "%"+pb+"%")
	}
	query = applyMsMailAccountSearchFilter(query, accounts)

	err := query.Order("id ASC").Find(&mails).Error
	return mails, err
}

// GetMicrosoftMailById 按 ID 获取
func GetMicrosoftMailById(id int) (*MicrosoftMail, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	var mail MicrosoftMail
	err := DB.Where("status != ?", CardStatusDeleted).First(&mail, "id = ?", id).Error
	return &mail, err
}

// GetMicrosoftMailsByIds 按 ID 列表获取（需有 token）
func GetMicrosoftMailsByIds(ids []int) ([]*MicrosoftMail, error) {
	if len(ids) == 0 {
		return nil, errors.New("id 列表为空")
	}
	var mails []*MicrosoftMail
	err := DB.Where("id IN ? AND status != ? AND token IS NOT NULL AND token != ''", ids, CardStatusDeleted).
		Find(&mails).Error
	return mails, err
}

// CreateMicrosoftMail 创建
func CreateMicrosoftMail(mail *MicrosoftMail) error {
	if mail.Account == "" {
		return errors.New("账号不能为空")
	}
	var count int64
	DB.Model(&MicrosoftMail{}).Where("account = ?", mail.Account).Count(&count)
	if count > 0 {
		return errors.New("账号已存在")
	}
	if mail.SellStatus == 0 {
		mail.SellStatus = 1
	}
	if mail.Status == 0 {
		mail.Status = CardStatusNormal
	}
	if mail.IsCheck == 0 {
		mail.IsCheck = -1
	}
	if mail.FreezeStatus == 0 {
		mail.FreezeStatus = -1
	}
	return DB.Create(mail).Error
}

// UpdateMicrosoftMail 更新可编辑字段
func UpdateMicrosoftMail(id int, mail *MicrosoftMail) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	if mail.Account == "" {
		return errors.New("账号不能为空")
	}
	var count int64
	DB.Model(&MicrosoftMail{}).Where("account = ? AND id != ?", mail.Account, id).Count(&count)
	if count > 0 {
		return errors.New("账号已存在")
	}
	updates := map[string]interface{}{
		"account":            mail.Account,
		"password":           mail.Password,
		"purchase_price":     mail.PurchasePrice,
		"purchase_from":      mail.PurchaseFrom,
		"purchase_by":        mail.PurchaseBy,
		"sell_price":         mail.SellPrice,
		"sell_to":            mail.SellTo,
		"sell_status":        mail.SellStatus,
		"status":             mail.Status,
		"token":              mail.Token,
		"2fa":                mail.TwoFA,
		"client_id":          mail.ClientId,
		"mail_url":           mail.MailUrl,
		"remark":             mail.Remark,
		"account_card_id":    mail.AccountCardId,
		"account_card_table": mail.AccountCardTable,
	}
	return DB.Model(&MicrosoftMail{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteMicrosoftMail 软删除（GORM DeletedAt）
func DeleteMicrosoftMail(id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	return DB.Where("id = ?", id).Delete(&MicrosoftMail{}).Error
}

func normalizeMicrosoftMailForCreate(mail *MicrosoftMail) {
	if mail.SellStatus == 0 {
		mail.SellStatus = 1
	}
	if mail.Status == 0 {
		mail.Status = CardStatusNormal
	}
	if mail.IsCheck == 0 {
		mail.IsCheck = -1
	}
	if mail.FreezeStatus == 0 {
		mail.FreezeStatus = -1
	}
}

// BatchCreateMicrosoftMails 批量导入（存在则更新非零字段）
// 分块处理并关闭 PrepareStmt，避免大批量导入触发 max_prepared_stmt_count
func BatchCreateMicrosoftMails(mails []MicrosoftMail) error {
	if len(mails) == 0 {
		return errors.New("没有要导入的数据")
	}

	// 同批重复账号以后者为准，减少无效写入
	deduped := make([]MicrosoftMail, 0, len(mails))
	seen := make(map[string]int, len(mails))
	for _, mail := range mails {
		account := strings.TrimSpace(mail.Account)
		if account == "" {
			continue
		}
		mail.Account = account
		if idx, ok := seen[account]; ok {
			deduped[idx] = mail
			continue
		}
		seen[account] = len(deduped)
		deduped = append(deduped, mail)
	}
	if len(deduped) == 0 {
		return errors.New("没有要导入的数据")
	}

	const chunkSize = 500
	const createBatchSize = 100

	return DB.Session(&gorm.Session{PrepareStmt: false}).Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(deduped); i += chunkSize {
			end := i + chunkSize
			if end > len(deduped) {
				end = len(deduped)
			}
			chunk := deduped[i:end]

			accounts := make([]string, 0, len(chunk))
			for _, mail := range chunk {
				accounts = append(accounts, mail.Account)
			}

			var existedList []MicrosoftMail
			if err := tx.Select("id", "account").Where("account IN ?", accounts).Find(&existedList).Error; err != nil {
				return fmt.Errorf("查询失败: %v", err)
			}
			existedMap := make(map[string]int, len(existedList))
			for _, item := range existedList {
				existedMap[item.Account] = item.Id
			}

			toCreate := make([]MicrosoftMail, 0, len(chunk))
			for _, mail := range chunk {
				if id, ok := existedMap[mail.Account]; ok {
					if err := tx.Where("id = ?", id).Updates(&mail).Error; err != nil {
						return fmt.Errorf("更新失败: %s, %v", mail.Account, err)
					}
					continue
				}
				normalizeMicrosoftMailForCreate(&mail)
				toCreate = append(toCreate, mail)
			}
			if len(toCreate) == 0 {
				continue
			}
			if err := tx.CreateInBatches(toCreate, createBatchSize).Error; err != nil {
				return fmt.Errorf("批量创建失败: %v", err)
			}
		}
		return nil
	})
}

// PickupMicrosoftMail 取货：FIFO 选出一条未售记录，标记为发货中
func PickupMicrosoftMail(format string) (*MicrosoftMail, error) {
	var mail MicrosoftMail
	err := DB.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&MicrosoftMail{}).
			Where("sell_status = ?", 1).
			Where("status NOT IN ?", []int{CardStatusDeleted, CardStatusBanned}).
			Order("id ASC")
		if format == "reverse" {
			query = query.Where("token != ?", "")
		}
		if err := query.First(&mail).Error; err != nil {
			return errors.New("没有可用的邮箱")
		}
		if err := tx.Model(&MicrosoftMail{}).Where("id = ?", mail.Id).Update("sell_status", 2).Error; err != nil {
			return err
		}
		mail.SellStatus = 2
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &mail, nil
}

// BatchPickupMicrosoftMails 批量取货：直接标记为已出售
func BatchPickupMicrosoftMails(ids []int, sellPrice *float64, sellTo string) (int64, error) {
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
	result := DB.Model(&MicrosoftMail{}).
		Where("id IN ? AND status NOT IN ? AND sell_status IN ?", ids, []int{CardStatusDeleted, CardStatusBanned}, []int{1, 2}).
		Updates(updates)
	return result.RowsAffected, result.Error
}

// CompleteMicrosoftMailPickup 完成取货
func CompleteMicrosoftMailPickup(id int, sellPrice *float64, sellTo string) error {
	if id == 0 {
		return errors.New("id 为空")
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
	return DB.Model(&MicrosoftMail{}).Where("id = ? AND status != ?", id, CardStatusDeleted).Updates(updates).Error
}

// RollbackMicrosoftMailPickup 回滚发货中 → 未出售
func RollbackMicrosoftMailPickup(id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	result := DB.Model(&MicrosoftMail{}).
		Where("id = ? AND status != ? AND sell_status = ?", id, CardStatusDeleted, 2).
		Update("sell_status", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("记录不存在或状态不是发货中")
	}
	return nil
}

// RollbackMicrosoftMailSold 回滚已售 → 未出售
func RollbackMicrosoftMailSold(id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	result := DB.Model(&MicrosoftMail{}).
		Where("id = ? AND status != ? AND sell_status = ?", id, CardStatusDeleted, 3).
		Updates(map[string]interface{}{
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
		return errors.New("记录不存在或状态不是已出售")
	}
	return nil
}

// BatchDeleteMicrosoftMails 批量软删（status=-1）
func BatchDeleteMicrosoftMails(ids []int) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("未选择任何记录")
	}
	result := DB.Model(&MicrosoftMail{}).Where("id IN ?", ids).Update("status", CardStatusDeleted)
	return result.RowsAffected, result.Error
}

// UpdateMicrosoftMailRemark 更新备注
func UpdateMicrosoftMailRemark(id int, remark string) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	return DB.Model(&MicrosoftMail{}).Where("id = ?", id).Update("remark", remark).Error
}

func microsoftMailCheckBaseQuery() *gorm.DB {
	return DB.Model(&MicrosoftMail{}).
		Where("status NOT IN ?", []int{CardStatusDeleted, CardStatusBanned}).
		Where("token IS NOT NULL AND token != ''").
		Where("client_id IS NOT NULL AND client_id != ''")
}

// GetUncheckedMicrosoftMailsForCheck 未检查过的记录（is_check=-1），按 id 分页
func GetUncheckedMicrosoftMailsForCheck(limit int) ([]*MicrosoftMail, error) {
	var mails []*MicrosoftMail
	query := microsoftMailCheckBaseQuery().
		Where("is_check = ?", -1).
		Order("id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&mails).Error
	return mails, err
}

// GetDueMicrosoftMailsForCheck 已检查过且到达下次检查时间的记录
func GetDueMicrosoftMailsForCheck(limit int) ([]*MicrosoftMail, error) {
	var mails []*MicrosoftMail
	now := time.Now().Unix()
	query := microsoftMailCheckBaseQuery().
		Where("is_check != ?", -1).
		Where("next_check_time IS NOT NULL AND next_check_time <= ?", now).
		Order("next_check_time ASC, id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&mails).Error
	return mails, err
}

// ApplyMicrosoftMailCheckSuccess 检查成功：回写 token 与检查时间
func ApplyMicrosoftMailCheckSuccess(id int, newRefreshToken string, nextIntervalDays int) error {
	now := time.Now().Unix()
	next := now + int64(nextIntervalDays)*86400
	updates := map[string]interface{}{
		"is_check":        1,
		"check_time":      now,
		"next_check_time": next,
	}
	if strings.TrimSpace(newRefreshToken) != "" {
		updates["token"] = newRefreshToken
	}
	return DB.Model(&MicrosoftMail{}).Where("id = ?", id).Updates(updates).Error
}

// ApplyMicrosoftMailCheckFailure 检查失败
// permanent=true（如 invalid_grant）清空 next_check_time，不再自动复查；否则短推迟 retryHours 小时
func ApplyMicrosoftMailCheckFailure(id int, permanent bool, retryHours int) error {
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"is_check":   2,
		"check_time": now,
	}
	if permanent {
		updates["next_check_time"] = nil
	} else {
		if retryHours <= 0 {
			retryHours = 6
		}
		next := now + int64(retryHours)*3600
		updates["next_check_time"] = next
	}
	return DB.Model(&MicrosoftMail{}).Where("id = ?", id).Updates(updates).Error
}
