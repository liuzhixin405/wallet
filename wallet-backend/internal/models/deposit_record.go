package models

import (
	"time"

	"gorm.io/gorm"
)

type DepositRecord struct {
	ID             uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         uint64         `json:"userid" gorm:"not null"`
	CurrencySymbol string         `json:"currency_symbol" gorm:"type:varchar(30);not null"`
	ChainType      string         `json:"chain_type" gorm:"type:varchar(30);not null"`
	Protocol       *string        `json:"protocol" gorm:"type:varchar(30)"`
	FromAddress    string         `json:"from_address" gorm:"type:varchar(100);not null"`
	ToAddress      string         `json:"to_address" gorm:"type:varchar(100);not null"`
	Amount         float64        `json:"amount" gorm:"type:decimal(36,18);not null"`
	Fee            float64        `json:"fee" gorm:"type:decimal(36,18);not null;default:0"`
	TxID           string         `json:"txid" gorm:"type:varchar(191);not null;uniqueIndex"`
	UniqueID       string         `json:"unique_id" gorm:"type:varchar(64);not null;uniqueIndex"`
	Status         bool           `json:"status" gorm:"not null"` // 0:充值确认中 1:完成
	IsInternal     bool           `json:"is_internal" gorm:"not null;default:false;index"`
	Confirmations  int            `json:"confirmations" gorm:"not null;default:0"`
	BlockHeight    *uint64        `json:"block_height"`
	NotifyStatus   bool           `json:"notify_status" gorm:"not null;default:false"`
	FailReason     string         `json:"fail_reason" gorm:"type:varchar(100);default:''"`
	ConfirmedTime  *time.Time     `json:"confirmed_time"`
	CreatedTime    time.Time      `json:"created_time" gorm:"not null;autoCreateTime;index"`
	UpdatedTime    time.Time      `json:"updated_time" gorm:"not null;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (DepositRecord) TableName() string {
	return "deposit_record"
}
