package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"wallet-backend/internal/config"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumWalletService 以太坊钱包服务
type EthereumWalletService struct {
	config *config.Config
	client *ethclient.Client
}

// NewEthereumWalletService 创建新的以太坊钱包服务
func NewEthereumWalletService(cfg *config.Config) (*EthereumWalletService, error) {
	// 连接到以太坊测试网
	client, err := ethclient.Dial(cfg.Ethereum.GetTestnetRPCURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum testnet: %v", err)
	}

	return &EthereumWalletService{
		config: cfg,
		client: client,
	}, nil
}

// GenerateAddress 生成以太坊地址
func (ews *EthereumWalletService) GenerateAddress(index uint32) (*models.AddressLibrary, error) {
	// 生成新的私钥（在实际应用中应该从HD钱包派生）
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// 获取公钥
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}

	// 生成地址
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 创建地址模型
	addressModel := &models.AddressLibrary{
		Address:   address.Hex(),
		ChainType: "Ethereum",
		Status:    0, // 未使用
		IndexNum:  uint64(index),
	}

	return addressModel, nil
}

// GetBalance 获取地址余额
func (ews *EthereumWalletService) GetBalance(address string) (*big.Int, error) {
	account := common.HexToAddress(address)
	balance, err := ews.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}
	return balance, nil
}

// GetTransactionCount 获取地址交易数量
func (ews *EthereumWalletService) GetTransactionCount(address string) (uint64, error) {
	account := common.HexToAddress(address)
	nonce, err := ews.client.PendingNonceAt(context.Background(), account)
	if err != nil {
		return 0, fmt.Errorf("failed to get transaction count: %v", err)
	}
	return nonce, nil
}

// CreateWithdrawal 创建提现交易
func (ews *EthereumWalletService) CreateWithdrawal(fromAddress, toAddress string, amount *big.Int, gasPrice *big.Int) (*types.Transaction, error) {
	// 获取发送方私钥（在实际应用中应该从安全的存储中获取）
	privateKey, err := ews.getPrivateKeyByAddress(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %v", err)
	}

	// 获取nonce
	nonce, err := ews.GetTransactionCount(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// 估算gas limit
	msg := ethereum.CallMsg{
		From:  common.HexToAddress(fromAddress),
		To:    &common.Address{},
		Value: amount,
		Data:  []byte{},
	}

	if toAddress != "" {
		toAddr := common.HexToAddress(toAddress)
		msg.To = &toAddr
	}

	gasLimit, err := ews.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %v", err)
	}

	// 创建交易
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(toAddress),
		amount,
		gasLimit,
		gasPrice,
		nil,
	)

	// 签名交易
	chainID, err := ews.client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return signedTx, nil
}

// SendTransaction 发送交易
func (ews *EthereumWalletService) SendTransaction(tx *types.Transaction) error {
	err := ews.client.SendTransaction(context.Background(), tx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}
	return nil
}

// GetTransactionStatus 获取交易状态
func (ews *EthereumWalletService) GetTransactionStatus(txHash string) (*models.ChainBill, error) {
	hash := common.HexToHash(txHash)

	// 获取交易
	_, isPending, err := ews.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}

	// 获取交易收据
	receipt, err := ews.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		if isPending {
			// 交易仍在待处理状态
			chainBill := &models.ChainBill{
				TxID:          txHash,
				Status:        0, // 待处理
				BlockHeight:   nil,
				Confirmations: 0,
			}
			return chainBill, nil
		}
		return nil, fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	// 获取当前区块号
	currentBlock, err := ews.client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get current block number: %v", err)
	}

	// 计算确认数
	confirmations := currentBlock - receipt.BlockNumber.Uint64()

	// 确定状态
	var status int
	if receipt.Status == 1 {
		status = 1 // 成功
	} else {
		status = 2 // 失败
	}

	blockHeight := receipt.BlockNumber.Uint64()

	chainBill := &models.ChainBill{
		TxID:          txHash,
		Status:        status,
		BlockHeight:   &blockHeight,
		Confirmations: int(confirmations),
	}

	return chainBill, nil
}

// getPrivateKeyByAddress 根据地址获取私钥
func (ews *EthereumWalletService) getPrivateKeyByAddress(address string) (*ecdsa.PrivateKey, error) {
	// 这里需要实现从地址索引到私钥的映射
	// 在实际应用中，这通常通过数据库查询或内存映射实现
	// 暂时返回错误，需要实现完整的地址管理
	return nil, fmt.Errorf("private key retrieval not implemented yet")
}

// Close 关闭客户端连接
func (ews *EthereumWalletService) Close() {
	if ews.client != nil {
		ews.client.Close()
	}
}
