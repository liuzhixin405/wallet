package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum/crypto"
)

// Wallet 钱包接口
type Wallet interface {
	GenerateAddress(index uint32) (*models.AddressLibrary, error)
	GetBalance(address string) (float64, error)
	ValidateAddress(address string) bool
}

// EthereumWallet 以太坊钱包
type EthereumWallet struct {
	privateKey *ecdsa.PrivateKey
}

// NewEthereumWallet 创建新的以太坊钱包
func NewEthereumWallet() (*EthereumWallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	return &EthereumWallet{
		privateKey: privateKey,
	}, nil
}

// GenerateAddress 生成以太坊地址
func (ew *EthereumWallet) GenerateAddress(index uint32) (*models.AddressLibrary, error) {
	publicKey := ew.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	addressModel := &models.AddressLibrary{
		Address:   address.Hex(),
		ChainType: "Ethereum",
		Status:    0, // 未使用
		IndexNum:  uint64(index),
	}

	return addressModel, nil
}

// GetBalance 获取地址余额（这里需要连接到以太坊节点）
func (ew *EthereumWallet) GetBalance(address string) (float64, error) {
	// 这里应该连接到以太坊节点获取余额
	// 暂时返回0
	return 0, nil
}

// ValidateAddress 验证以太坊地址
func (ew *EthereumWallet) ValidateAddress(address string) bool {
	if len(address) != 42 || address[:2] != "0x" {
		return false
	}

	// 检查地址格式
	_, err := crypto.HexToAddress(address)
	return err == nil
}

// GetPrivateKey 获取私钥（仅用于测试）
func (ew *EthereumWallet) GetPrivateKey() *ecdsa.PrivateKey {
	return ew.privateKey
}
