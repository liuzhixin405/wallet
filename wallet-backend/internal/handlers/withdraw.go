package handlers

import (
	"net/http"
	"time"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"
	"wallet-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// CreateWithdraw 创建提币申请
func CreateWithdraw(c *gin.Context) {
	var req struct {
		CurrencySymbol string  `json:"currency_symbol" binding:"required"`
		ChainType      string  `json:"chain_type" binding:"required"`
		Protocol       string  `json:"protocol,omitempty"`
		ToAddress      string  `json:"to_address" binding:"required"`
		Amount         float64 `json:"amount" binding:"required,gt=0"`
		Remark         string  `json:"remark,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userID, _ := c.Get("user_id")

	// 检查余额是否足够
	var balance models.Balance
	if err := database.GetDB().Where("currency_symbol = ? AND chain_type = ? AND address IN (SELECT address FROM address_library WHERE user_id = ?)",
		req.CurrencySymbol, req.ChainType, userID).First(&balance).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	if balance.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// 获取手续费配置
	var fee float64 = 0.001 // 默认手续费，实际应该从配置表获取

	withdraw := models.WithdrawRecord{
		CurrencySymbol: req.CurrencySymbol,
		ChainType:      req.ChainType,
		Protocol:       req.Protocol,
		UserID:         userID.(uint64),
		FromAddress:    balance.Address,
		ToAddress:      req.ToAddress,
		Amount:         req.Amount,
		Fee:            fee,
		TotalAmount:    req.Amount + fee,
		UniqueID:       generateUniqueID(),
		Status:         0, // 待转手续费
		CreatedAt:      time.Now(),
		Type:           &[]int{1}[0], // 1:提币
	}

	if err := database.GetDB().Create(&withdraw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create withdraw request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": withdraw})
}

// GetWithdraws 获取提币记录
func GetWithdraws(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var withdraws []models.WithdrawRecord
	if err := database.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&withdraws).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdraws"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": withdraws})
}

// GetWithdrawByID 获取指定提币记录
func GetWithdrawByID(c *gin.Context) {
	userID, _ := c.Get("user_id")
	withdrawID := c.Param("id")

	var withdraw models.WithdrawRecord
	if err := database.GetDB().Where("id = ? AND user_id = ?", withdrawID, userID).First(&withdraw).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdraw record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": withdraw})
}

// 生成唯一ID
func generateUniqueID() string {
	// 生成提现单号
	randomHex, err := utils.GenerateRandomHex(8)
	if err != nil {
		// 如果生成失败，使用时间戳作为备选
		return "W" + time.Now().Format("20060102150405")
	}
	return "W" + time.Now().Format("20060102150405") + randomHex
}
