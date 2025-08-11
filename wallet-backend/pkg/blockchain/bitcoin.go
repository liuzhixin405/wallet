package blockchain

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
)

// BitcoinClient 比特币客户端
type BitcoinClient struct {
	network *chaincfg.Params
}

// NewBitcoinClient 创建新的比特币客户端
func NewBitcoinClient(isTestnet bool) *BitcoinClient {
	var network *chaincfg.Params
	if isTestnet {
		network = &chaincfg.TestNet3Params
	} else {
		network = &chaincfg.MainNetParams
	}

	return &BitcoinClient{
		network: network,
	}
}

// GenerateAddress 生成比特币地址
func (bc *BitcoinClient) GenerateAddress() (string, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %v", err)
	}

	publicKey := privateKey.PubKey()
	address, err := publicKey.Address(bc.network)
	if err != nil {
		return "", fmt.Errorf("failed to generate address: %v", err)
	}

	return address.String(), nil
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
	if bc.network == &chaincfg.TestNet3Params {
		return "Testnet"
	}
	return "Mainnet"
}
