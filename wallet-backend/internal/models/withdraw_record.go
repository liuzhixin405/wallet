package models

import (
	"time"

	"gorm.io/gorm"
)

type WithdrawRecord struct {
	ID             uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CurrencySymbol string         `json:"currency_symbol" gorm:"type:varchar(30);not null;index"`
	ChainType      string         `json:"chain_type" gorm:"type:varchar(30);not null;index"`
	Protocol       string         `json:"protocol" gorm:"type:varchar(30);default:''"`
	UserID         uint64         `json:"user_id" gorm:"not null"`
	FromAddress    string         `json:"from_address" gorm:"type:varchar(100);not null;index"`
	ToAddress      string         `json:"to_address" gorm:"type:varchar(100);not null;index"`
	PreSignData    *string        `json:"pre_sign_data" gorm:"type:text"`
	PostSignData   *string        `json:"post_sign_data" gorm:"type:text"`
	TxID           *string        `json:"txid" gorm:"type:varchar(191);uniqueIndex"`
	Amount         float64        `json:"amount" gorm:"type:decimal(36,18);not null;default:0"`
	Fee            float64        `json:"fee" gorm:"type:decimal(36,18);not null;default:0"`
	TotalAmount    float64        `json:"total_amount" gorm:"type:decimal(36,18);not null;default:0"`
	UniqueID       string         `json:"unique_id" gorm:"type:varchar(64);not null;uniqueIndex"`
	Status         int            `json:"status" gorm:"not null;default:0;index"` // 0-待转手续费,1-待签名,2-签名成功,3-发送成功,4-确认成功,10-待转手续失败,11-签名失败,12-发送失败
	BlockHeight    *uint64        `json:"block_height"`
	Confirmations  int            `json:"confirmations" gorm:"not null;default:0"`
	IsInternal     bool           `json:"is_internal" gorm:"not null;default:false"`
	NotifyStatus   bool           `json:"notify_status" gorm:"not null;default:false"`
	FailReason     string         `json:"fail_reason" gorm:"type:varchar(100);default:''"`
	Remark         *string        `json:"remark" gorm:"type:varchar(255)"`
	ConfirmedTime  *time.Time     `json:"confirmed_time"`
	CreatedAt      time.Time      `json:"created_at" gorm:"not null;autoCreateTime;index"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"not null;autoUpdateTime"`
	Type           *int           `json:"type"` // 1:提币 2:提币手续费 3:充值4:归集 5:管理员提币
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (WithdrawRecord) TableName() string {
	return "withdraw_record"
}
