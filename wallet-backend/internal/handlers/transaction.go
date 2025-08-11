package handlers

import (
	"net/http"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetTransactions 获取交易记录
func GetTransactions(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var transactions []models.ChainBill
	if err := database.GetDB().Where("user_id = ?", userID).Order("created_time DESC").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

// GetTransactionByID 获取指定交易记录
func GetTransactionByID(c *gin.Context) {
	userID, _ := c.Get("user_id")
	txID := c.Param("id")

	var transaction models.ChainBill
	if err := database.GetDB().Where("id = ? AND user_id = ?", txID, userID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transaction})
}

// GetTransactionsByAddress 根据地址获取交易记录
func GetTransactionsByAddress(c *gin.Context) {
	userID, _ := c.Get("user_id")
	address := c.Param("address")

	var transactions []models.ChainBill
	if err := database.GetDB().Where("user_id = ? AND address = ?", userID, address).Order("created_time DESC").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

// GetTransactionsByCurrency 根据货币获取交易记录
func GetTransactionsByCurrency(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currencySymbol := c.Param("currency")
	chainType := c.Query("chain_type")

	query := database.GetDB().Where("user_id = ? AND currency_symbol = ?", userID, currencySymbol)
	if chainType != "" {
		query = query.Where("chain_type = ?", chainType)
	}

	var transactions []models.ChainBill
	if err := query.Order("created_time DESC").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}
