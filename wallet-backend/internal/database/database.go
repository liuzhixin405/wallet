package database

import (
	"fmt"
	"log"
	"time"
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

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	log.Println("Starting database migration...")

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
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// CreateDefaultData 创建默认数据
func CreateDefaultData() error {
	log.Println("Creating default data...")

	// 创建默认货币配置
	defaultCurrencies := []models.CurrencyChainConfig{
		{
			CurrencySymbol: "ETH",
			ChainType:      "Ethereum",
			Protocol:       nil,
			Decimals:       18,
			MinWithdraw:    0.001,
			MaxWithdraw:    100.0,
			WithdrawFee:    0.001,
			DepositFee:     0.0,
			Status:         true,
			CreatedTime:    time.Now(),
		},
		{
			CurrencySymbol: "BTC",
			ChainType:      "Bitcoin",
			Protocol:       nil,
			Decimals:       8,
			MinWithdraw:    0.0001,
			MaxWithdraw:    10.0,
			WithdrawFee:    0.0001,
			DepositFee:     0.0,
			Status:         true,
			CreatedTime:    time.Now(),
		},
	}

	for _, currency := range defaultCurrencies {
		var existing models.CurrencyChainConfig
		if err := DB.Where("currency_symbol = ? AND chain_type = ?", currency.CurrencySymbol, currency.ChainType).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := DB.Create(&currency).Error; err != nil {
					log.Printf("Failed to create default currency %s: %v", currency.CurrencySymbol, err)
				} else {
					log.Printf("Created default currency: %s", currency.CurrencySymbol)
				}
			}
		}
	}

	log.Println("Default data creation completed")
	return nil
}
