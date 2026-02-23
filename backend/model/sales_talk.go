package model

import (
	"errors"
	"time"
)

// SalesTalk 话术模型
type SalesTalk struct {
	Id        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title     string     `json:"title" gorm:"type:varchar(100);not null;comment:菜单标题"`
	Tag       string     `json:"tag" gorm:"type:varchar(100);default:null;index:idx_tag;comment:标签"`
	Sort      int        `json:"sort" gorm:"type:int;default:0;index:idx_sort;comment:排序权重，数值越小越靠前"`
	ZhContent string     `json:"zh_content" gorm:"type:text;not null;comment:中文内容"`
	EnContent string     `json:"en_content" gorm:"type:text;not null;comment:英文内容"`
	RuContent string     `json:"ru_content" gorm:"type:text;not null;comment:俄文内容"`
	CreatedAt *time.Time `json:"created_at"`
}

// GetSalesTalkList 获取话术列表（按 sort 升序）
func GetSalesTalkList(page, pageSize int, keyword, tag string) ([]SalesTalk, int64, error) {
	var list []SalesTalk
	var total int64

	query := DB.Model(&SalesTalk{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR zh_content LIKE ? OR en_content LIKE ? OR ru_content LIKE ?",
			like, like, like, like)
	}
	if tag != "" {
		query = query.Where("tag = ?", tag)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sort asc, id asc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetSalesTalkById 根据 ID 获取话术
func GetSalesTalkById(id int64) (*SalesTalk, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	var item SalesTalk
	if err := DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// Create 创建话术
func (s *SalesTalk) Create() error {
	if s.Title == "" {
		return errors.New("标题不能为空")
	}
	return DB.Create(s).Error
}

// Update 更新话术
func (s *SalesTalk) Update() error {
	if s.Id == 0 {
		return errors.New("id 为空")
	}
	if s.Title == "" {
		return errors.New("标题不能为空")
	}
	return DB.Model(s).Updates(map[string]interface{}{
		"title":      s.Title,
		"tag":        s.Tag,
		"sort":       s.Sort,
		"zh_content": s.ZhContent,
		"en_content": s.EnContent,
		"ru_content": s.RuContent,
	}).Error
}

// Delete 删除话术
func DeleteSalesTalkById(id int64) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	return DB.Delete(&SalesTalk{}, id).Error
}

// BatchUpdateSalesTalkTag 批量更新标签
func BatchUpdateSalesTalkTag(ids []int64, tag string) error {
	if len(ids) == 0 {
		return errors.New("ids 为空")
	}
	return DB.Model(&SalesTalk{}).Where("id IN ?", ids).Update("tag", tag).Error
}

