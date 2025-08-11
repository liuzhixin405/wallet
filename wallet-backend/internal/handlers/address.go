package handlers

import (
	"net/http"
	"time"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"
	"wallet-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetAddresses 获取用户地址列表
func GetAddresses(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var addresses []models.AddressLibrary
	if err := database.GetDB().Where("userid = ?", userID).Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch addresses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": addresses})
}

// GenerateAddress 生成新地址
func GenerateAddress(c *gin.Context) {
	var req struct {
		ChainType string `json:"chain_type" binding:"required"`
		Protocol  string `json:"protocol,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userID, _ := c.Get("user_id")

	// 类型断言
	userIDUint64, ok := userID.(uint64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 这里应该调用钱包服务生成地址，暂时使用模拟数据
	randomHex, err := utils.GenerateRandomHex(40)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate address"})
		return
	}

	address := models.AddressLibrary{
		UserID:      &userIDUint64,
		Address:     "0x" + randomHex, // 模拟地址
		ChainType:   req.ChainType,
		Status:      0, // 未使用
		IndexNum:    0,
		CreatedTime: time.Now(),
	}

	if err := database.GetDB().Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate address"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": address})
}

// BindAddress 绑定地址
func BindAddress(c *gin.Context) {
	var req struct {
		Address   string `json:"address" binding:"required"`
		ChainType string `json:"chain_type" binding:"required"`
		Note      string `json:"note,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userID, _ := c.Get("user_id")

	// 类型断言
	userIDUint64, ok := userID.(uint64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 检查地址是否已存在
	var existingAddress models.AddressLibrary
	if err := database.GetDB().Where("address = ? AND chain_type = ?", req.Address, req.ChainType).First(&existingAddress).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Address already exists"})
		return
	}

	address := models.AddressLibrary{
		UserID:      &userIDUint64,
		Address:     req.Address,
		ChainType:   req.ChainType,
		Status:      1, // 已激活
		BindTime:    &[]time.Time{time.Now()}[0],
		Note:        req.Note,
		CreatedTime: time.Now(),
	}

	if err := database.GetDB().Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bind address"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": address})
}
