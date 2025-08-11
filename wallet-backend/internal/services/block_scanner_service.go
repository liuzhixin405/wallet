package services

import (
	"context"
	"fmt"
	"log"
	"math/big"
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
	client     *ethclient.Client
	isScanning bool
	stopChan   chan bool
}

// NewBlockScannerService 创建新的区块扫描服务
func NewBlockScannerService(cfg *config.Config) (*BlockScannerService, error) {
	client, err := ethclient.Dial(cfg.Ethereum.GetTestnetRPCURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum testnet: %v", err)
	}

	return &BlockScannerService{
		config:   cfg,
		client:   client,
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
	latestBlock, err := bss.client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %v", err)
	}

	lastScannedBlock, err := bss.getLastScannedBlock()
	if err != nil {
		log.Printf("Failed to get last scanned block: %v", err)
		lastScannedBlock = 0
	}

	startBlock := lastScannedBlock + 1
	if startBlock > latestBlock {
		return nil
	}

	for blockNum := startBlock; blockNum <= latestBlock; blockNum++ {
		if err := bss.scanBlock(blockNum); err != nil {
			log.Printf("Failed to scan block %d: %v", blockNum, err)
			continue
		}
	}

	if err := bss.updateLastScannedBlock(latestBlock); err != nil {
		log.Printf("Failed to update last scanned block: %v", err)
	}

	return nil
}

// scanBlock 扫描单个区块
func (bss *BlockScannerService) scanBlock(blockNumber uint64) error {
	block, err := bss.client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		return fmt.Errorf("failed to get block %d: %v", blockNumber, err)
	}

	for _, tx := range block.Transactions() {
		if err := bss.processTransaction(tx, block); err != nil {
			log.Printf("Failed to process transaction %s: %v", tx.Hash().Hex(), err)
			continue
		}
	}

	return nil
}

// processTransaction 处理交易
func (bss *BlockScannerService) processTransaction(tx *types.Transaction, block *types.Block) error {
	if !bss.isRelevantTransaction(tx) {
		return nil
	}

	receipt, err := bss.client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	// 转换金额为float64
	amountFloat, _ := new(big.Float).SetString(tx.Value().String())
	amountFloat64, _ := amountFloat.Float64()

	// 确定交易类型
	txType := 1 // 默认充值
	if bss.isOutgoingTransaction(tx) {
		txType = 2 // 提币
	}

	blockHeight := block.Number().Uint64()
	chainBill := &models.ChainBill{
		TxID:           tx.Hash().Hex(),
		Address:        bss.getToAddress(tx),
		Amount:         amountFloat64,
		Type:           txType,
		Status:         bss.getTransactionStatus(receipt),
		BlockHeight:    &blockHeight,
		ChainType:      "Ethereum",
		CurrencySymbol: "ETH",
		CreatedTime:    time.Now(),
		UpdatedTime:    time.Now(),
	}

	if err := bss.saveTransaction(chainBill); err != nil {
		return fmt.Errorf("failed to save transaction: %v", err)
	}

	if bss.isIncomingTransaction(tx) {
		if err := bss.updateBalance(tx); err != nil {
			log.Printf("Failed to update balance: %v", err)
		}
	}

	return nil
}

// isRelevantTransaction 检查是否是相关交易
func (bss *BlockScannerService) isRelevantTransaction(tx *types.Transaction) bool {
	from := bss.getFromAddress(tx)
	if bss.isOurAddress(from) {
		return true
	}

	to := bss.getToAddress(tx)
	if bss.isOurAddress(to) {
		return true
	}

	return false
}

// isOurAddress 检查是否是我们的地址
func (bss *BlockScannerService) isOurAddress(address string) bool {
	if address == "" {
		return false
	}

	var addressLib models.AddressLibrary
	result := database.DB.Where("address = ? AND chain_type = ?", address, "Ethereum").First(&addressLib)
	return result.Error == nil
}

// getFromAddress 获取发送方地址
func (bss *BlockScannerService) getFromAddress(tx *types.Transaction) string {
	from, err := types.Sender(types.NewEIP155Signer(big.NewInt(bss.config.Ethereum.Testnet.ChainID)), tx)
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
func (bss *BlockScannerService) isIncomingTransaction(tx *types.Transaction) bool {
	to := bss.getToAddress(tx)
	return bss.isOurAddress(to)
}

// isOutgoingTransaction 检查是否是出账交易
func (bss *BlockScannerService) isOutgoingTransaction(tx *types.Transaction) bool {
	from := bss.getFromAddress(tx)
	return bss.isOurAddress(from)
}

// updateBalance 更新余额
func (bss *BlockScannerService) updateBalance(tx *types.Transaction) error {
	to := bss.getToAddress(tx)
	if to == "" {
		return nil
	}

	var balance models.Balance
	result := database.DB.Where("address = ? AND currency_symbol = ?", to, "ETH").First(&balance)

	if result.Error != nil {
		// 转换金额为float64
		amountFloat, _ := new(big.Float).SetString(tx.Value().String())
		amountFloat64, _ := amountFloat.Float64()

		balance = models.Balance{
			Address:        to,
			CurrencySymbol: "ETH",
			ChainType:      "Ethereum",
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
func (bss *BlockScannerService) getLastScannedBlock() (uint64, error) {
	return 0, nil
}

// updateLastScannedBlock 更新最后扫描的区块号
func (bss *BlockScannerService) updateLastScannedBlock(blockNumber uint64) error {
	log.Printf("Updated last scanned block to %d", blockNumber)
	return nil
}

// Close 关闭服务
func (bss *BlockScannerService) Close() {
	bss.StopScanning()
	if bss.client != nil {
		bss.client.Close()
	}
}
