package main

import (
	"log"
	"wallet-backend/internal/config"
	"wallet-backend/internal/database"
	"wallet-backend/internal/routes"
	"wallet-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 执行数据库迁移
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建默认数据
	if err := database.CreateDefaultData(); err != nil {
		log.Fatalf("Failed to create default data: %v", err)
	}

	// 初始化服务
	hdWalletService := services.NewHDWalletService(cfg)
	wsService := services.NewWebSocketService()

	blockScannerService, _ := services.NewBlockScannerService(cfg)
	collectionService, _ := services.NewCollectionService(cfg)
	
	// 创建定时任务服务
	schedulerService := services.NewSchedulerService(cfg, blockScannerService, collectionService)

	// 暂时注释掉有问题的服务
	// transactionService, err := services.NewTransactionService(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to create transaction service: %v", err)
	// }
	// defer transactionService.Close()

	// 启动WebSocket服务
	go wsService.Start()
	
	// 启动定时任务服务
	go schedulerService.Start()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	r := gin.Default()

	// 创建服务配置
	serviceConfig := &services.Config{
		AppConfig:          cfg,
		HDWalletService:    hdWalletService,
		WSService:          wsService,
		BlockScannerService: blockScannerService,
		CollectionService:  collectionService,
	}

	// 设置路由
	routes.SetupRoutes(r, serviceConfig)

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
