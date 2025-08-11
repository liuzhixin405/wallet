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
	config  *config.Config
	clients map[string]*ethclient.Client // 支持多链
	stop    chan struct{}
}

// NewCollectionService 创建新的归集服务
func NewCollectionService(cfg *config.Config) (*CollectionService, error) {
	clients := make(map[string]*ethclient.Client)
	
	// 初始化以太坊客户端
	ethClient, err := ethclient.Dial(cfg.Ethereum.GetTestnetRPCURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum testnet: %v", err)
	}
	clients["ETH"] = ethClient
	
	// 初始化BSC客户端（如果有配置）
	if cfg.BSC != nil && cfg.BSC.RPCURL != "" {
		bscClient, err := ethclient.Dial(cfg.BSC.RPCURL)
		if err != nil {
			log.Printf("Warning: failed to connect to BSC: %v", err)
		} else {
			clients["BSC"] = bscClient
		}
	}

	return &CollectionService{
		config:  cfg,
		clients: clients,
		stop:    make(chan struct{}, 1),
	}, nil
}

// StartCollection 开始归集监控
func (cs *CollectionService) StartCollection() error {
	// 解析归集间隔
	dur, err := time.ParseDuration(cs.config.Wallet.HotWallet.CollectionThreshold)
	if err != nil {
		// 如果解析失败，使用默认值
		dur = 5 * time.Minute
		log.Printf("Invalid CollectionThreshold, using default: %v", dur)
	}
	
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for {
		select {
		case <-cs.stop:
			log.Println("Collection service stopped")
			return nil
		case <-ticker.C:
			if err := cs.checkAndCollect(); err != nil {
				log.Printf("Collection error: %v", err)
			}
		}
	}
}

// Stop 停止归集
func (cs *CollectionService) Stop() {
	select {
	case cs.stop <- struct{}{}:
	default:
	}
}

// TriggerOnce 手动触发一次归集检查
func (cs *CollectionService) TriggerOnce() error {
	return cs.checkAndCollect()
}

// checkAndCollect 检查并执行归集
func (cs *CollectionService) checkAndCollect() error {
	// 获取所有启用的币种配置
	currencies, err := cs.getEnabledCurrencies()
	if err != nil {
		return fmt.Errorf("failed to get enabled currencies: %v", err)
	}

	for _, currency := range currencies {
		if err := cs.processCurrencyCollection(currency); err != nil {
			log.Printf("Failed to process collection for currency %s: %v", currency.Symbol, err)
			continue
		}
	}

	return nil
}

// processCurrencyCollection 处理指定币种的归集
func (cs *CollectionService) processCurrencyCollection(currency *models.CurrencyChainConfig) error {
	// 获取该币种的热钱包地址
	hotWallets, err := cs.getHotWalletsForSymbol(currency.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get hot wallets for symbol %s: %v", currency.Symbol, err)
	}

	for _, wallet := range hotWallets {
		if err := cs.processWallet(wallet, currency); err != nil {
			log.Printf("Failed to process wallet %s for symbol %s: %v", wallet.Address, currency.Symbol, err)
			continue
		}
	}

	return nil
}

// getHotWalletsForSymbol 获取指定币种的热钱包地址
func (cs *CollectionService) getHotWalletsForSymbol(symbol string) ([]models.AddressLibrary, error) {
	var addresses []models.AddressLibrary
	chainType := cs.getChainTypeForSymbol(symbol)
	result := database.DB.Where("chain_type = ? AND wallet_type = ?", chainType, "hot").Find(&addresses)
	if result.Error != nil {
		return nil, result.Error
	}
	return addresses, nil
}

// processWallet 处理单个钱包
func (cs *CollectionService) processWallet(wallet models.AddressLibrary, currency *models.CurrencyChainConfig) error {
	// 获取对应链的客户端
	client, err := cs.getClientForSymbol(currency.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get client for symbol %s: %v", currency.Symbol, err)
	}

	// 获取余额
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(wallet.Address), nil)
	if err != nil {
		return fmt.Errorf("failed to get balance: %v", err)
	}

	// 检查是否需要归集
	if !cs.shouldCollect(balance, currency) {
		return nil
	}

	// 归集资金
	if err := cs.CollectFunds(currency.Symbol, wallet.Address, cs.config.Wallet.ColdWallet.Address, balance, client); err != nil {
		log.Printf("Failed to collect funds from %s for symbol %s: %v", wallet.Address, currency.Symbol, err)
		return err
	}

	return nil
}

// shouldCollect 检查是否应该归集
func (cs *CollectionService) shouldCollect(balance *big.Int, currency *models.CurrencyChainConfig) bool {
	// 获取归集阈值
	threshold := new(big.Int)
	threshold.SetString(cs.config.Wallet.HotWallet.CollectionThreshold, 10)

	// 如果余额超过阈值，则归集
	return balance.Cmp(threshold) > 0
}

