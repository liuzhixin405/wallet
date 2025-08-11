package handlers

import (
	"net/http"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetBalances 获取用户余额列表
func GetBalances(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var balances []models.Balance
	if err := database.GetDB().Where("address IN (SELECT address FROM address_library WHERE userid = ?)", userID).Find(&balances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch balances"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": balances})
}

// GetBalanceByCurrency 获取指定货币的余额
func GetBalanceByCurrency(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currencySymbol := c.Param("currency")
	chainType := c.Query("chain_type")

	query := database.GetDB().Where("address IN (SELECT address FROM address_library WHERE userid = ?)", userID)
	if currencySymbol != "" {
		query = query.Where("currency_symbol = ?", currencySymbol)
	}
	if chainType != "" {
		query = query.Where("chain_type = ?", chainType)
	}

	var balances []models.Balance
	if err := query.Find(&balances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": balances})
}
