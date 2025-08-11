package services

import (
	"wallet-backend/internal/config"
)

// Config 服务配置结构体，用于传递所有服务实例
type Config struct {
	AppConfig          *config.Config
	HDWalletService    *HDWalletService
	WSService          *WebSocketService
	BlockScannerService *BlockScannerService
	CollectionService  *CollectionService
} 