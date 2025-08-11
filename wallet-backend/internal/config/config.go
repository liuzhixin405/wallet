package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 配置结构
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Ethereum EthereumConfig `mapstructure:"ethereum"`
	BSC      *BSCConfig     `mapstructure:"bsc"`
	Wallet   WalletConfig   `mapstructure:"wallet"`
	Scanner  ScannerConfig  `mapstructure:"scanner"`
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

// EthereumConfig 以太坊配置
type EthereumConfig struct {
	Testnet TestnetConfig `mapstructure:"testnet"`
	Mainnet MainnetConfig `mapstructure:"mainnet"`
}

// BSCConfig BSC配置
type BSCConfig struct {
	RPCURL       string `mapstructure:"rpc_url"`
	ChainID      int64  `mapstructure:"chain_id"`
	Confirmations int   `mapstructure:"confirmations"`
}

// TestnetConfig 测试网配置
type TestnetConfig struct {
	RPCURL       string `mapstructure:"rpc_url"`
	ChainID      int64  `mapstructure:"chain_id"`
	Confirmations int   `mapstructure:"confirmations"`
}

// MainnetConfig 主网配置
type MainnetConfig struct {
	RPCURL        string `mapstructure:"rpc_url"`
	ChainID       int64  `mapstructure:"chain_id"`
	Confirmations int    `mapstructure:"confirmations"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	RPCURL        string `mapstructure:"rpc_url"`
	ChainID       int64  `mapstructure:"chain_id"`
	Confirmations int    `mapstructure:"confirmations"`
}

// WalletConfig 钱包配置
type WalletConfig struct {
	HDWallet   HDWalletConfig   `mapstructure:"hd_wallet"`
	HotWallet  HotWalletConfig  `mapstructure:"hot_wallet"`
	ColdWallet ColdWalletConfig `mapstructure:"cold_wallet"`
}

// HDWalletConfig HD钱包配置
type HDWalletConfig struct {
	Mnemonic       string `mapstructure:"mnemonic"`
	DerivationPath string `mapstructure:"derivation_path"`
}

// HotWalletConfig 热钱包配置
type HotWalletConfig struct {
	MaxBalance          string `mapstructure:"max_balance"`
	CollectionThreshold string `mapstructure:"collection_threshold"`
}

// ColdWalletConfig 冷钱包配置
type ColdWalletConfig struct {
	Address string `mapstructure:"address"`
}

// ScannerConfig 扫描配置
type ScannerConfig struct {
	ScanInterval     int `mapstructure:"scan_interval"`
	MaxBlocksPerScan int `mapstructure:"max_blocks_per_scan"`
	RetryAttempts    int `mapstructure:"retry_attempts"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	// 设置环境变量前缀
	viper.SetEnvPrefix("WALLET")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// 设置默认值
	config.setDefaults()

	return &config, nil
}

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	if c.Database.Port == "" {
		c.Database.Port = "3306"
	}
	if c.Database.Charset == "" {
		c.Database.Charset = "utf8mb4"
	}
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Scanner.ScanInterval == 0 {
		c.Scanner.ScanInterval = 15
	}
	if c.Scanner.MaxBlocksPerScan == 0 {
		c.Scanner.MaxBlocksPerScan = 100
	}
	if c.Scanner.RetryAttempts == 0 {
		c.Scanner.RetryAttempts = 3
	}
	if c.JWT.ExpirationHours == 0 {
		c.JWT.ExpirationHours = 24
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&innodb_strict_mode=0",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
	)
}

// GetTestnetRPCURL 获取测试网RPC URL
func (c *EthereumConfig) GetTestnetRPCURL() string {
	return c.Testnet.RPCURL
}

// GetMainnetRPCURL 获取主网RPC URL
func (c *EthereumConfig) GetMainnetRPCURL() string {
	return c.Mainnet.RPCURL
}

// GetChainID 获取链ID
func (c *NetworkConfig) GetChainID() int64 {
	return c.ChainID
}

// GetConfirmations 获取确认数
func (c *NetworkConfig) GetConfirmations() int {
	return c.Confirmations
}
