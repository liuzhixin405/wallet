package services

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"wallet-backend/internal/config"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

// HDWalletService HD钱包服务
type HDWalletService struct {
	config *config.Config
	path   string
}

// NewHDWalletService 创建新的HD钱包服务
func NewHDWalletService(cfg *config.Config) *HDWalletService {
	return &HDWalletService{
		config: cfg,
		path:   cfg.Wallet.HDWallet.DerivationPath,
	}
}

// GenerateMnemonic 生成助记词
func (hws *HDWalletService) GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256) // 24个单词
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %v", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %v", err)
	}

	return mnemonic, nil
}

// ValidateMnemonic 验证助记词
func (hws *HDWalletService) ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// GenerateSeed 从助记词生成种子
func (hws *HDWalletService) GenerateSeed(mnemonic string) ([]byte, error) {
	if !hws.ValidateMnemonic(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, "")
	return seed, nil
}

// DeriveEthereumAddress 派生以太坊地址（简化版本）
func (hws *HDWalletService) DeriveEthereumAddress(mnemonic string, index uint32) (*models.AddressLibrary, error) {
	// 生成种子
	seed, err := hws.GenerateSeed(mnemonic)
	if err != nil {
		return nil, err
	}

	// 使用种子生成私钥（简化实现）
	privateKey, err := crypto.ToECDSA(seed[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// 生成公钥和地址
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

// DeriveBitcoinAddress 派生比特币地址（简化版本）
func (hws *HDWalletService) DeriveBitcoinAddress(mnemonic string, index uint32) (*models.AddressLibrary, error) {
	// 简化实现，生成一个模拟的比特币地址
	// seed, err := hws.GenerateSeed(mnemonic)
	// if err != nil {
	// 	return nil, err
	// }

	// 使用种子生成一个模拟地址
	address := fmt.Sprintf("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa%d", index)

	addressModel := &models.AddressLibrary{
		Address:   address,
		ChainType: "Bitcoin",
		Status:    0, // 未使用
		IndexNum:  uint64(index),
	}

	return addressModel, nil
}

// GetPrivateKey 获取私钥（用于签名交易）
func (hws *HDWalletService) GetPrivateKey(mnemonic string, chainType string, index uint32) (*ecdsa.PrivateKey, error) {
	seed, err := hws.GenerateSeed(mnemonic)
	if err != nil {
		return nil, err
	}

	// 使用种子生成私钥（简化实现）
	privateKey, err := crypto.ToECDSA(seed[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	return privateKey, nil
}

// ValidateAddress 验证地址格式
func (hws *HDWalletService) ValidateAddress(address, chainType string) bool {
	switch strings.ToLower(chainType) {
	case "ethereum":
		return hws.validateEthereumAddress(address)
	case "bitcoin":
		return hws.validateBitcoinAddress(address)
	default:
		return false
	}
}

// validateEthereumAddress 验证以太坊地址
func (hws *HDWalletService) validateEthereumAddress(address string) bool {
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}

	// 检查地址格式
	// if !keystore.ValidateAddressFormat(address) {
	// 	return false
	// }

	return true
}

// validateBitcoinAddress 验证比特币地址
func (hws *HDWalletService) validateBitcoinAddress(address string) bool {
	// 简化的比特币地址验证
	if len(address) < 26 || len(address) > 35 {
		return false
	}

	// 检查地址前缀
	if address[0] != '1' && address[0] != '3' && !strings.HasPrefix(address, "bc1") {
		return false
	}

	return true
}

// GetDerivationPath 获取派生路径
func (hws *HDWalletService) GetDerivationPath(chainType string) string {
	switch strings.ToLower(chainType) {
	case "ethereum":
		return "m/44'/60'/0'/0"
	case "bitcoin":
		return "m/44'/0'/0'/0"
	default:
		return hws.path
	}
}
