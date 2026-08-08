package model

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Hold 状态
const (
	GptRtHoldStatusHeld      = "held"
	GptRtHoldStatusConfirmed = "confirmed"
	GptRtHoldStatusReleased  = "released"
	GptRtHoldStatusExpired   = "expired"
	GptRtHoldStatusRevoked   = "revoked"
)

// GptRtLicense GPT RT 提取工具许可证
type GptRtLicense struct {
	Id               int        `json:"id" gorm:"primaryKey;autoIncrement"`
	LicenseKey       string     `json:"license_key" gorm:"type:varchar(64);uniqueIndex;not null;comment:激活码"`
	AppId            string     `json:"app_id" gorm:"type:varchar(64);not null;comment:应用标识"`
	Customer         string     `json:"customer" gorm:"type:varchar(128);comment:客户备注"`
	Status           int        `json:"status" gorm:"type:tinyint;default:1;comment:1正常 0禁用"`
	ExpiresAt        time.Time  `json:"expires_at" gorm:"not null;comment:到期时间"`
	ActivatedAt      *time.Time `json:"activated_at"`
	LastVerifiedAt   *time.Time `json:"last_verified_at"`
	LastUsingIP      string     `json:"last_using_ip" gorm:"type:varchar(64);comment:最后使用IP"`
	UsedCount        int        `json:"used_count" gorm:"type:int;default:0;comment:已提取RT次数"`
	AvailableCount   int        `json:"available_count" gorm:"type:int;not null;comment:剩余可提取RT次数"`
	HeldCount        int        `json:"held_count" gorm:"type:int;default:0;comment:预占中次数"`
	MaxDevices       int        `json:"max_devices" gorm:"type:int;default:1;comment:可绑定设备数量"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	BoundDeviceCount int        `json:"bound_device_count" gorm:"-"`
}

func (GptRtLicense) TableName() string {
	return "gpt_rt_licenses"
}

// GptRtLicenseDevice 许可证设备绑定记录
type GptRtLicenseDevice struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	LicenseId int       `json:"license_id" gorm:"not null;uniqueIndex:uk_license_machine;comment:许可证ID"`
	MachineId string    `json:"machine_id" gorm:"type:varchar(128);not null;uniqueIndex:uk_license_machine;comment:机器指纹"`
	LoginIP   string    `json:"login_ip" gorm:"type:varchar(64);comment:登录IP"`
	UserAgent string    `json:"user_agent" gorm:"type:varchar(512);comment:User-Agent"`
	BoundAt   time.Time `json:"bound_at" gorm:"not null;comment:绑定时间"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (GptRtLicenseDevice) TableName() string {
	return "gpt_rt_license_devices"
}

// GptRtLicenseHold 许可证配额预占记录
type GptRtLicenseHold struct {
	HoldId    string    `json:"hold_id" gorm:"primaryKey;type:varchar(64);comment:预占ID"`
	LicenseId int       `json:"license_id" gorm:"index;not null;comment:许可证ID"`
	MachineId string    `json:"machine_id" gorm:"type:varchar(128);not null;comment:机器指纹"`
	Email     string    `json:"email" gorm:"type:varchar(256);comment:邮箱(可选)"`
	Status    string    `json:"status" gorm:"type:varchar(32);not null;index;comment:held|confirmed|released|expired|revoked"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index;comment:预占过期时间"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (GptRtLicenseHold) TableName() string {
	return "gpt_rt_license_holds"
}

func generateGptRtHoldID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

var ErrGptRtLicenseDisabled = errors.New("许可证已禁用")

// GetGptRtLicenseList 分页查询
func GetGptRtLicenseList(page, pageSize int, status int, keyword string) ([]GptRtLicense, int64, error) {
	var list []GptRtLicense
	var total int64

	query := DB.Model(&GptRtLicense{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("license_key LIKE ? OR customer LIKE ? OR app_id LIKE ? OR last_using_ip LIKE ?",
			kw, kw, kw, kw)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if len(list) > 0 {
		ids := make([]int, len(list))
		for i := range list {
			ids[i] = list[i].Id
		}
		type countRow struct {
			LicenseId int
			Cnt       int64
		}
		var rows []countRow
		_ = DB.Model(&GptRtLicenseDevice{}).
			Select("license_id, COUNT(*) as cnt").
			Where("license_id IN ?", ids).
			Group("license_id").
			Scan(&rows).Error
		countMap := make(map[int]int, len(rows))
		for _, r := range rows {
			countMap[r.LicenseId] = int(r.Cnt)
		}
		for i := range list {
			list[i].BoundDeviceCount = countMap[list[i].Id]
		}
	}
	return list, total, nil
}

// GetGptRtLicenseByID 按 ID 查询
func GetGptRtLicenseByID(id int) (*GptRtLicense, error) {
	var lic GptRtLicense
	if err := DB.First(&lic, id).Error; err != nil {
		return nil, err
	}
	return &lic, nil
}

// GetGptRtLicenseByKey 按激活码查询
func GetGptRtLicenseByKey(key string) (*GptRtLicense, error) {
	var lic GptRtLicense
	if err := DB.Where("license_key = ?", key).First(&lic).Error; err != nil {
		return nil, err
	}
	return &lic, nil
}

// CreateGptRtLicense 创建
func CreateGptRtLicense(lic *GptRtLicense) error {
	return DB.Create(lic).Error
}

// UpdateGptRtLicense 更新
func UpdateGptRtLicense(id int, updates map[string]interface{}) error {
	return DB.Model(&GptRtLicense{}).Where("id = ?", id).Updates(updates).Error
}

// ExprAddAvailableCount 原子增加可用数量
func ExprAddAvailableCount(delta int) interface{} {
	return gorm.Expr("available_count + ?", delta)
}

// DeleteGptRtLicense 删除（同时清理设备绑定与预占）
func DeleteGptRtLicense(id int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("license_id = ?", id).Delete(&GptRtLicenseHold{}).Error; err != nil {
			return err
		}
		if err := tx.Where("license_id = ?", id).Delete(&GptRtLicenseDevice{}).Error; err != nil {
			return err
		}
		return tx.Delete(&GptRtLicense{}, id).Error
	})
}

// SaveGptRtLicense 保存（激活/验证时更新）
func SaveGptRtLicense(lic *GptRtLicense) error {
	return DB.Save(lic).Error
}

// IsGptRtLicenseNotFound 是否未找到
func IsGptRtLicenseNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// GetGptRtLicenseDeviceByMachine 按许可证+机器指纹查询绑定
func GetGptRtLicenseDeviceByMachine(licenseID int, machineID string) (*GptRtLicenseDevice, error) {
	var dev GptRtLicenseDevice
	if err := DB.Where("license_id = ? AND machine_id = ?", licenseID, machineID).First(&dev).Error; err != nil {
		return nil, err
	}
	return &dev, nil
}

// CountGptRtLicenseDevices 统计已绑定设备数
func CountGptRtLicenseDevices(licenseID int) (int64, error) {
	var count int64
	err := DB.Model(&GptRtLicenseDevice{}).Where("license_id = ?", licenseID).Count(&count).Error
	return count, err
}

// CreateGptRtLicenseDevice 新增设备绑定
func CreateGptRtLicenseDevice(dev *GptRtLicenseDevice) error {
	return DB.Create(dev).Error
}

// UpdateGptRtLicenseDeviceMeta 更新设备最近登录 IP / UA
func UpdateGptRtLicenseDeviceMeta(licenseID int, machineID, ip, ua string) error {
	return DB.Model(&GptRtLicenseDevice{}).
		Where("license_id = ? AND machine_id = ?", licenseID, machineID).
		Updates(map[string]interface{}{
			"login_ip":   ip,
			"user_agent": ua,
		}).Error
}

// GetGptRtLicenseDevices 获取许可证下所有绑定设备
func GetGptRtLicenseDevices(licenseID int) ([]GptRtLicenseDevice, error) {
	var list []GptRtLicenseDevice
	err := DB.Where("license_id = ?", licenseID).Order("id DESC").Find(&list).Error
	return list, err
}

// ConsumeGptRtLicenseQuota 原子扣减可用次数并增加已用次数
func ConsumeGptRtLicenseQuota(id int) error {
	res := DB.Model(&GptRtLicense{}).
		Where("id = ? AND available_count > 0", id).
		Updates(map[string]interface{}{
			"available_count": gorm.Expr("available_count - 1"),
			"used_count":      gorm.Expr("used_count + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ReserveGptRtLicenseQuota 原子预占一次配额
func ReserveGptRtLicenseQuota(licenseID int, machineID, email string, ttl time.Duration) (holdID string, lic *GptRtLicense, err error) {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&GptRtLicense{}).
			Where("id = ? AND status = 1 AND available_count > 0 AND expires_at > ?", licenseID, now).
			Updates(map[string]interface{}{
				"available_count": gorm.Expr("available_count - 1"),
				"held_count":      gorm.Expr("held_count + 1"),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		id, genErr := generateGptRtHoldID()
		if genErr != nil {
			return genErr
		}
		hold := &GptRtLicenseHold{
			HoldId:    id,
			LicenseId: licenseID,
			MachineId: machineID,
			Email:     email,
			Status:    GptRtHoldStatusHeld,
			ExpiresAt: now.Add(ttl),
		}
		if err := tx.Create(hold).Error; err != nil {
			return err
		}
		holdID = id

		var updated GptRtLicense
		if err := tx.First(&updated, licenseID).Error; err != nil {
			return err
		}
		lic = &updated
		return nil
	})
	return holdID, lic, err
}

// ConfirmGptRtLicenseHold 确认预占：held→confirmed，held-1 used+1
func ConfirmGptRtLicenseHold(holdID string) (lic *GptRtLicense, err error) {
	disabled := false
	err = DB.Transaction(func(tx *gorm.DB) error {
		var hold GptRtLicenseHold
		if err := tx.Where("hold_id = ? AND status = ?", holdID, GptRtHoldStatusHeld).First(&hold).Error; err != nil {
			return err
		}

		var license GptRtLicense
		if err := tx.First(&license, hold.LicenseId).Error; err != nil {
			return err
		}

		if license.Status != 1 {
			// 禁用时改为释放预占；事务内必须返回 nil 才能提交，业务错误在外层返回
			res := tx.Model(&GptRtLicenseHold{}).
				Where("hold_id = ? AND status = ?", holdID, GptRtHoldStatusHeld).
				Update("status", GptRtHoldStatusReleased)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			res = tx.Model(&GptRtLicense{}).
				Where("id = ? AND held_count > 0", hold.LicenseId).
				Updates(map[string]interface{}{
					"held_count":      gorm.Expr("held_count - 1"),
					"available_count": gorm.Expr("available_count + 1"),
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			if err := tx.First(&license, hold.LicenseId).Error; err != nil {
				return err
			}
			lic = &license
			disabled = true
			return nil
		}

		res := tx.Model(&GptRtLicenseHold{}).
			Where("hold_id = ? AND status = ?", holdID, GptRtHoldStatusHeld).
			Update("status", GptRtHoldStatusConfirmed)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		res = tx.Model(&GptRtLicense{}).
			Where("id = ? AND held_count > 0", hold.LicenseId).
			Updates(map[string]interface{}{
				"held_count": gorm.Expr("held_count - 1"),
				"used_count": gorm.Expr("used_count + 1"),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.First(&license, hold.LicenseId).Error; err != nil {
			return err
		}
		lic = &license
		return nil
	})
	if err != nil {
		return lic, err
	}
	if disabled {
		return lic, ErrGptRtLicenseDisabled
	}
	return lic, nil
}

// ReleaseGptRtLicenseHold 释放预占：held→released，held-1 available+1
func ReleaseGptRtLicenseHold(holdID string) (lic *GptRtLicense, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		var hold GptRtLicenseHold
		if err := tx.Where("hold_id = ? AND status = ?", holdID, GptRtHoldStatusHeld).First(&hold).Error; err != nil {
			return err
		}

		res := tx.Model(&GptRtLicenseHold{}).
			Where("hold_id = ? AND status = ?", holdID, GptRtHoldStatusHeld).
			Update("status", GptRtHoldStatusReleased)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		res = tx.Model(&GptRtLicense{}).
			Where("id = ? AND held_count > 0", hold.LicenseId).
			Updates(map[string]interface{}{
				"held_count":      gorm.Expr("held_count - 1"),
				"available_count": gorm.Expr("available_count + 1"),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var license GptRtLicense
		if err := tx.First(&license, hold.LicenseId).Error; err != nil {
			return err
		}
		lic = &license
		return nil
	})
	return lic, err
}

// ExpireGptRtLicenseHolds 过期预占并恢复可用次数
func ExpireGptRtLicenseHolds(limit int) (n int, err error) {
	if limit <= 0 {
		limit = 50
	}
	now := time.Now()
	var holds []GptRtLicenseHold
	if err = DB.Where("status = ? AND expires_at < ?", GptRtHoldStatusHeld, now).
		Order("expires_at ASC").Limit(limit).Find(&holds).Error; err != nil {
		return 0, err
	}
	for _, hold := range holds {
		e := DB.Transaction(func(tx *gorm.DB) error {
			res := tx.Model(&GptRtLicenseHold{}).
				Where("hold_id = ? AND status = ?", hold.HoldId, GptRtHoldStatusHeld).
				Update("status", GptRtHoldStatusExpired)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}
			res = tx.Model(&GptRtLicense{}).
				Where("id = ? AND held_count > 0", hold.LicenseId).
				Updates(map[string]interface{}{
					"held_count":      gorm.Expr("held_count - 1"),
					"available_count": gorm.Expr("available_count + 1"),
				})
			if res.Error != nil {
				return res.Error
			}
			n++
			return nil
		})
		if e != nil {
			return n, e
		}
	}
	return n, nil
}

// RevokeGptRtLicenseHolds 撤销许可证下所有 held 预占并恢复可用次数
func RevokeGptRtLicenseHolds(licenseID int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var holds []GptRtLicenseHold
		if err := tx.Where("license_id = ? AND status = ?", licenseID, GptRtHoldStatusHeld).
			Find(&holds).Error; err != nil {
			return err
		}
		if len(holds) == 0 {
			return nil
		}
		res := tx.Model(&GptRtLicenseHold{}).
			Where("license_id = ? AND status = ?", licenseID, GptRtHoldStatusHeld).
			Update("status", GptRtHoldStatusRevoked)
		if res.Error != nil {
			return res.Error
		}
		n := int(res.RowsAffected)
		if n == 0 {
			return nil
		}
		return tx.Model(&GptRtLicense{}).
			Where("id = ?", licenseID).
			Updates(map[string]interface{}{
				"held_count":      gorm.Expr("GREATEST(held_count - ?, 0)", n),
				"available_count": gorm.Expr("available_count + ?", n),
			}).Error
	})
}

// MigrateGptRtLicenseLegacyMachineID 将主表旧 machine_id 迁移到设备绑定子表
func MigrateGptRtLicenseLegacyMachineID() error {
	type legacyRow struct {
		Id          int
		MachineId   string
		ActivatedAt *time.Time
	}
	var rows []legacyRow
	err := DB.Raw(`SELECT id, machine_id, activated_at FROM gpt_rt_licenses
		WHERE machine_id IS NOT NULL AND machine_id != ''`).Scan(&rows).Error
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "unknown column") || strings.Contains(msg, "no such column") {
			return nil
		}
		return err
	}
	for _, r := range rows {
		var cnt int64
		if err := DB.Model(&GptRtLicenseDevice{}).
			Where("license_id = ? AND machine_id = ?", r.Id, r.MachineId).
			Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			continue
		}
		boundAt := time.Now()
		if r.ActivatedAt != nil {
			boundAt = *r.ActivatedAt
		}
		if err := DB.Create(&GptRtLicenseDevice{
			LicenseId: r.Id,
			MachineId: r.MachineId,
			BoundAt:   boundAt,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
