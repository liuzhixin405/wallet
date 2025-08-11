package models

import (
	"time"

	"gorm.io/gorm"
)

type CurrencyChainConfig struct {
	ID             uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CurrencySymbol string         `json:"currency_symbol" gorm:"type:varchar(30);not null;uniqueIndex"`
	ChainType      string         `json:"chain_type" gorm:"type:varchar(30);not null;index"`
	Protocol       *string        `json:"protocol" gorm:"type:varchar(30)"`
	Decimals       int            `json:"decimals" gorm:"not null;default:18"`
	MinWithdraw    float64        `json:"min_withdraw" gorm:"type:decimal(36,18);not null;default:0"`
	MaxWithdraw    float64        `json:"max_withdraw" gorm:"type:decimal(36,18);not null;default:0"`
	WithdrawFee    float64        `json:"withdraw_fee" gorm:"type:decimal(36,18);not null;default:0"`
	DepositFee     float64        `json:"deposit_fee" gorm:"type:decimal(36,18);not null;default:0"`
	Status         bool           `json:"status" gorm:"not null;default:true"`
	CreatedTime    time.Time      `json:"created_time" gorm:"not null;autoCreateTime"`
	UpdatedTime    time.Time      `json:"updated_time" gorm:"not null;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (CurrencyChainConfig) TableName() string {
	return "currency_chain_config"
}
