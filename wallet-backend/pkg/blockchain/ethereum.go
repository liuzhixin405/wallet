package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumClient 以太坊客户端
type EthereumClient struct {
	client *ethclient.Client
	url    string
}

// NewEthereumClient 创建新的以太坊客户端
func NewEthereumClient(url string) (*EthereumClient, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}

	return &EthereumClient{
		client: client,
		url:    url,
	}, nil
}

// GetBalance 获取地址余额
func (ec *EthereumClient) GetBalance(address string) (*big.Int, error) {
	account := common.HexToAddress(address)
	balance, err := ec.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	return balance, nil
}

// GetTransactionCount 获取地址交易数量
func (ec *EthereumClient) GetTransactionCount(address string) (uint64, error) {
	_ = common.HexToAddress(address) // 验证地址格式
	count, err := ec.client.PendingTransactionCount(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get transaction count: %v", err)
	}

	return count, nil
}

// GetBlockNumber 获取最新区块号
func (ec *EthereumClient) GetBlockNumber() (uint64, error) {
	blockNumber, err := ec.client.BlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %v", err)
	}

	return blockNumber, nil
}

// Close 关闭客户端连接
func (ec *EthereumClient) Close() {
	if ec.client != nil {
		ec.client.Close()
	}
}
