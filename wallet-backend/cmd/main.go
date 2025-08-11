package main

import (
	"log"
	"wallet-backend/internal/config"
	"wallet-backend/internal/database"
	"wallet-backend/internal/handlers"
	"wallet-backend/internal/middleware"
	"wallet-backend/internal/models"
	"wallet-backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	cfg, err := config.LoadConfig("config.yaml")
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

	// 暂时注释掉有问题的服务
	// transactionService, err := services.NewTransactionService(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to create transaction service: %v", err)
	// }
	// defer transactionService.Close()

	// 启动WebSocket服务
	go wsService.Start()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// 健康检查路由
	r.GET("/health", handlers.HealthCheck)
	r.GET("/ready", handlers.ReadinessCheck)
	r.GET("/live", handlers.LivenessCheck)
	r.GET("/metrics", handlers.Metrics)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
		}

		// WebSocket路由
		wsHandler := handlers.NewWebSocketHandler(wsService)
		api.GET("/ws", middleware.AuthMiddleware(), wsHandler.HandleWebSocket)
		api.GET("/ws/stats", middleware.AuthMiddleware(), wsHandler.GetWebSocketStats)

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware())
		{
			// 地址管理
			addresses := authorized.Group("/addresses")
			{
				addresses.GET("", handlers.GetAddresses)
				addresses.POST("/generate", handlers.GenerateAddress)
				addresses.POST("/bind", handlers.BindAddress)
			}

			// 余额管理
			balances := authorized.Group("/balances")
			{
				balances.GET("", handlers.GetBalances)
				balances.GET("/:currency", handlers.GetBalanceByCurrency)
			}

			// 提币管理
			withdraws := authorized.Group("/withdraws")
			{
				withdraws.GET("", handlers.GetWithdraws)
				withdraws.POST("", handlers.CreateWithdraw)
				withdraws.GET("/:id", handlers.GetWithdrawByID)
			}

			// 充值记录
			deposits := authorized.Group("/deposits")
			{
				deposits.GET("", handlers.GetDeposits)
				deposits.GET("/:id", handlers.GetDepositByID)
				deposits.GET("/address/:address", handlers.GetDepositsByAddress)
			}

			// 交易记录
			transactions := authorized.Group("/transactions")
			{
				transactions.GET("", handlers.GetTransactions)
				transactions.GET("/:id", handlers.GetTransactionByID)
				transactions.GET("/address/:address", handlers.GetTransactionsByAddress)
				transactions.GET("/currency/:currency", handlers.GetTransactionsByCurrency)
			}

			// 货币配置
			currencies := authorized.Group("/currencies")
			{
				currencies.GET("", handlers.GetCurrencies)
				currencies.GET("/:symbol", handlers.GetCurrencyBySymbol)
				currencies.GET("/chains/supported", handlers.GetSupportedChains)
			}

			// HD钱包相关
			wallet := authorized.Group("/wallet")
			{
				wallet.POST("/mnemonic/generate", func(c *gin.Context) {
					mnemonic, err := hdWalletService.GenerateMnemonic()
					if err != nil {
						c.JSON(500, gin.H{"error": err.Error()})
						return
					}
					c.JSON(200, gin.H{"data": gin.H{"mnemonic": mnemonic}})
				})

				wallet.POST("/mnemonic/validate", func(c *gin.Context) {
					var req struct {
						Mnemonic string `json:"mnemonic" binding:"required"`
					}
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(400, gin.H{"error": "Invalid request"})
						return
					}

					isValid := hdWalletService.ValidateMnemonic(req.Mnemonic)
					c.JSON(200, gin.H{"data": gin.H{"valid": isValid}})
				})

				wallet.POST("/address/derive", func(c *gin.Context) {
					var req struct {
						Mnemonic  string `json:"mnemonic" binding:"required"`
						ChainType string `json:"chain_type" binding:"required"`
						Index     uint32 `json:"index"`
					}
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(400, gin.H{"error": "Invalid request"})
						return
					}

					var address *models.AddressLibrary
					var err error
					switch req.ChainType {
					case "Ethereum":
						address, err = hdWalletService.DeriveEthereumAddress(req.Mnemonic, req.Index)
					case "Bitcoin":
						address, err = hdWalletService.DeriveBitcoinAddress(req.Mnemonic, req.Index)
					default:
						c.JSON(400, gin.H{"error": "Unsupported chain type"})
						return
					}

					if err != nil {
						c.JSON(500, gin.H{"error": err.Error()})
						return
					}

					c.JSON(200, gin.H{"data": address})
				})
			}
		}
	}

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
