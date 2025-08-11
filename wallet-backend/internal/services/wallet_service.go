package services

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum/crypto"
)

// WalletService 钱包服务
type WalletService struct{}

// NewWalletService 创建新的钱包服务实例
func NewWalletService() *WalletService {
	return &WalletService{}
}

// GenerateEthereumAddress 生成以太坊地址
func (ws *WalletService) GenerateEthereumAddress(index uint32) (*models.AddressLibrary, error) {
	// 这里应该从配置中获取主密钥和HD路径
	// 暂时使用随机生成的方式
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public()
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

// GenerateBitcoinAddress 生成比特币地址
func (ws *WalletService) GenerateBitcoinAddress(index uint32) (*models.AddressLibrary, error) {
	// 这里应该从配置中获取主密钥和HD路径
	// 暂时使用随机生成的方式
	// privateKey, err := btcec.NewPrivateKey()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to generate private key: %v", err)
	// }

	// publicKey := privateKey.PubKey()
	// address, err := publicKey.Address(&chaincfg.MainNetParams)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to generate address: %v", err)
	// }
	
	// 暂时使用模拟地址
	address := fmt.Sprintf("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa%d", index)

	addressModel := &models.AddressLibrary{
		Address:   address,
		ChainType: "Bitcoin",
		Status:    0, // 未使用
		IndexNum:  uint64(index),
	}

	return addressModel, nil
}

// ValidateAddress 验证地址格式
func (ws *WalletService) ValidateAddress(address, chainType string) bool {
	switch chainType {
	case "Ethereum":
		return ws.validateEthereumAddress(address)
	case "Bitcoin":
		return ws.validateBitcoinAddress(address)
	default:
		return false
	}
}

// validateEthereumAddress 验证以太坊地址
func (ws *WalletService) validateEthereumAddress(address string) bool {
	if len(address) != 42 || address[:2] != "0x" {
		return false
	}

	_, err := hex.DecodeString(address[2:])
	return err == nil
}

// validateBitcoinAddress 验证比特币地址
func (ws *WalletService) validateBitcoinAddress(address string) bool {
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

// GenerateRandomHex 生成随机十六进制字符串
func (ws *WalletService) GenerateRandomHex(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
