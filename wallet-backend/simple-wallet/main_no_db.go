package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret string

func main() {
	// 设置JWT密钥
	jwtSecret = os.Getenv("WALLET_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-here-change-in-production"
	}

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

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "wallet-backend",
			"version":   "1.0.0",
		})
	})

	r.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ready",
			"checks": gin.H{
				"database": "ok",
			},
		})
	})

	r.GET("/live", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})

	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"timestamp": time.Now().Unix(),
			"uptime":    "0s",
			"version":   "1.0.0",
			"connections": gin.H{
				"total":    0,
				"active":   0,
				"inactive": 0,
			},
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/register", registerHandler)
			auth.POST("/login", loginHandler)
		}

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(authMiddleware())
		{
			// 地址管理
			addresses := authorized.Group("/addresses")
			{
				addresses.GET("", getAddressesHandler)
				addresses.POST("/generate", generateAddressHandler)
			}

			// HD钱包相关
			wallet := authorized.Group("/wallet")
			{
				wallet.POST("/mnemonic/generate", generateMnemonicHandler)
				wallet.POST("/mnemonic/validate", validateMnemonicHandler)
				wallet.POST("/address/derive", deriveAddressHandler)
			}
		}
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// 认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]

		// 解析JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 从token中获取用户信息
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID, exists := claims["user_id"]
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}
			c.Set("user_id", userID)
		}

		c.Next()
	}
}

// 注册处理器
func registerHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 模拟用户创建
	userID := uint64(time.Now().Unix())

	c.JSON(200, gin.H{"data": gin.H{
		"user_id":  userID,
		"username": req.Username,
		"email":    req.Email,
	}})
}

// 登录处理器
func loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 模拟用户验证
	if req.Username == "" || req.Password == "" {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  uint64(time.Now().Unix()),
		"username": req.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       uint64(time.Now().Unix()),
			"username": req.Username,
			"email":    req.Username + "@example.com",
		},
	}})
}

// 获取地址处理器
func getAddressesHandler(c *gin.Context) {
	// 返回模拟地址
	addresses := []gin.H{
		{
			"id":         1,
			"user_id":    1,
			"address":    "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
			"chain_type": "Ethereum",
			"status":     0,
			"index_num":  0,
		},
		{
			"id":         2,
			"user_id":    1,
			"address":    "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			"chain_type": "Bitcoin",
			"status":     0,
			"index_num":  0,
		},
	}

	c.JSON(200, gin.H{"data": addresses})
}

// 生成地址处理器
func generateAddressHandler(c *gin.Context) {
	var req struct {
		ChainType string `json:"chain_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 生成随机地址
	bytes := make([]byte, 20)
	rand.Read(bytes)
	address := "0x" + hex.EncodeToString(bytes)

	addr := gin.H{
		"id":         uint64(time.Now().Unix()),
		"user_id":    1,
		"address":    address,
		"chain_type": req.ChainType,
		"status":     0,
		"index_num":  0,
	}

	c.JSON(200, gin.H{"data": addr})
}

// 生成助记词处理器
func generateMnemonicHandler(c *gin.Context) {
	// 生成一个示例助记词
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	c.JSON(200, gin.H{"data": gin.H{"mnemonic": mnemonic}})
}

// 验证助记词处理器
func validateMnemonicHandler(c *gin.Context) {
	var req struct {
		Mnemonic string `json:"mnemonic" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 简单的验证逻辑
	isValid := len(req.Mnemonic) > 0 && len(req.Mnemonic) < 200

	c.JSON(200, gin.H{"data": gin.H{"valid": isValid}})
}

// 派生地址处理器
func deriveAddressHandler(c *gin.Context) {
	var req struct {
		Mnemonic  string `json:"mnemonic" binding:"required"`
		ChainType string `json:"chain_type" binding:"required"`
		Index     uint32 `json:"index"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 简化的地址派生
	var address string
	switch req.ChainType {
	case "Ethereum":
		bytes := make([]byte, 20)
		rand.Read(bytes)
		address = "0x" + hex.EncodeToString(bytes)
	case "Bitcoin":
		address = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	default:
		c.JSON(400, gin.H{"error": "Unsupported chain type"})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{
		"address":    address,
		"chain_type": req.ChainType,
		"status":     0,
		"index_num":  req.Index,
	}})
}
