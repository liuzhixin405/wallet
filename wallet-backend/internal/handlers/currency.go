package handlers

import (
	"net/http"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetCurrencies 获取所有币种配置
func GetCurrencies(c *gin.Context) {
	var currencies []models.CurrencyChainConfig
	if err := database.DB.Find(&currencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get currencies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": currencies})
}

// GetCurrencyBySymbol 根据符号获取币种配置
func GetCurrencyBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	var currency models.CurrencyChainConfig
	if err := database.DB.Where("symbol = ?", symbol).First(&currency).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Currency not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": currency})
}

// CreateCurrency 创建币种配置
func CreateCurrency(c *gin.Context) {
	var currency models.CurrencyChainConfig
	if err := c.ShouldBindJSON(&currency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := database.DB.Create(&currency).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create currency"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": currency})
}

// UpdateCurrency 更新币种配置
func UpdateCurrency(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	var currency models.CurrencyChainConfig
	if err := database.DB.Where("symbol = ?", symbol).First(&currency).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Currency not found"})
		return
	}

	var updateData models.CurrencyChainConfig
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := database.DB.Model(&currency).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update currency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": currency})
}

// DeleteCurrency 删除币种配置
func DeleteCurrency(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	if err := database.DB.Where("symbol = ?", symbol).Delete(&models.CurrencyChainConfig{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete currency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Currency deleted successfully"})
}

// EnableCurrency 启用币种
func EnableCurrency(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	if err := database.DB.Model(&models.CurrencyChainConfig{}).Where("symbol = ?", symbol).Update("is_enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable currency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Currency enabled successfully"})
}

// DisableCurrency 禁用币种
func DisableCurrency(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	if err := database.DB.Model(&models.CurrencyChainConfig{}).Where("symbol = ?", symbol).Update("is_enabled", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable currency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Currency disabled successfully"})
}

// GetSupportedChains 获取支持的链类型
func GetSupportedChains(c *gin.Context) {
	chains := []string{
		"Ethereum",
		"BSC",
		"Polygon",
		"Arbitrum",
		"Optimism",
	}

	c.JSON(http.StatusOK, gin.H{"data": chains})
}
