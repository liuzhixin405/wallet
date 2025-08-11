package services

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
	"wallet-backend/internal/config"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BlockScannerService 区块扫描服务
type BlockScannerService struct {
	config     *config.Config
	clients    map[string]*ethclient.Client // 支持多链
	isScanning bool
	stopChan   chan bool
}

// NewBlockScannerService 创建新的区块扫描服务
func NewBlockScannerService(cfg *config.Config) (*BlockScannerService, error) {
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

	return &BlockScannerService{
		config:   cfg,
		clients:  clients,
		stopChan: make(chan bool),
	}, nil
}

// StartScanning 开始扫描区块
func (bss *BlockScannerService) StartScanning() error {
	if bss.isScanning {
		return fmt.Errorf("block scanner is already running")
	}

	bss.isScanning = true
	go bss.scanBlocks()
	return nil
}

// StopScanning 停止扫描区块
func (bss *BlockScannerService) StopScanning() {
	if bss.isScanning {
		bss.isScanning = false
		bss.stopChan <- true
	}
}

// Status 返回扫描状态
func (bss *BlockScannerService) Status() bool {
	return bss.isScanning
}

// ScanOnce 手动扫描一次最新区块范围
func (bss *BlockScannerService) ScanOnce() error {
	return bss.scanLatestBlock()
}

// ScanBlocks 扫描指定区块范围
func (bss *BlockScannerService) ScanBlocks(symbol string, startBlock uint64, endBlock uint64, addresses []string) error {
	// 验证参数
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if startBlock >= endBlock {
		return fmt.Errorf("startBlock must be less than endBlock")
	}
	if len(addresses) == 0 {
		return fmt.Errorf("addresses list cannot be empty")
	}

	log.Printf("Starting block scan for symbol: %s, blocks: %d-%d, addresses: %d", 
		symbol, startBlock, endBlock, len(addresses))

	// 获取对应链的客户端
	client, err := bss.getClientForSymbol(symbol)
	if err != nil {
		return fmt.Errorf("failed to get client for symbol %s: %v", symbol, err)
	}

	// 获取最新区块号
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %v", err)
	}
	
	// 如果 endBlock 为 0，则扫描到最新区块
	if endBlock == 0 {
		endBlock = currentBlock
	}
	
	// 确保扫描范围不超过最新区块
	if endBlock > currentBlock {
		endBlock = currentBlock
	}

	log.Printf("Adjusted scan range: %d-%d", startBlock, endBlock)

	for blockNumber := startBlock; blockNumber <= endBlock; blockNumber++ {
		if err := bss.scanBlock(blockNumber, symbol, addresses, client); err != nil {
			log.Printf("Failed to scan block %d for symbol %s: %v", blockNumber, symbol, err)
			continue
		}
		
		// 每扫描10个区块输出一次进度
		if blockNumber%10 == 0 {
			log.Printf("Scan progress: %d/%d (%.1f%%)", 
				blockNumber-startBlock+1, endBlock-startBlock+1, 
				float64(blockNumber-startBlock+1)/float64(endBlock-startBlock+1)*100)
		}
	}

	log.Printf("Block scan completed for symbol: %s, blocks: %d-%d", symbol, startBlock, endBlock)
	return nil
}

// scanBlocks 扫描区块的主循环
func (bss *BlockScannerService) scanBlocks() {
	ticker := time.NewTicker(time.Duration(bss.config.Scanner.ScanInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bss.stopChan:
			log.Println("Block scanner stopped")
			return
		case <-ticker.C:
			if err := bss.scanLatestBlock(); err != nil {
				log.Printf("Error scanning block: %v", err)
			}
		}
	}
}

// scanLatestBlock 扫描最新区块
func (bss *BlockScannerService) scanLatestBlock() error {
	// 获取所有启用的币种配置
	currencies, err := bss.getEnabledCurrencies()
	if err != nil {
		return fmt.Errorf("failed to get enabled currencies: %v", err)
	}

	for _, currency := range currencies {
		if err := bss.scanCurrencyLatestBlock(currency); err != nil {
			log.Printf("Failed to scan latest block for currency %s: %v", currency.Symbol, err)
			continue
		}
	}

	return nil
}

