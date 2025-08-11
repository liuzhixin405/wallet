package blockchain

import (
	"fmt"
)

// BitcoinClient 比特币客户端
type BitcoinClient struct {
	isTestnet bool
}

// NewBitcoinClient 创建新的比特币客户端
func NewBitcoinClient(isTestnet bool) *BitcoinClient {
	return &BitcoinClient{isTestnet: isTestnet}
}

// GenerateAddress 生成比特币地址
func (bc *BitcoinClient) GenerateAddress() (string, error) {
	// 留空实现或调用服务层生成
	return "", fmt.Errorf("not implemented")
}

// ValidateAddress 验证比特币地址
func (bc *BitcoinClient) ValidateAddress(address string) bool {
	// 简单的比特币地址格式验证
	if len(address) < 26 || len(address) > 35 {
		return false
	}

	// 检查地址前缀
	if address[0] != '1' && address[0] != '3' && address[:3] != "bc1" {
		return false
	}

	return true
}

// GetNetworkName 获取网络名称
func (bc *BitcoinClient) GetNetworkName() string {
	if bc.isTestnet {
		return "Testnet"
	}
	return "Mainnet"
}
