package handlers

import (
	"net/http"
	"strings"
	"wallet-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ToolsHandler 工具管理处理器
type ToolsHandler struct {
	Scanner    *services.BlockScannerService
	Collector  *services.CollectionService
}

// NewToolsHandler 创建新的工具处理器
func NewToolsHandler(scanner *services.BlockScannerService, collector *services.CollectionService) *ToolsHandler {
	return &ToolsHandler{
		Scanner:   scanner,
		Collector: collector,
	}
}

// TriggerCollection 触发归集
func (h *ToolsHandler) TriggerCollection(c *gin.Context) {
	var req struct {
		Symbol  string `json:"symbol" binding:"required"`
		Address string `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 验证地址格式
	if !isValidAddress(req.Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address format"})
		return
	}

	// 执行归集
	if err := h.Collector.CollectFromAddress(req.Symbol, req.Address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Collection triggered successfully"})
}

// ScanBlocks 扫描指定区块范围
func (h *ToolsHandler) ScanBlocks(c *gin.Context) {
	var req struct {
		Symbol    string   `json:"symbol" binding:"required"`
		StartBlock uint64  `json:"start_block" binding:"required"`
		EndBlock   uint64  `json:"end_block" binding:"required"`
		Addresses  []string `json:"addresses" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 验证区块范围
	if req.StartBlock >= req.EndBlock {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start block must be less than end block"})
		return
	}

	// 验证地址列表
	if len(req.Addresses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one address is required"})
		return
	}

	for _, addr := range req.Addresses {
		if !isValidAddress(addr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address format: " + addr})
			return
		}
	}

	// 执行区块扫描
	if err := h.Scanner.ScanBlocks(req.Symbol, req.StartBlock, req.EndBlock, req.Addresses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Block scanning triggered successfully"})
}

// GetBalances 获取指定币种的余额
func (h *ToolsHandler) GetBalances(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	// 这里需要实现获取指定币种余额的逻辑
	// 暂时返回空数组
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// GetAddresses 获取指定币种的地址列表
func (h *ToolsHandler) GetAddresses(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	// 这里需要实现获取指定币种地址列表的逻辑
	// 暂时返回空数组
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// isValidAddress 验证地址格式
func isValidAddress(address string) bool {
	// 基本的地址格式验证
	if len(address) < 26 || len(address) > 50 {
		return false
	}
	
	// 以太坊地址格式验证
	if strings.HasPrefix(address, "0x") && len(address) == 42 {
		return true
	}
	
	// 其他地址格式可以在这里添加
	return true
} 