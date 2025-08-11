package models

import (
	"time"
)

// CurrencyChainConfig 币种链配置
type CurrencyChainConfig struct {
	ID                uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Symbol            string    `json:"symbol" gorm:"type:varchar(50);uniqueIndex;not null"`           // 币种符号
	ChainType         string    `json:"chain_type" gorm:"type:varchar(50);not null"`                   // 链类型
	IsEnabled         bool      `json:"is_enabled" gorm:"default:true"`               // 是否启用
	LastScannedBlock  *uint64   `json:"last_scanned_block"`                           // 最后扫描的区块号
	RPCURL            string    `json:"rpc_url" gorm:"type:varchar(500);not null"`                      // RPC地址
	ChainID           int64     `json:"chain_id" gorm:"not null"`                     // 链ID
	Confirmations     int       `json:"confirmations" gorm:"default:12"`              // 确认数
	TokenAddress      *string   `json:"token_address" gorm:"type:varchar(100)"`                                // 代币合约地址（如果是代币）
	Decimals          int       `json:"decimals" gorm:"default:18"`                   // 小数位数
	CollectionEnabled bool      `json:"collection_enabled" gorm:"default:true"`       // 是否启用归集
	CollectionThreshold string  `json:"collection_threshold" gorm:"type:varchar(50);default:'0.1'"`    // 归集阈值
	CreatedTime       time.Time `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime       time.Time `json:"updated_time" gorm:"autoUpdateTime"`
}
