package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// AdConfig 广告配置
type AdConfig struct {
	Id         int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title      string     `json:"title" gorm:"type:varchar(500);not null;comment:广告标题/顶部通知文案"`
	Image      string     `json:"image" gorm:"type:varchar(500);default:'';comment:图片URL（侧栏广告必填）"`
	Link       string     `json:"link" gorm:"type:varchar(500);default:'';comment:跳转链接"`
	Position   string     `json:"position" gorm:"type:varchar(16);not null;index:idx_ad_position;comment:left/right/top"`
	Buyer      string     `json:"buyer" gorm:"type:varchar(100);default:'';comment:广告位购买人"`
	ClickCount int64      `json:"click_count" gorm:"type:bigint;default:0;comment:点击次数"`
	Sort       int        `json:"sort" gorm:"type:int;default:0;index:idx_ad_sort;comment:排序，越小越优先"`
	Status     int        `json:"status" gorm:"type:tinyint;default:1;comment:1启用 0禁用"`
	StartTime  time.Time  `json:"start_time" gorm:"not null;comment:生效时间"`
	EndTime    time.Time  `json:"end_time" gorm:"not null;comment:失效时间"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

func (AdConfig) TableName() string {
	return "ad_configs"
}

// GetAdConfigList 获取广告配置列表
func GetAdConfigList(page, pageSize int, keyword, position string) ([]AdConfig, int64, error) {
	var list []AdConfig
	var total int64

	query := DB.Model(&AdConfig{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR buyer LIKE ? OR link LIKE ?", like, like, like)
	}
	if position != "" {
		query = query.Where("position = ?", position)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sort asc, id desc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetAdConfigById 根据 ID 获取广告配置
func GetAdConfigById(id int64) (*AdConfig, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	var item AdConfig
	if err := DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// ListActiveAdsByPosition 获取某侧当前全部生效广告（按 sort 升序）
func ListActiveAdsByPosition(position string, now time.Time) ([]AdConfig, error) {
	var list []AdConfig
	err := DB.Where("position = ? AND status = ? AND start_time <= ? AND end_time >= ?",
		position, 1, now, now).
		Order("sort asc, id desc").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetLatestActiveAdByPosition 获取某位置当前生效的最新一条（按 id 降序）
func GetLatestActiveAdByPosition(position string, now time.Time) (*AdConfig, error) {
	var item AdConfig
	err := DB.Where("position = ? AND status = ? AND start_time <= ? AND end_time >= ?",
		position, 1, now, now).
		Order("id desc").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func validateAdConfig(a *AdConfig) error {
	if a.Title == "" {
		return errors.New("标题不能为空")
	}
	if a.Position != "left" && a.Position != "right" && a.Position != "top" {
		return errors.New("位置必须为 left、right 或 top")
	}
	if a.Position == "left" || a.Position == "right" {
		if a.Image == "" {
			return errors.New("侧栏广告图片不能为空")
		}
		if a.Link == "" {
			return errors.New("侧栏广告链接不能为空")
		}
	}
	if a.EndTime.Before(a.StartTime) {
		return errors.New("失效时间不能早于生效时间")
	}
	return nil
}

// Create 创建广告配置
func (a *AdConfig) Create() error {
	if err := validateAdConfig(a); err != nil {
		return err
	}
	if a.Status != 0 && a.Status != 1 {
		a.Status = 1
	}
	a.ClickCount = 0
	return DB.Create(a).Error
}

// Update 更新广告配置（不改 click_count）
func (a *AdConfig) Update() error {
	if a.Id == 0 {
		return errors.New("id 为空")
	}
	if err := validateAdConfig(a); err != nil {
		return err
	}
	return DB.Model(a).Updates(map[string]interface{}{
		"title":      a.Title,
		"image":      a.Image,
		"link":       a.Link,
		"position":   a.Position,
		"buyer":      a.Buyer,
		"sort":       a.Sort,
		"status":     a.Status,
		"start_time": a.StartTime,
		"end_time":   a.EndTime,
	}).Error
}

// DeleteAdConfigById 删除广告配置
func DeleteAdConfigById(id int64) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	return DB.Delete(&AdConfig{}, id).Error
}

// IncrementAdClickCount 点击次数 +1
func IncrementAdClickCount(id int64) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	return DB.Model(&AdConfig{}).Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}
