package services

import (
	"strings"
	"testing"
	"wallet-backend/internal/config"
)

func TestHDWalletService_GenerateMnemonic(t *testing.T) {
	cfg := &config.Config{}
	service := NewHDWalletService(cfg)

	mnemonic, err := service.GenerateMnemonic()
	if err != nil {
		t.Fatalf("Failed to generate mnemonic: %v", err)
	}

	if mnemonic == "" {
		t.Error("Generated mnemonic is empty")
	}

	// 验证助记词格式
	words := strings.Fields(mnemonic)
	if len(words) != 24 {
		t.Errorf("Expected 24 words, got %d", len(words))
	}
}

func TestHDWalletService_ValidateMnemonic(t *testing.T) {
	cfg := &config.Config{}
	service := NewHDWalletService(cfg)

	// 测试有效的助记词
	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	if !service.ValidateMnemonic(validMnemonic) {
		t.Error("Valid mnemonic should be accepted")
	}

	// 测试无效的助记词
	invalidMnemonic := "invalid mnemonic phrase"
	if service.ValidateMnemonic(invalidMnemonic) {
		t.Error("Invalid mnemonic should be rejected")
	}
}

func TestHDWalletService_GenerateSeed(t *testing.T) {
	cfg := &config.Config{}
	service := NewHDWalletService(cfg)

	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed, err := service.GenerateSeed(mnemonic)
	if err != nil {
		t.Fatalf("Failed to generate seed: %v", err)
	}

	if len(seed) == 0 {
		t.Error("Generated seed is empty")
	}

	// 种子应该是64字节
	if len(seed) != 64 {
		t.Errorf("Expected 64 bytes, got %d", len(seed))
	}
}

func TestHDWalletService_ValidateAddress(t *testing.T) {
	cfg := &config.Config{}
	service := NewHDWalletService(cfg)

	// 测试以太坊地址
	validEthAddress := "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
	if !service.ValidateAddress(validEthAddress, "Ethereum") {
		t.Error("Valid Ethereum address should be accepted")
	}

	invalidEthAddress := "0xinvalid"
	if service.ValidateAddress(invalidEthAddress, "Ethereum") {
		t.Error("Invalid Ethereum address should be rejected")
	}

	// 测试比特币地址
	validBtcAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	if !service.ValidateAddress(validBtcAddress, "Bitcoin") {
		t.Error("Valid Bitcoin address should be accepted")
	}

	invalidBtcAddress := "invalid"
	if service.ValidateAddress(invalidBtcAddress, "Bitcoin") {
		t.Error("Invalid Bitcoin address should be rejected")
	}
}

func TestHDWalletService_GetDerivationPath(t *testing.T) {
	cfg := &config.Config{}
	service := NewHDWalletService(cfg)

	// 测试以太坊派生路径
	ethPath := service.GetDerivationPath("Ethereum")
	expectedEthPath := "m/44'/60'/0'/0"
	if ethPath != expectedEthPath {
		t.Errorf("Expected %s, got %s", expectedEthPath, ethPath)
	}

	// 测试比特币派生路径
	btcPath := service.GetDerivationPath("Bitcoin")
	expectedBtcPath := "m/44'/0'/0'/0"
	if btcPath != expectedBtcPath {
		t.Errorf("Expected %s, got %s", expectedBtcPath, btcPath)
	}
}
