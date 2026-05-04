package model

import (
	"time"

	"gorm.io/gorm"
)

// GptCdk 自建CDK
type GptCdk struct {
	Id           int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Key          string         `json:"key" gorm:"type:varchar(200);not null;uniqueIndex;comment:CDK密钥"`
	GptMail      string         `json:"gpt_mail" gorm:"type:varchar(100);comment:目的账号"`
	Buyer        string         `json:"buyer" gorm:"type:varchar(100);comment:买家"`
	SellStatus   int            `json:"sell_status" gorm:"type:tinyint(2);default:1;comment:出售状态 -1作废 1待售 2已售;index:idx_sell_status"`
	ExpireTime   *int64         `json:"expire_time" gorm:"type:bigint(20);comment:过期时间"`
	UseStatus    int            `json:"use_status" gorm:"type:tinyint(2);default:1;comment:使用状态 1未使用 2占用中 3已使用;index:idx_use_status"`
	UseTime      *int64         `json:"use_time" gorm:"type:bigint(20);comment:使用时间"`
	CardId       *int           `json:"card_id" gorm:"comment:关联gpt_cards.id"`
	SubStartTime *int64         `json:"sub_start_time" gorm:"type:bigint(20);comment:订阅开始时间"`
	SubEndTime   *int64         `json:"sub_end_time" gorm:"type:bigint(20);comment:订阅结束时间"`
	SubResult    string         `json:"sub_result" gorm:"type:varchar(500);comment:订阅结果"`
	IpAddr       string         `json:"ip_addr" gorm:"type:varchar(50);comment:IP地址"`
	DeviceInfo   string         `json:"device_info" gorm:"type:varchar(500);comment:设备信息"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (GptCdk) TableName() string {
	return "gpt_cdk"
}

// GetGptCdkList 分页查询CDK列表
func GetGptCdkList(page, pageSize int, sellStatus, useStatus int, keyword string) ([]GptCdk, int64, error) {
	var list []GptCdk
	var total int64

	query := DB.Model(&GptCdk{})

	if sellStatus != 0 {
		query = query.Where("sell_status = ?", sellStatus)
	}
	if useStatus != 0 {
		query = query.Where("use_status = ?", useStatus)
	}
	if keyword != "" {
		query = query.Where("`key` LIKE ? OR gpt_mail LIKE ? OR buyer LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// BatchCreateGptCdks 批量创建CDK
func BatchCreateGptCdks(cdks []GptCdk) error {
	if len(cdks) == 0 {
		return nil
	}
	return DB.CreateInBatches(cdks, 100).Error
}

// UpdateGptCdk 更新CDK
func UpdateGptCdk(id int, updates map[string]interface{}) error {
	return DB.Model(&GptCdk{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteGptCdk 软删除CDK
func DeleteGptCdk(id int) error {
	return DB.Delete(&GptCdk{}, id).Error
}

// BatchDeleteGptCdks 批量软删除CDK
func BatchDeleteGptCdks(ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return DB.Delete(&GptCdk{}, ids).Error
}

// GetGptCdkByKey 按key查询
func GetGptCdkByKey(key string) (*GptCdk, error) {
	var cdk GptCdk
	err := DB.Where("`key` = ?", key).First(&cdk).Error
	if err != nil {
		return nil, err
	}
	return &cdk, nil
}

// GetGptCdkByKeyForUpdate 在事务中按key查询并加行锁
func GetGptCdkByKeyForUpdate(tx *gorm.DB, key string) (*GptCdk, error) {
	var cdk GptCdk
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("`key` = ?", key).
		First(&cdk).Error
	if err != nil {
		return nil, err
	}
	return &cdk, nil
}

// OccupyCdk 将CDK使用状态设为占用中（2），需在事务内调用
func OccupyCdk(tx *gorm.DB, id int) error {
	return tx.Model(&GptCdk{}).Where("id = ? AND use_status = 1", id).
		Update("use_status", 2).Error
}

// ReleaseCdk 回滚CDK使用状态为未使用（1）
func ReleaseCdk(id int) error {
	return DB.Model(&GptCdk{}).Where("id = ? AND use_status = 2", id).
		Update("use_status", 1).Error
}

// CompleteCdk 将CDK使用状态设为已使用（3），记录使用时间和关联卡密ID
func CompleteCdk(id, cardId int, useTime int64) error {
	return DB.Model(&GptCdk{}).Where("id = ?", id).Updates(map[string]interface{}{
		"use_status": 3,
		"use_time":   useTime,
		"card_id":    cardId,
	}).Error
}
