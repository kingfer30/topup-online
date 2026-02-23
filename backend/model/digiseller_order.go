package model

import (
	"time"

	"github.com/kingfer30/topup-online/utils/logger"
)

// DigisellerOrder Digiseller 订单记录
type DigisellerOrder struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Inv            int64      `gorm:"uniqueIndex:uq_inv;not null" json:"inv" comment:"Digiseller 发票号，全局唯一"`
	UniqueCode     string     `gorm:"size:20;not null;index" json:"unique_code" comment:"用户 16 位唯一码"`
	IdGoods        int64      `gorm:"default:null" json:"id_goods" comment:"商品 ID"`
	Amount         float64    `gorm:"type:decimal(10,2);default:null" json:"amount" comment:"实收金额"`
	TypeCurr       string     `gorm:"size:10;default:null" json:"type_curr" comment:"收款货币类型"`
	AmountUsd      float64    `gorm:"type:decimal(10,2);default:null" json:"amount_usd" comment:"等值 USD 金额"`
	Profit         string     `gorm:"size:50;default:null" json:"profit" comment:"扣佣后净收益"`
	DatePay        *time.Time `gorm:"default:null" json:"date_pay" comment:"支付时间"`
	Email          string     `gorm:"size:255;default:null" json:"email" comment:"买家邮箱"`
	AgentId        int        `gorm:"default:null" json:"agent_id" comment:"代理商 ID"`
	AgentPercent   float64    `gorm:"type:decimal(5,2);default:null" json:"agent_percent" comment:"代理商佣金比例"`
	CntGoods       int        `gorm:"default:null" json:"cnt_goods" comment:"购买数量"`
	PromoCode      string     `gorm:"size:100;default:null" json:"promo_code" comment:"买家使用的优惠码"`
	BonusCode      string     `gorm:"size:100;default:null" json:"bonus_code" comment:"赠送给买家的优惠码"`
	CartUid        string     `gorm:"size:100;default:null" json:"cart_uid" comment:"购物车 UID"`
	UcState        int8       `gorm:"default:null" json:"uc_state" comment:"唯一码状态：1未验证 2已交付待确认 3已确认 4已驳回 5已验证未交付"`
	UcDateCheck    *time.Time `gorm:"default:null" json:"uc_date_check" comment:"唯一码验证时间"`
	UcDateDelivery *time.Time `gorm:"default:null" json:"uc_date_delivery" comment:"商品交付时间"`
	UcDateConfirmed *time.Time `gorm:"default:null" json:"uc_date_confirmed" comment:"交付确认时间"`
	UcDateRefuted  *time.Time `gorm:"default:null" json:"uc_date_refuted" comment:"交付驳回时间"`
	OptionsJson    string     `gorm:"type:text;default:null" json:"options_json" comment:"额外参数列表（JSON）"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (DigisellerOrder) TableName() string {
	return "digiseller_orders"
}

// CreateOrUpdateDigisellerOrder 基于 inv 做 upsert，先插入；若 inv 唯一键冲突则更新所有字段
func CreateOrUpdateDigisellerOrder(order *DigisellerOrder) error {
	// 尝试插入，若 inv 已存在（唯一键冲突）则更新
	result := DB.Where(DigisellerOrder{Inv: order.Inv}).FirstOrInit(order)
	if result.Error != nil {
		logger.SysError("CreateOrUpdateDigisellerOrder FirstOrInit 失败: " + result.Error.Error())
		return result.Error
	}

	// 使用 Save 做完整更新（若已存在 ID 则 UPDATE，否则 INSERT）
	if err := DB.Save(order).Error; err != nil {
		logger.SysError("CreateOrUpdateDigisellerOrder Save 失败: " + err.Error())
		return err
	}
	return nil
}

// GetDigisellerOrderByInv 根据发票号查询订单
func GetDigisellerOrderByInv(inv int64) (*DigisellerOrder, error) {
	var order DigisellerOrder
	if err := DB.Where("inv = ?", inv).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetDigisellerOrderByUniqueCode 根据唯一码查询最新订单
func GetDigisellerOrderByUniqueCode(uniqueCode string) (*DigisellerOrder, error) {
	var order DigisellerOrder
	if err := DB.Where("unique_code = ?", uniqueCode).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
