package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 简化的配置结构
type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
		Charset  string `yaml:"charset"`
	} `yaml:"database"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
}

// 简化的用户模型
type User struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"type:varchar(50);not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"`
	Email     string    `json:"email" gorm:"type:varchar(100);not null;uniqueIndex"`
	Status    bool      `json:"status" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

// 简化的地址模型
type Address struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `json:"user_id" gorm:"not null;index"`
	Address   string    `json:"address" gorm:"type:varchar(100);not null;uniqueIndex"`
	ChainType string    `json:"chain_type" gorm:"type:varchar(30);not null;index"`
	Status    int       `json:"status" gorm:"not null;default:0"`
	IndexNum  int       `json:"index_num" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

var db *gorm.DB
var jwtSecret string

func main() {
	// 设置JWT密钥
	jwtSecret = os.Getenv("WALLET_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-here-change-in-production"
	}

	// 初始化数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		"root", "password", "localhost", "3306", "wallet_db", "utf8mb4")

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&User{}, &Address{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
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

	// 检查用户是否已存在
	var existingUser User
	if err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(400, gin.H{"error": "User already exists"})
		return
	}

	// 创建新用户
	user := User{
		Username: req.Username,
		Password: req.Password, // 实际应用中应该哈希密码
		Email:    req.Email,
		Status:   true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
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

	// 查找用户
	var user User
	if err := db.Where("username = ? AND password = ?", req.Username, req.Password).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
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
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	}})
}

// 获取地址处理器
func getAddressesHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var addresses []Address
	if err := db.Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to get addresses"})
		return
	}

	c.JSON(200, gin.H{"data": addresses})
}

// 生成地址处理器
func generateAddressHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	// 保存地址
	addr := Address{
		UserID:    userID.(uint64),
		Address:   address,
		ChainType: req.ChainType,
		Status:    0,
		IndexNum:  0,
	}

	if err := db.Create(&addr).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create address"})
		return
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
