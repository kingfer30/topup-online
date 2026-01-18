package model

import (
	"errors"
	"time"
)

// MirrorCard 镜像卡密
type MirrorCard struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	Username     string     `json:"username" gorm:"type:varchar(100);not null"`
	Password     string     `json:"password" gorm:"type:varchar(100);not null"`
	BindStatus   int        `json:"bind_status" gorm:"type:int;default:0;comment:'0-未绑定 1-已绑定'"`
	BindUserId   int        `json:"bind_user_id" gorm:"type:int;default:0;index"`
	BindUserName string     `json:"bind_user_name" gorm:"-"` // 不存储到数据库，仅用于返回
	BindTime     *time.Time `json:"bind_time" gorm:"type:datetime"`
	Status       int        `json:"status" gorm:"type:int;default:1;comment:'1-启用 2-禁用'"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

const (
	MirrorCardBindStatusUnbound = 0
	MirrorCardBindStatusBound   = 1

	MirrorCardStatusEnabled  = 1
	MirrorCardStatusDisabled = 2
)

// GetMirrorCardList 获取镜像卡密列表
func GetMirrorCardList(page, pageSize int, keyword string) ([]*MirrorCard, int64, error) {
	var cards []*MirrorCard
	var total int64

	query := DB.Model(&MirrorCard{})

	// 搜索：支持用户名、密码和绑定用户账号搜索
	if keyword != "" {
		// 先查找匹配的用户ID
		var userIds []int
		DB.Model(&User{}).Where("username LIKE ?", "%"+keyword+"%").Pluck("id", &userIds)
		
		if len(userIds) > 0 {
			query = query.Where("username LIKE ? OR password LIKE ? OR bind_user_id IN ?", "%"+keyword+"%", "%"+keyword+"%", userIds)
		} else {
			query = query.Where("username LIKE ? OR password LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	// 填充绑定用户的用户名
	for _, card := range cards {
		if card.BindUserId > 0 {
			var user User
			if err := DB.Select("username").First(&user, card.BindUserId).Error; err == nil {
				card.BindUserName = user.Username
			}
		}
	}

	return cards, total, nil
}

// GetMirrorCardById 根据ID获取镜像卡密
func GetMirrorCardById(id int) (*MirrorCard, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	var card MirrorCard
	if err := DB.First(&card, id).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

// CreateMirrorCard 创建镜像卡密
func CreateMirrorCard(card *MirrorCard) error {
	if card.Username == "" || card.Password == "" {
		return errors.New("用户名或密码为空")
	}
	return DB.Create(card).Error
}

// UpdateMirrorCard 更新镜像卡密
func UpdateMirrorCard(card *MirrorCard) error {
	if card.ID == 0 {
		return errors.New("id 为空")
	}
	return DB.Model(card).Updates(card).Error
}

// DeleteMirrorCard 删除镜像卡密
func DeleteMirrorCard(id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}

	// 先检查是否有用户绑定了这个卡密
	var card MirrorCard
	if err := DB.First(&card, id).Error; err != nil {
		return err
	}

	// 如果已绑定，需要解除用户的绑定关系
	if card.BindStatus == MirrorCardBindStatusBound && card.BindUserId > 0 {
		// 清空用户表中的 mirror_card_id
		if err := DB.Model(&User{}).Where("id = ?", card.BindUserId).Update("mirror_card_id", 0).Error; err != nil {
			return err
		}
	}

	// 删除卡密记录
	return DB.Delete(&MirrorCard{}, id).Error
}

// GetAvailableMirrorCard 获取一个可用的镜像卡密（状态=启用，绑定状态=未绑）
func GetAvailableMirrorCard() (*MirrorCard, error) {
	var card MirrorCard
	err := DB.Where("status = ? AND bind_status = ?", MirrorCardStatusEnabled, MirrorCardBindStatusUnbound).
		First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// BindMirrorCardToUser 绑定镜像卡密到用户
func BindMirrorCardToUser(cardId, userId int) error {
	if cardId == 0 || userId == 0 {
		return errors.New("cardId 或 userId 为空")
	}

	now := time.Now()
	return DB.Model(&MirrorCard{}).Where("id = ?", cardId).Updates(map[string]interface{}{
		"bind_status":  MirrorCardBindStatusBound,
		"bind_user_id": userId,
		"bind_time":    now,
	}).Error
}

// UnbindMirrorCard 解绑镜像卡密
func UnbindMirrorCard(cardId int) error {
	if cardId == 0 {
		return errors.New("cardId 为空")
	}

	return DB.Model(&MirrorCard{}).Where("id = ?", cardId).Updates(map[string]interface{}{
		"bind_status":  MirrorCardBindStatusUnbound,
		"bind_user_id": 0,
		"bind_time":    nil,
	}).Error
}

// BatchCreateMirrorCards 批量创建镜像卡密
func BatchCreateMirrorCards(cards []*MirrorCard) error {
	if len(cards) == 0 {
		return errors.New("卡密列表为空")
	}
	return DB.Create(&cards).Error
}
