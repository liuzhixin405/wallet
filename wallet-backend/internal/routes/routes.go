package routes

import (
	"wallet-backend/internal/handlers"
	"wallet-backend/internal/middleware"
	"wallet-backend/internal/models"
	"wallet-backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, cfg *services.Config) {
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
		wsHandler := handlers.NewWebSocketHandler(cfg.WSService)
		api.GET("/ws", middleware.AuthMiddleware(), wsHandler.HandleWebSocket)
		api.GET("/ws/stats", middleware.AuthMiddleware(), wsHandler.GetWebSocketStats)

		// 运维控制路由（需要认证）
		opsHandler := handlers.NewOpsHandler(cfg.BlockScannerService, cfg.CollectionService)
		toolsHandler := handlers.NewToolsHandler(cfg.BlockScannerService, cfg.CollectionService)

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

			// 货币配置相关
			currencies := authorized.Group("/currencies")
			{
				currencies.GET("", handlers.GetCurrencies)
				currencies.GET("/:symbol", handlers.GetCurrencyBySymbol)
				currencies.POST("", handlers.CreateCurrency)
				currencies.PUT("/:symbol", handlers.UpdateCurrency)
				currencies.DELETE("/:symbol", handlers.DeleteCurrency)
				currencies.POST("/:symbol/enable", handlers.EnableCurrency)
				currencies.POST("/:symbol/disable", handlers.DisableCurrency)
				currencies.GET("/chains/supported", handlers.GetSupportedChains)
			}

			// HD钱包相关
			wallet := authorized.Group("/wallet")
			{
				wallet.POST("/mnemonic/generate", func(c *gin.Context) {
					mnemonic, err := cfg.HDWalletService.GenerateMnemonic()
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

					isValid := cfg.HDWalletService.ValidateMnemonic(req.Mnemonic)
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
						address, err = cfg.HDWalletService.DeriveEthereumAddress(req.Mnemonic, req.Index)
					case "Bitcoin":
						address, err = cfg.HDWalletService.DeriveBitcoinAddress(req.Mnemonic, req.Index)
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

			// 操作相关
			ops := authorized.Group("/ops")
			{
				// 扫描器操作
				scanner := ops.Group("/scanner")
				{
					scanner.POST("/start", opsHandler.StartScanner)
					scanner.POST("/stop", opsHandler.StopScanner)
					scanner.POST("/scan-once", opsHandler.ScanOnce)
					scanner.GET("/status", opsHandler.ScannerStatus)
					scanner.POST("/scan-blocks", toolsHandler.ScanBlocks)
				}

				// 归集操作
				collection := ops.Group("/collection")
				{
					collection.POST("/start", opsHandler.StartCollection)
					collection.POST("/stop", opsHandler.StopCollection)
					collection.POST("/trigger", toolsHandler.TriggerCollection)
				}
			}

			// 工具管理
			tools := authorized.Group("/tools")
			{
				tools.GET("/balances", toolsHandler.GetBalances)
				tools.GET("/addresses", toolsHandler.GetAddresses)
			}
		}
	}
} 