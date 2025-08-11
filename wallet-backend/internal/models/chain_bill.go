package models

import (
	"time"

	"gorm.io/gorm"
)

type ChainBill struct {
	ID             uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         uint64         `json:"user_id" gorm:"not null;index"`
	CurrencySymbol string         `json:"currency_symbol" gorm:"type:varchar(30);not null;index"`
	ChainType      string         `json:"chain_type" gorm:"type:varchar(30);not null;index"`
	Protocol       *string        `json:"protocol" gorm:"type:varchar(30)"`
	Address        string         `json:"address" gorm:"type:varchar(191);not null;index"`
	TxID           string         `json:"txid" gorm:"type:varchar(191);not null;uniqueIndex"`
	Type           int            `json:"type" gorm:"not null;index"` // 1:充值 2:提币 3:归集 4:手续费
	Amount         float64        `json:"amount" gorm:"type:decimal(36,18);not null;default:0"`
	Fee            float64        `json:"fee" gorm:"type:decimal(36,18);not null;default:0"`
	Balance        float64        `json:"balance" gorm:"type:decimal(36,18);not null;default:0"`
	BlockHeight    *uint64        `json:"block_height"`
	Confirmations  int            `json:"confirmations" gorm:"not null;default:0"`
	Status         int            `json:"status" gorm:"not null;default:0;index"` // 0:确认中 1:已确认 2:失败
	Remark         *string        `json:"remark" gorm:"type:varchar(255)"`
	CreatedTime    time.Time      `json:"created_time" gorm:"not null;autoCreateTime;index"`
	UpdatedTime    time.Time      `json:"updated_time" gorm:"not null;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ChainBill) TableName() string {
	return "chain_bill"
}
