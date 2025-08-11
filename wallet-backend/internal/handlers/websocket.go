package handlers

import (
	"net/http"
	"wallet-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	wsService *services.WebSocketService
}

// NewWebSocketHandler 创建新的WebSocket处理器
func NewWebSocketHandler(wsService *services.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		wsService: wsService,
	}
}

// HandleWebSocket WebSocket连接处理
func (wh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 从JWT token中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 转换为uint64
	userIDUint, ok := userID.(uint64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// 处理WebSocket连接
	wh.wsService.HandleWebSocket(c.Writer, c.Request, userIDUint)
}

// GetWebSocketStats 获取WebSocket统计信息
func (wh *WebSocketHandler) GetWebSocketStats(c *gin.Context) {
	stats := gin.H{
		"total_clients":      wh.wsService.GetConnectedClientsCount(),
		"active_connections": wh.wsService.GetConnectedClientsCount(),
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
