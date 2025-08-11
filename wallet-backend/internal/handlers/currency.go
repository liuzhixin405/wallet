package handlers

import (
	"net/http"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetCurrencies 获取所有支持的货币配置
func GetCurrencies(c *gin.Context) {
	var currencies []models.CurrencyChainConfig
	if err := database.GetDB().Where("status = ?", true).Find(&currencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch currencies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": currencies})
}

// GetCurrencyBySymbol 根据货币符号获取配置
func GetCurrencyBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	chainType := c.Query("chain_type")

	query := database.GetDB().Where("currency_symbol = ? AND status = ?", symbol, true)
	if chainType != "" {
		query = query.Where("chain_type = ?", chainType)
	}

	var currencies []models.CurrencyChainConfig
	if err := query.Find(&currencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch currency config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": currencies})
}

// GetSupportedChains 获取支持的链类型
func GetSupportedChains(c *gin.Context) {
	var chains []string
	if err := database.GetDB().Model(&models.CurrencyChainConfig{}).
		Where("status = ?", true).
		Distinct().
		Pluck("chain_type", &chains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch supported chains"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": chains})
}