// CollectFunds 归集资金到冷钱包
func (cs *CollectionService) CollectFunds(symbol string, fromAddress string, toAddress string, amount *big.Int, client *ethclient.Client) error {
	// 验证参数
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if fromAddress == "" {
		return fmt.Errorf("fromAddress is required")
	}
	if toAddress == "" {
		return fmt.Errorf("toAddress is required")
	}
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// 获取冷钱包地址
	if toAddress == "0x0000000000000000000000000000000000000000" {
		return fmt.Errorf("invalid target address")
	}

	// 获取发送方私钥
	privateKey, err := cs.getPrivateKeyByAddress(fromAddress, symbol)
	if err != nil {
		return fmt.Errorf("failed to get private key: %v", err)
	}

	// 获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(fromAddress))
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// 估算gas
	gasPrice, err := client.SuggestGasPrice(context.Background())
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

	toAddr := common.HexToAddress(toAddress)
	msg.To = &toAddr

	gasLimit, err := client.EstimateGas(context.Background(), msg)
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
		common.HexToAddress(toAddress),
		actualAmount,
		gasLimit,
		gasPrice,
		nil,
	)

	// 签名交易
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// 发送交易
	if err := client.SendTransaction(context.Background(), signedTx); err != nil {
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
		ChainType:      cs.getChainTypeForSymbol(symbol),
		CurrencySymbol: symbol,
		CreatedTime:    time.Now(),
		UpdatedTime:    time.Now(),
	}

	if err := database.DB.Create(chainBill).Error; err != nil {
		log.Printf("Failed to save collection transaction: %v", err)
	}

	log.Printf("Collection transaction sent: %s, symbol: %s, amount: %s", signedTx.Hash().Hex(), symbol, actualAmount.String())
	return nil
}

// CollectFromAddress 从指定地址归集资金
func (cs *CollectionService) CollectFromAddress(symbol string, address string) error {
	// 获取对应链的客户端
	client, err := cs.getClientForSymbol(symbol)
	if err != nil {
		return fmt.Errorf("failed to get client for symbol %s: %v", symbol, err)
	}

	// 获取余额
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return fmt.Errorf("failed to get balance: %v", err)
	}

	// 检查余额
	if balance.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("insufficient balance for collection")
	}

	// 执行归集
	if err := cs.CollectFunds(symbol, address, cs.config.Wallet.ColdWallet.Address, balance, client); err != nil {
		return fmt.Errorf("failed to collect funds: %v", err)
	}

	return nil
}

// getPrivateKeyByAddress 根据地址获取私钥
func (cs *CollectionService) getPrivateKeyByAddress(address string, symbol string) (*ecdsa.PrivateKey, error) {
	// 这里需要实现从地址索引到私钥的映射
	// 在实际应用中，这通常通过数据库查询或内存映射实现
	// 暂时返回错误，需要实现完整的地址管理
	return nil, fmt.Errorf("private key retrieval not implemented yet")
}

// getClientForSymbol 根据币种获取对应的客户端
func (cs *CollectionService) getClientForSymbol(symbol string) (*ethclient.Client, error) {
	// 根据币种确定使用哪个客户端
	switch symbol {
	case "ETH", "USDT", "USDC":
		if client, exists := cs.clients["ETH"]; exists {
			return client, nil
		}
	case "BNB", "BUSD":
		if client, exists := cs.clients["BSC"]; exists {
			return client, nil
		}
		// 如果没有BSC客户端，使用ETH客户端
		if client, exists := cs.clients["ETH"]; exists {
			return client, nil
		}
	default:
		// 默认使用ETH客户端
		if client, exists := cs.clients["ETH"]; exists {
			return client, nil
		}
	}
	
	return nil, fmt.Errorf("no client available for symbol %s", symbol)
}

// getChainTypeForSymbol 根据币种获取链类型
func (cs *CollectionService) getChainTypeForSymbol(symbol string) string {
	switch symbol {
	case "ETH", "USDT", "USDC":
		return "Ethereum"
	case "BNB", "BUSD":
		return "BSC"
	default:
		return "Ethereum"
	}
}

// getEnabledCurrencies 获取所有启用的币种配置
func (cs *CollectionService) getEnabledCurrencies() ([]*models.CurrencyChainConfig, error) {
	var currencies []*models.CurrencyChainConfig
	result := database.DB.Where("is_enabled = ?", true).Find(&currencies)
	if result.Error != nil {
		return nil, result.Error
	}
	return currencies, nil
}

// Close 关闭服务
func (cs *CollectionService) Close() {
	for _, client := range cs.clients {
		if client != nil {
			client.Close()
		}
	}
}