// scanCurrencyLatestBlock 扫描指定币种的最新区块
func (bss *BlockScannerService) scanCurrencyLatestBlock(currency *models.CurrencyChainConfig) error {
	client, err := bss.getClientForSymbol(currency.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get client for symbol %s: %v", currency.Symbol, err)
	}

	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %v", err)
	}

	lastScannedBlock := currency.LastScannedBlock
	if lastScannedBlock == nil {
		lastScannedBlock = new(uint64)
		*lastScannedBlock = 0
	}

	startBlock := *lastScannedBlock + 1
	if startBlock > latestBlock {
		return nil
	}

	// 获取该币种的所有地址
	addresses, err := bss.getAddressesForSymbol(currency.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get addresses for symbol %s: %v", currency.Symbol, err)
	}

	for blockNum := startBlock; blockNum <= latestBlock; blockNum++ {
		if err := bss.scanBlock(blockNum, currency.Symbol, addresses, client); err != nil {
			log.Printf("Failed to scan block %d for symbol %s: %v", blockNum, currency.Symbol, err)
			continue
		}
	}

	// 更新最后扫描的区块号
	if err := bss.updateLastScannedBlock(currency.Symbol, latestBlock); err != nil {
		log.Printf("Failed to update last scanned block for symbol %s: %v", currency.Symbol, err)
	}

	return nil
}

// scanBlock 扫描单个区块
func (bss *BlockScannerService) scanBlock(blockNumber uint64, symbol string, addresses []string, client *ethclient.Client) error {
    // 获取区块信息
    block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
    if err != nil {
        return fmt.Errorf("failed to get block %d: %v", blockNumber, err)
    }

    // 检查区块是否包含我们关心的地址的交易
    hasRelevantTx := false
    
    // 遍历区块中的交易
    for _, tx := range block.Transactions() {
        // 检查交易是否涉及指定的地址
        relevant, err := bss.isRelevantTransaction(tx, addresses, client)
        if err != nil {
            log.Printf("Failed to check transaction relevance %s: %v", tx.Hash().Hex(), err)
            continue
        }
        
        if relevant {
            hasRelevantTx = true
            if err := bss.processTransaction(tx, blockNumber, symbol, client); err != nil {
                log.Printf("Failed to process transaction %s in block %d: %v", tx.Hash().Hex(), blockNumber, err)
            }
        }
    }
    
    if hasRelevantTx {
        log.Printf("Block %d contains relevant transactions for symbol %s", blockNumber, symbol)
    }

    return nil
}

// processTransaction 处理交易
func (bss *BlockScannerService) processTransaction(tx *types.Transaction, blockNumber uint64, symbol string, client *ethclient.Client) error {
	// 获取交易收据
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	// 转换金额为float64
	amountFloat, _ := new(big.Float).SetString(tx.Value().String())
	amountFloat64, _ := amountFloat.Float64()

	// 确定交易类型
	txType := 1 // 默认充值
	if bss.isOutgoingTransaction(tx, client) {
		txType = 2 // 提币
	}

	blockHeight := blockNumber
	chainBill := &models.ChainBill{
		TxID:           tx.Hash().Hex(),
		Address:        bss.getToAddress(tx),
		Amount:         amountFloat64,
		Type:           txType,
		Status:         bss.getTransactionStatus(receipt),
		BlockHeight:    &blockHeight,
		ChainType:      bss.getChainTypeForSymbol(symbol),
		CurrencySymbol: symbol,
		CreatedTime:    time.Now(),
		UpdatedTime:    time.Now(),
	}

	if err := bss.saveTransaction(chainBill); err != nil {
		return fmt.Errorf("failed to save transaction: %v", err)
	}

	if bss.isIncomingTransaction(tx, client) {
		if err := bss.updateBalance(tx, symbol); err != nil {
			log.Printf("Failed to update balance: %v", err)
		}
	}

	log.Printf("Processed transaction %s in block %d for symbol %s", tx.Hash().Hex(), blockNumber, symbol)
	return nil
}

