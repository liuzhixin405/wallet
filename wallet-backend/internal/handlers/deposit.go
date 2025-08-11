package handlers

import (
	"net/http"
	"wallet-backend/internal/database"
	"wallet-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetDeposits 获取充值记录
func GetDeposits(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var deposits []models.DepositRecord
	if err := database.GetDB().Where("userid = ?", userID).Order("created_time DESC").Find(&deposits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": deposits})
}

// GetDepositByID 获取指定充值记录
func GetDepositByID(c *gin.Context) {
	userID, _ := c.Get("user_id")
	depositID := c.Param("id")

	var deposit models.DepositRecord
	if err := database.GetDB().Where("id = ? AND userid = ?", depositID, userID).First(&deposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": deposit})
}

// GetDepositsByAddress 根据地址获取充值记录
func GetDepositsByAddress(c *gin.Context) {
	userID, _ := c.Get("user_id")
	address := c.Param("address")

	var deposits []models.DepositRecord
	if err := database.GetDB().Where("userid = ? AND to_address = ?", userID, address).Order("created_time DESC").Find(&deposits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": deposits})
}
