package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"time"
	"wallet-backend/internal/config"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TransactionService 交易服务
type TransactionService struct {
	config *config.Config
	client *ethclient.Client
}

// TransactionRequest 交易请求
type TransactionRequest struct {
	FromAddress string   `json:"from_address" binding:"required"`
	ToAddress   string   `json:"to_address" binding:"required"`
	Amount      string   `json:"amount" binding:"required"`
	GasPrice    *big.Int `json:"gas_price,omitempty"`
	GasLimit    uint64   `json:"gas_limit,omitempty"`
	Data        []byte   `json:"data,omitempty"`
}

// TransactionResponse 交易响应
type TransactionResponse struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
	GasPrice  string `json:"gas_price"`
	GasLimit  uint64 `json:"gas_limit"`
	Nonce     uint64 `json:"nonce"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// NewTransactionService 创建新的交易服务
func NewTransactionService(cfg *config.Config) (*TransactionService, error) {
	client, err := ethclient.Dial(cfg.Ethereum.GetTestnetRPCURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum: %v", err)
	}

	return &TransactionService{
		config: cfg,
		client: client,
	}, nil
}

// SendTransaction 发送交易
func (ts *TransactionService) SendTransaction(req *TransactionRequest, privateKey *ecdsa.PrivateKey) (*TransactionResponse, error) {
	// 验证地址
	if !common.IsHexAddress(req.FromAddress) || !common.IsHexAddress(req.ToAddress) {
		return nil, fmt.Errorf("invalid address format")
	}

	fromAddress := common.HexToAddress(req.FromAddress)
	toAddress := common.HexToAddress(req.ToAddress)

	// 解析金额
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format")
	}

	// 获取nonce
	nonce, err := ts.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// 获取gas价格
	var gasPrice *big.Int
	if req.GasPrice != nil {
		gasPrice = req.GasPrice
	} else {
		gasPrice, err = ts.client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %v", err)
		}
	}

	// 估算gas limit
	var gasLimit uint64
	if req.GasLimit > 0 {
		gasLimit = req.GasLimit
	} else {
		// 使用默认gas limit
		gasLimit = 21000
	}

	// 创建交易
	tx := types.NewTransaction(
		nonce,
		toAddress,
		amount,
		gasLimit,
		gasPrice,
		req.Data,
	)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(ts.config.Ethereum.Testnet.ChainID)), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// 广播交易
	err = ts.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	// 解析金额为float64
	amountFloat, _ := strconv.ParseFloat(req.Amount, 64)
	gasPriceFloat, _ := strconv.ParseFloat(gasPrice.String(), 64)

	// 保存交易记录
	chainBill := &models.ChainBill{
		UserID:         0, // 需要从上下文获取
		CurrencySymbol: "ETH",
		ChainType:      "Ethereum",
		Address:        req.FromAddress,
		TxID:           signedTx.Hash().Hex(),
		Type:           2, // 提币
		Amount:         amountFloat,
		Fee:            gasPriceFloat * float64(gasLimit),
		Balance:        0, // 需要计算
		Status:         0, // 待确认
		CreatedTime:    time.Now(),
	}

	if err := database.GetDB().Create(chainBill).Error; err != nil {
		return nil, fmt.Errorf("failed to save transaction: %v", err)
	}

	response := &TransactionResponse{
		Hash:      signedTx.Hash().Hex(),
		From:      req.FromAddress,
		To:        req.ToAddress,
		Amount:    req.Amount,
		GasPrice:  gasPrice.String(),
		GasLimit:  gasLimit,
		Nonce:     nonce,
		Status:    "pending",
		Timestamp: time.Now().Unix(),
	}

	return response, nil
}

// GetTransactionStatus 获取交易状态
func (ts *TransactionService) GetTransactionStatus(txHash string) (*TransactionResponse, error) {
	hash := common.HexToHash(txHash)

	// 获取交易收据
	receipt, err := ts.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	// 获取交易详情
	tx, _, err := ts.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}

	// 获取发送方地址
	from, err := types.Sender(types.NewEIP155Signer(big.NewInt(ts.config.Ethereum.Testnet.ChainID)), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %v", err)
	}

	// 确定状态
	status := "pending"
	if receipt.Status == types.ReceiptStatusSuccessful {
		status = "success"
	} else if receipt.Status == types.ReceiptStatusFailed {
		status = "failed"
	}

	response := &TransactionResponse{
		Hash:      txHash,
		From:      from.Hex(),
		To:        tx.To().Hex(),
		Amount:    tx.Value().String(),
		GasPrice:  tx.GasPrice().String(),
		GasLimit:  tx.Gas(),
		Nonce:     tx.Nonce(),
		Status:    status,
		Timestamp: time.Now().Unix(),
	}

	return response, nil
}

// EstimateGas 估算gas费用
func (ts *TransactionService) EstimateGas(req *TransactionRequest) (uint64, *big.Int, error) {
	// fromAddress := common.HexToAddress(req.FromAddress)
	// toAddress := common.HexToAddress(req.ToAddress)

	// amount, ok := new(big.Int).SetString(req.Amount, 10)
	// if !ok {
	// 	return 0, nil, fmt.Errorf("invalid amount format")
	// }

	// 使用默认gas limit
	gasLimit := uint64(21000)

	// 获取gas价格
	gasPrice, err := ts.client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	return gasLimit, gasPrice, nil
}

// GetBalance 获取余额
func (ts *TransactionService) GetBalance(address string) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid address format")
	}

	account := common.HexToAddress(address)
	balance, err := ts.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	return balance, nil
}

// GetNonce 获取nonce
func (ts *TransactionService) GetNonce(address string) (uint64, error) {
	if !common.IsHexAddress(address) {
		return 0, fmt.Errorf("invalid address format")
	}

	account := common.HexToAddress(address)
	nonce, err := ts.client.PendingNonceAt(context.Background(), account)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %v", err)
	}

	return nonce, nil
}

// WaitForTransaction 等待交易确认
func (ts *TransactionService) WaitForTransaction(txHash string, confirmations int) error {
	// hash := common.HexToHash(txHash)

	// 等待交易被包含在区块中
	// receipt, err := ts.client.WaitMined(context.Background(), &types.Transaction{})
	// if err != nil {
	// 	return fmt.Errorf("failed to wait for transaction: %v", err)
	// }

	// 暂时返回nil，需要实现正确的等待逻辑
	return nil
}

// UpdateTransactionStatus 更新交易状态
func (ts *TransactionService) UpdateTransactionStatus(txHash string) error {
	response, err := ts.GetTransactionStatus(txHash)
	if err != nil {
		return err
	}

	// 更新数据库中的交易状态
	var chainBill models.ChainBill
	if err := database.GetDB().Where("txid = ?", txHash).First(&chainBill).Error; err != nil {
		return fmt.Errorf("failed to find transaction: %v", err)
	}

	// 更新状态
	var status int
	switch response.Status {
	case "success":
		status = 1
	case "failed":
		status = 2
	default:
		status = 0
	}

	chainBill.Status = status
	chainBill.UpdatedTime = time.Now()

	if err := database.GetDB().Save(&chainBill).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %v", err)
	}

	return nil
}

// Close 关闭连接
func (ts *TransactionService) Close() {
	if ts.client != nil {
		ts.client.Close()
	}
}