// isRelevantTransaction 检查是否是相关交易
func (bss *BlockScannerService) isRelevantTransaction(tx *types.Transaction, addresses []string, client *ethclient.Client) (bool, error) {
    // 获取链ID
    chainID, err := client.ChainID(context.Background())
    if err != nil {
        return false, fmt.Errorf("failed to get chain ID: %v", err)
    }

    // 获取交易的发送方地址
    signer := types.NewEIP155Signer(chainID)
    from, err := types.Sender(signer, tx)
    if err != nil {
        return false, fmt.Errorf("failed to get sender address: %v", err)
    }

    // 获取交易的接收方地址
    to := tx.To()
    
    // 检查发送方地址
    for _, addr := range addresses {
        if strings.EqualFold(from.Hex(), addr) {
            return true, nil
        }
    }
    
    // 检查接收方地址
    if to != nil {
        for _, addr := range addresses {
            if strings.EqualFold(to.Hex(), addr) {
                return true, nil
            }
        }
    }
    
    return false, nil
}

// isOurAddress 检查是否是我们的地址
func (bss *BlockScannerService) isOurAddress(address string, symbol string) bool {
	if address == "" {
		return false
	}

	var addressLib models.AddressLibrary
	result := database.DB.Where("address = ? AND chain_type = ?", address, bss.getChainTypeForSymbol(symbol)).First(&addressLib)
	return result.Error == nil
}

// getFromAddress 获取发送方地址
func (bss *BlockScannerService) getFromAddress(tx *types.Transaction, client *ethclient.Client) string {
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return ""
	}
	from, err := types.Sender(types.NewEIP155Signer(chainID), tx)
	if err != nil {
		return ""
	}
	return from.Hex()
}

// getToAddress 获取接收方地址
func (bss *BlockScannerService) getToAddress(tx *types.Transaction) string {
	if tx.To() == nil {
		return ""
	}
	return tx.To().Hex()
}

// getTransactionStatus 获取交易状态
func (bss *BlockScannerService) getTransactionStatus(receipt *types.Receipt) int {
	if receipt.Status == 1 {
		return 1
	}
	return 2
}

// isIncomingTransaction 检查是否是入账交易
func (bss *BlockScannerService) isIncomingTransaction(tx *types.Transaction, client *ethclient.Client) bool {
	to := bss.getToAddress(tx)
	// 这里需要根据具体的币种来判断
	return bss.isOurAddress(to, "ETH") // 暂时硬编码，后续需要改进
}

// isOutgoingTransaction 检查是否是出账交易
func (bss *BlockScannerService) isOutgoingTransaction(tx *types.Transaction, client *ethclient.Client) bool {
	from := bss.getFromAddress(tx, client)
	// 这里需要根据具体的币种来判断
	return bss.isOurAddress(from, "ETH") // 暂时硬编码，后续需要改进
}

// updateBalance 更新余额
func (bss *BlockScannerService) updateBalance(tx *types.Transaction, symbol string) error {
	to := bss.getToAddress(tx)
	if to == "" {
		return nil
	}

	var balance models.Balance
	result := database.DB.Where("address = ? AND currency_symbol = ?", to, symbol).First(&balance)

	if result.Error != nil {
		// 转换金额为float64
		amountFloat, _ := new(big.Float).SetString(tx.Value().String())
		amountFloat64, _ := amountFloat.Float64()

		balance = models.Balance{
			Address:        to,
			CurrencySymbol: symbol,
			ChainType:      bss.getChainTypeForSymbol(symbol),
			Balance:        amountFloat64,
			CreatedTime:    time.Now(),
			UpdatedTime:    time.Now(),
		}
	} else {
		// 转换当前余额为big.Int
		currentBalance := new(big.Int)
		currentBalance.SetString(fmt.Sprintf("%.0f", balance.Balance), 10)
		newBalance := new(big.Int).Add(currentBalance, tx.Value())

		// 转换新余额为float64
		newBalanceFloat, _ := new(big.Float).SetString(newBalance.String())
		newBalanceFloat64, _ := newBalanceFloat.Float64()

		balance.Balance = newBalanceFloat64
		balance.UpdatedTime = time.Now()
	}

	if err := database.DB.Save(&balance).Error; err != nil {
		return fmt.Errorf("failed to save balance: %v", err)
	}

	return nil
}

