package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"
	"wallet-backend/internal/config"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CollectionService 归集服务
type CollectionService struct {
	config *config.Config
	client *ethclient.Client
}

// NewCollectionService 创建新的归集服务
func NewCollectionService(cfg *config.Config) (*CollectionService, error) {
	client, err := ethclient.Dial(cfg.Ethereum.GetTestnetRPCURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum testnet: %v", err)
	}

	return &CollectionService{
		config: cfg,
		client: client,
	}, nil
}

// StartCollection 开始归集监控
func (cs *CollectionService) StartCollection() error {
	ticker := time.NewTicker(time.Duration(cs.config.Scanner.ScanInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cs.checkAndCollect(); err != nil {
				log.Printf("Collection error: %v", err)
			}
		}
	}
}

// checkAndCollect 检查并执行归集
func (cs *CollectionService) checkAndCollect() error {
	// 获取所有热钱包地址
	hotWallets, err := cs.getHotWallets()
	if err != nil {
		return fmt.Errorf("failed to get hot wallets: %v", err)
	}

	for _, wallet := range hotWallets {
		if err := cs.processWallet(wallet); err != nil {
			log.Printf("Failed to process wallet %s: %v", wallet.Address, err)
			continue
		}
	}

	return nil
}

// getHotWallets 获取热钱包地址
func (cs *CollectionService) getHotWallets() ([]models.AddressLibrary, error) {
	var addresses []models.AddressLibrary
	result := database.DB.Where("chain_type = ? AND wallet_type = ?", "Ethereum", "hot").Find(&addresses)
	if result.Error != nil {
		return nil, result.Error
	}
	return addresses, nil
}

// processWallet 处理单个钱包
func (cs *CollectionService) processWallet(wallet models.AddressLibrary) error {
	// 获取余额
	balance, err := cs.client.BalanceAt(context.Background(), common.HexToAddress(wallet.Address), nil)
	if err != nil {
		return fmt.Errorf("failed to get balance: %v", err)
	}

	// 检查是否需要归集
	if !cs.shouldCollect(balance) {
		return nil
	}

	// 执行归集
	if err := cs.collectFunds(wallet.Address, balance); err != nil {
		return fmt.Errorf("failed to collect funds: %v", err)
	}

	return nil
}

// shouldCollect 检查是否应该归集
func (cs *CollectionService) shouldCollect(balance *big.Int) bool {
	// 获取归集阈值
	threshold := new(big.Int)
	threshold.SetString(cs.config.Wallet.HotWallet.CollectionThreshold, 10)

	// 如果余额超过阈值，则归集
	return balance.Cmp(threshold) > 0
}

// collectFunds 归集资金
func (cs *CollectionService) collectFunds(fromAddress string, amount *big.Int) error {
	// 获取冷钱包地址
	coldWalletAddress := cs.config.Wallet.ColdWallet.Address
	if coldWalletAddress == "0x0000000000000000000000000000000000000000" {
		return fmt.Errorf("cold wallet address not configured")
	}

	// 获取发送方私钥
	privateKey, err := cs.getPrivateKeyByAddress(fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get private key: %v", err)
	}

	// 获取nonce
	nonce, err := cs.client.PendingNonceAt(context.Background(), common.HexToAddress(fromAddress))
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// 估算gas
	gasPrice, err := cs.client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %v", err)
	}

	// 估算gas limit
	msg := ethereum.CallMsg{
		From:  common.HexToAddress(fromAddress),
		To:    &common.Address{},
		Value: amount,
		Data:  []byte{},
	}

	toAddr := common.HexToAddress(coldWalletAddress)
	msg.To = &toAddr

	gasLimit, err := cs.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to estimate gas: %v", err)
	}

	// 计算实际发送金额（减去gas费用）
	gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	actualAmount := new(big.Int).Sub(amount, gasCost)

	if actualAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("insufficient balance for gas fees")
	}

	// 创建交易
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(coldWalletAddress),
		actualAmount,
		gasLimit,
		gasPrice,
		nil,
	)

	// 签名交易
	chainID, err := cs.client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// 发送交易
	if err := cs.client.SendTransaction(context.Background(), signedTx); err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	// 转换金额为float64
	amountFloat, _ := new(big.Float).SetString(actualAmount.String())
	amountFloat64, _ := amountFloat.Float64()

	// 记录归集交易
	chainBill := &models.ChainBill{
		TxID:           signedTx.Hash().Hex(),
		Address:        fromAddress,
		Amount:         amountFloat64,
		Type:           3, // 归集
		Status:         0, // 待处理
		ChainType:      "Ethereum",
		CurrencySymbol: "ETH",
		CreatedTime:    time.Now(),
		UpdatedTime:    time.Now(),
	}

	if err := database.DB.Create(chainBill).Error; err != nil {
		log.Printf("Failed to save collection transaction: %v", err)
	}

	log.Printf("Collection transaction sent: %s, amount: %s ETH", signedTx.Hash().Hex(), actualAmount.String())
	return nil
}

// getPrivateKeyByAddress 根据地址获取私钥
func (cs *CollectionService) getPrivateKeyByAddress(address string) (*ecdsa.PrivateKey, error) {
	// 这里需要实现从地址索引到私钥的映射
	// 在实际应用中，这通常通过数据库查询或内存映射实现
	// 暂时返回错误，需要实现完整的地址管理
	return nil, fmt.Errorf("private key retrieval not implemented yet")
}

// Close 关闭服务
func (cs *CollectionService) Close() {
	if cs.client != nil {
		cs.client.Close()
	}
}
