package services

import (
	"context"
	"fmt"
	"math/big"
	"wallet-backend/internal/models"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BlockchainService 区块链服务
type BlockchainService struct {
	ethClient *ethclient.Client
	// btcClient *rpcclient.Client // 移除实际依赖
}

// NewBlockchainService 创建新的区块链服务实例
func NewBlockchainService() *BlockchainService {
	return &BlockchainService{}
}

// GetEthereumBalance 获取以太坊地址余额
func (bs *BlockchainService) GetEthereumBalance(address string) (*big.Int, error) {
	if bs.ethClient == nil {
		// 这里应该从配置中获取以太坊节点URL
		client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Ethereum node: %v", err)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Ethereum node: %v", err)
		}
		bs.ethClient = client
	}

	account := common.HexToAddress(address)
	balance, err := bs.ethClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	return balance, nil
}

// GetBitcoinBalance 获取比特币地址余额
func (bs *BlockchainService) GetBitcoinBalance(address string) (float64, error) {
	// TODO: 接入实际的 Bitcoin 节点 RPC 客户端

	// 获取地址信息
	// addressInfo, err := bs.btcClient.GetAddressInfo(address)
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to get address info: %v", err)
	// }

	// 转换为BTC单位
	// 注意：GetAddressInfo可能不返回Balance字段，需要根据实际API调整
	balance := float64(0) / 100000000 // 临时设置为0，需要根据实际API调整
	return balance, nil
}

// GetTransactionStatus 获取交易状态
func (bs *BlockchainService) GetTransactionStatus(txHash, chainType string) (*models.ChainBill, error) {
	switch chainType {
	case "Ethereum":
		return bs.getEthereumTransactionStatus(txHash)
	case "Bitcoin":
		return bs.getBitcoinTransactionStatus(txHash)
	default:
		return nil, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

// getEthereumTransactionStatus 获取以太坊交易状态
func (bs *BlockchainService) getEthereumTransactionStatus(txHash string) (*models.ChainBill, error) {
	if bs.ethClient == nil {
		return nil, fmt.Errorf("Ethereum client not initialized")
	}

	hash := common.HexToHash(txHash)
	_, isPending, err := bs.ethClient.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}

	receipt, err := bs.ethClient.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	status := 0 // 确认中
	if !isPending {
		if receipt.Status == 1 {
			status = 1 // 已确认
		} else {
			status = 2 // 失败
		}
	}

	blockHeight := receipt.BlockNumber.Uint64()
	chainBill := &models.ChainBill{
		TxID:          txHash,
		Status:        status,
		BlockHeight:   &blockHeight,
		Confirmations: 0, // 需要计算确认数
	}

	return chainBill, nil
}

// getBitcoinTransactionStatus 获取比特币交易状态
func (bs *BlockchainService) getBitcoinTransactionStatus(txHash string) (*models.ChainBill, error) {
	// if bs.btcClient == nil {
	// 	return nil, fmt.Errorf("Bitcoin client not initialized")
	// }

	// 获取交易信息
	// hash, err := chainhash.NewHashFromStr(txHash)
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid transaction hash: %v", err)
	// }
	
	// tx, err := bs.btcClient.GetRawTransaction(hash)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get transaction: %v", err)
	// }

	status := 0
	var blockHeight *uint64
	// 注意：比特币交易的区块信息可能需要通过其他方式获取
	// 这里暂时设置为未确认状态
	// if tx.BlockHash != nil {
	// 	status = 1 // 已确认
	// 	blockHeight = &tx.BlockIndex
	// }

	chainBill := &models.ChainBill{
		TxID:          txHash,
		Status:        status,
		BlockHeight:   blockHeight,
		Confirmations: 0, // 需要计算确认数
	}

	return chainBill, nil
}

// EstimateGas 估算以太坊交易Gas费用
func (bs *BlockchainService) EstimateGas(from, to string, value *big.Int, data []byte) (uint64, error) {
	if bs.ethClient == nil {
		return 0, fmt.Errorf("Ethereum client not initialized")
	}

	msg := ethereum.CallMsg{
		From:  common.HexToAddress(from),
		To:    &common.Address{},
		Value: value,
		Data:  data,
	}

	if to != "" {
		toAddr := common.HexToAddress(to)
		msg.To = &toAddr
	}

	gasLimit, err := bs.ethClient.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}

	return gasLimit, nil
}
