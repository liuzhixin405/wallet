package models

import (
	"time"

	"gorm.io/gorm"
)

type Balance struct {
	ID             uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CurrencySymbol string         `json:"currency_symbol" gorm:"type:varchar(30);not null;index"`
	ChainType      string         `json:"chain_type" gorm:"type:varchar(30);not null;index"`
	Protocol       *string        `json:"protocol" gorm:"type:varchar(30)"`
	Address        string         `json:"address" gorm:"type:varchar(191);not null;index"`
	Balance        float64        `json:"balance" gorm:"type:decimal(36,18);not null;default:0"`
	Frozen         float64        `json:"frozen" gorm:"type:decimal(36,18);not null;default:0"`
	Total          float64        `json:"total" gorm:"-"` // 计算字段，不在数据库中存储
	CreatedTime    time.Time      `json:"created_time" gorm:"not null;autoCreateTime"`
	UpdatedTime    time.Time      `json:"updated_time" gorm:"not null;autoUpdateTime;index"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Balance) TableName() string {
	return "balance"
}

// GetTotal 计算总余额
func (b *Balance) GetTotal() float64 {
	return b.Balance + b.Frozen
}

// BeforeSave 保存前钩子，计算Total字段
func (b *Balance) BeforeSave(tx *gorm.DB) error {
	b.Total = b.GetTotal()
	return nil
}

// AfterFind 查询后钩子，计算Total字段
func (b *Balance) AfterFind(tx *gorm.DB) error {
	b.Total = b.GetTotal()
	return nil
}
