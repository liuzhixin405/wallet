package database

import (
	"fmt"
	"log"
	"wallet-backend/internal/config"
	"wallet-backend/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	dsn := cfg.Database.GetDSN()

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// 测试连接
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connected successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	// 自动迁移所有模型
	err := DB.AutoMigrate(
		&models.User{},
		&models.AddressLibrary{},
		&models.Balance{},
		&models.WithdrawRecord{},
		&models.DepositRecord{},
		&models.ChainBill{},
		&models.CurrencyChainConfig{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// CreateDefaultData 创建默认数据
func CreateDefaultData() error {
	// 创建默认币种配置
	if err := createDefaultCurrencies(); err != nil {
		return fmt.Errorf("failed to create default currencies: %v", err)
	}

	return nil
}

// createDefaultCurrencies 创建默认币种配置
func createDefaultCurrencies() error {
	// 检查是否已存在币种配置
	var count int64
	DB.Model(&models.CurrencyChainConfig{}).Count(&count)
	if count > 0 {
		return nil // 已存在配置，跳过
	}

	// 创建默认币种配置
	defaultCurrencies := []models.CurrencyChainConfig{
		{
			Symbol:            "ETH",
			ChainType:         "Ethereum",
			IsEnabled:         true,
			RPCURL:            "https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID",
			ChainID:           11155111,
			Confirmations:     12,
			Decimals:          18,
			CollectionEnabled: true,
			CollectionThreshold: "0.1",
		},
		{
			Symbol:            "USDT",
			ChainType:         "Ethereum",
			IsEnabled:         true,
			RPCURL:            "https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID",
			ChainID:           11155111,
			Confirmations:     12,
			Decimals:          6,
			CollectionEnabled: true,
			CollectionThreshold: "10",
		},
		{
			Symbol:            "USDC",
			ChainType:         "Ethereum",
			IsEnabled:         true,
			RPCURL:            "https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID",
			ChainID:           11155111,
			Confirmations:     12,
			Decimals:          6,
			CollectionEnabled: true,
			CollectionThreshold: "10",
		},
		{
			Symbol:            "BNB",
			ChainType:         "BSC",
			IsEnabled:         true,
			RPCURL:            "https://bsc-dataseed1.binance.org/",
			ChainID:           56,
			Confirmations:     15,
			Decimals:          18,
			CollectionEnabled: true,
			CollectionThreshold: "0.1",
		},
		{
			Symbol:            "BUSD",
			ChainType:         "BSC",
			IsEnabled:         true,
			RPCURL:            "https://bsc-dataseed1.binance.org/",
			ChainID:           56,
			Confirmations:     15,
			Decimals:          18,
			CollectionEnabled: true,
			CollectionThreshold: "10",
		},
	}

	for _, currency := range defaultCurrencies {
		if err := DB.Create(&currency).Error; err != nil {
			return fmt.Errorf("failed to create currency %s: %v", currency.Symbol, err)
		}
	}

	log.Println("Default currencies created successfully")
	return nil
}
