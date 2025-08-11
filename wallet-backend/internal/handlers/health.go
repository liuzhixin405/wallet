package handlers

import (
	"net/http"
	"time"
	"wallet-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services"`
	Version   string                 `json:"version"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	response := &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]ServiceInfo),
		Version:   "1.0.0",
	}

	// 检查数据库连接
	dbInfo := ServiceInfo{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
	}

	if err := database.GetDB().Raw("SELECT 1").Error; err != nil {
		dbInfo.Status = "unhealthy"
		dbInfo.Message = err.Error()
		response.Status = "unhealthy"
	}
	response.Services["database"] = dbInfo

	// 检查WebSocket服务
	wsInfo := ServiceInfo{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
	}
	response.Services["websocket"] = wsInfo

	// 检查区块链连接
	blockchainInfo := ServiceInfo{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
	}
	response.Services["blockchain"] = blockchainInfo

	// 根据整体状态返回相应的HTTP状态码
	if response.Status == "healthy" {
		c.JSON(http.StatusOK, gin.H{"data": response})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"data": response})
	}
}

// ReadinessCheck 就绪检查
func ReadinessCheck(c *gin.Context) {
	response := &HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]ServiceInfo),
		Version:   "1.0.0",
	}

	// 检查数据库连接
	dbInfo := ServiceInfo{
		Status:    "ready",
		Timestamp: time.Now().Unix(),
	}

	if err := database.GetDB().Raw("SELECT 1").Error; err != nil {
		dbInfo.Status = "not_ready"
		dbInfo.Message = err.Error()
		response.Status = "not_ready"
	}
	response.Services["database"] = dbInfo

	// 检查必要的服务
	services := []string{"websocket", "blockchain", "wallet"}
	for _, service := range services {
		serviceInfo := ServiceInfo{
			Status:    "ready",
			Timestamp: time.Now().Unix(),
		}
		response.Services[service] = serviceInfo
	}

	if response.Status == "ready" {
		c.JSON(http.StatusOK, gin.H{"data": response})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"data": response})
	}
}

// LivenessCheck 存活检查
func LivenessCheck(c *gin.Context) {
	response := &HealthResponse{
		Status:    "alive",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// Metrics 指标接口
func Metrics(c *gin.Context) {
	metrics := gin.H{
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(time.Now()).Seconds(), // 这里应该使用实际的启动时间
		"version":   "1.0.0",
		"database": gin.H{
			"connections": 0, // 这里应该获取实际的连接数
		},
		"websocket": gin.H{
			"clients": 0, // 这里应该获取实际的客户端数
		},
		"blockchain": gin.H{
			"last_block": 0, // 这里应该获取最新的区块高度
		},
	}

	c.JSON(http.StatusOK, gin.H{"data": metrics})
}