// saveTransaction 保存交易记录
func (bss *BlockScannerService) saveTransaction(chainBill *models.ChainBill) error {
	var existing models.ChainBill
	result := database.DB.Where("tx_id = ?", chainBill.TxID).First(&existing)

	if result.Error != nil {
		return database.DB.Create(chainBill).Error
	} else {
		chainBill.ID = existing.ID
		return database.DB.Save(chainBill).Error
	}
}

// getLastScannedBlock 获取最后扫描的区块号
func (bss *BlockScannerService) getLastScannedBlock(symbol string) (uint64, error) {
	var currency models.CurrencyChainConfig
	result := database.DB.Where("symbol = ?", symbol).First(&currency)
	if result.Error != nil {
		return 0, result.Error
	}
	if currency.LastScannedBlock == nil {
		return 0, nil
	}
	return *currency.LastScannedBlock, nil
}

// updateLastScannedBlock 更新最后扫描的区块号
func (bss *BlockScannerService) updateLastScannedBlock(symbol string, blockNumber uint64) error {
	result := database.DB.Model(&models.CurrencyChainConfig{}).
		Where("symbol = ?", symbol).
		Update("last_scanned_block", blockNumber)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last scanned block: %v", result.Error)
	}
	
	log.Printf("Updated last scanned block for symbol %s to %d", symbol, blockNumber)
	return nil
}

// getClientForSymbol 根据币种获取对应的客户端
func (bss *BlockScannerService) getClientForSymbol(symbol string) (*ethclient.Client, error) {
	// 根据币种确定使用哪个客户端
	switch symbol {
	case "ETH", "USDT", "USDC":
		if client, exists := bss.clients["ETH"]; exists {
			return client, nil
		}
	case "BNB", "BUSD":
		if client, exists := bss.clients["BSC"]; exists {
			return client, nil
		}
		// 如果没有BSC客户端，使用ETH客户端
		if client, exists := bss.clients["ETH"]; exists {
			return client, nil
		}
	default:
		// 默认使用ETH客户端
		if client, exists := bss.clients["ETH"]; exists {
			return client, nil
		}
	}
	
	return nil, fmt.Errorf("no client available for symbol %s", symbol)
}

// getChainTypeForSymbol 根据币种获取链类型
func (bss *BlockScannerService) getChainTypeForSymbol(symbol string) string {
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
func (bss *BlockScannerService) getEnabledCurrencies() ([]*models.CurrencyChainConfig, error) {
	var currencies []*models.CurrencyChainConfig
	result := database.DB.Where("is_enabled = ?", true).Find(&currencies)
	if result.Error != nil {
		return nil, result.Error
	}
	return currencies, nil
}

// getAddressesForSymbol 获取指定币种的所有地址
func (bss *BlockScannerService) getAddressesForSymbol(symbol string) ([]string, error) {
	var addresses []models.AddressLibrary
	chainType := bss.getChainTypeForSymbol(symbol)
	
	result := database.DB.Where("chain_type = ?", chainType).Find(&addresses)
	if result.Error != nil {
		return nil, result.Error
	}
	
	var addressStrings []string
	for _, addr := range addresses {
		addressStrings = append(addressStrings, addr.Address)
	}
	
	return addressStrings, nil
}

// Close 关闭服务
func (bss *BlockScannerService) Close() {
	bss.StopScanning()
	for _, client := range bss.clients {
		if client != nil {
			client.Close()
		}
	}
}
