package models

import (
	"time"

	"gorm.io/gorm"
)

type AddressLibrary struct {
	ID          uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      *uint64        `json:"userid" gorm:"index"`
	Address     string         `json:"address" gorm:"type:varchar(191);not null;uniqueIndex:idx_address_chain"`
	ChainType   string         `json:"chain_type" gorm:"type:varchar(30);not null;uniqueIndex:idx_address_chain;index"`
	Status      int            `json:"status" gorm:"default:0;index"` // 0-未使用, 1-已激活, 2-已冻结
	BindTime    *time.Time     `json:"bind_time" gorm:"index"`
	IndexNum    uint64         `json:"index_num" gorm:"default:0"`
	Note        string         `json:"note" gorm:"type:varchar(100);default:''"`
	CreatedTime time.Time      `json:"created_time" gorm:"not null;index"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AddressLibrary) TableName() string {
	return "address_library"
}
