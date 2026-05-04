package model

import (
	"time"

	"gorm.io/gorm"
)

// GptCard GPT卡密（第三方供应商）
type GptCard struct {
	Id           int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Supplier     string         `json:"supplier" gorm:"type:varchar(50);not null;comment:供应商;index:idx_supplier"`
	Key          string         `json:"key" gorm:"type:varchar(500);not null;uniqueIndex;comment:卡密"`
	GptMail      string         `json:"gpt_mail" gorm:"type:varchar(100);comment:目的账号"`
	Buyer        string         `json:"buyer" gorm:"type:varchar(100);comment:买家"`
	Status       int            `json:"status" gorm:"type:tinyint(2);default:1;comment:状态 -1作废 1待使用 2已使用 3占用中;index:idx_status"`
	ImportTime   *int64         `json:"import_time" gorm:"type:bigint(20);comment:导入时间"`
	UseTime      *int64         `json:"use_time" gorm:"type:bigint(20);comment:使用时间"`
	SubStartTime *int64         `json:"sub_start_time" gorm:"type:bigint(20);comment:订阅开始时间"`
	SubEndTime   *int64         `json:"sub_end_time" gorm:"type:bigint(20);comment:订阅结束时间"`
	SubResult    string         `json:"sub_result" gorm:"type:varchar(500);comment:订阅结果"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (GptCard) TableName() string {
	return "gpt_cards"
}

// GetGptCardList 分页查询GPT卡密列表
func GetGptCardList(page, pageSize int, supplier string, status int, keyword string) ([]GptCard, int64, error) {
	var cards []GptCard
	var total int64

	query := DB.Model(&GptCard{})

	if supplier != "" {
		query = query.Where("supplier = ?", supplier)
	}
	if status != 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("`key` LIKE ? OR gpt_mail LIKE ? OR buyer LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

// BatchCreateGptCards 批量创建GPT卡密
func BatchCreateGptCards(cards []GptCard) error {
	if len(cards) == 0 {
		return nil
	}
	return DB.CreateInBatches(cards, 100).Error
}

// UpdateGptCard 更新GPT卡密
func UpdateGptCard(id int, updates map[string]interface{}) error {
	return DB.Model(&GptCard{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteGptCard 软删除GPT卡密
func DeleteGptCard(id int) error {
	return DB.Delete(&GptCard{}, id).Error
}

// BatchDeleteGptCards 批量软删除GPT卡密
func BatchDeleteGptCards(ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return DB.Delete(&GptCard{}, ids).Error
}

// GetGptCardById 按ID查询
func GetGptCardById(id int) (*GptCard, error) {
	var card GptCard
	err := DB.First(&card, id).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// GetGptCardsByIds 按ID列表批量查询
func GetGptCardsByIds(ids []int) ([]GptCard, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var cards []GptCard
	err := DB.Where("id IN ?", ids).Find(&cards).Error
	return cards, err
}

// PickRandomAvailableGptCard 在事务中随机取一张待使用的卡密并加行锁
func PickRandomAvailableGptCard(tx *gorm.DB) (*GptCard, error) {
	var card GptCard
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("status = 1").
		Order("RAND()").
		Limit(1).
		First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// OccupyGptCard 将卡密状态设为占用中（3），需在事务内调用
func OccupyGptCard(tx *gorm.DB, id int) error {
	return tx.Model(&GptCard{}).Where("id = ? AND status = 1", id).
		Update("status", 3).Error
}

// ReleaseGptCard 回滚卡密状态为待使用（1）
func ReleaseGptCard(id int) error {
	return DB.Model(&GptCard{}).Where("id = ? AND status = 3", id).
		Update("status", 1).Error
}

// CompleteGptCard 将卡密状态设为已使用（2），记录使用时间和订阅结果
func CompleteGptCard(id int, useTime int64, subResult string) error {
	return DB.Model(&GptCard{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     2,
		"use_time":   useTime,
		"sub_result": subResult,
	}).Error
}
