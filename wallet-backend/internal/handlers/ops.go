package handlers

import (
	"net/http"
	"wallet-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// OpsHandler operational endpoints for scanner and collection
type OpsHandler struct {
	Scanner    *services.BlockScannerService
	Collector  *services.CollectionService
}

func NewOpsHandler(scanner *services.BlockScannerService, collector *services.CollectionService) *OpsHandler {
	return &OpsHandler{Scanner: scanner, Collector: collector}
}

// POST /ops/scanner/start
func (h *OpsHandler) StartScanner(c *gin.Context) {
	if err := h.Scanner.StartScanning(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"running": true}})
}

// POST /ops/scanner/stop
func (h *OpsHandler) StopScanner(c *gin.Context) {
	h.Scanner.StopScanning()
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"running": false}})
}

// POST /ops/scanner/scan-once
func (h *OpsHandler) ScanOnce(c *gin.Context) {
	if err := h.Scanner.ScanOnce(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "ok"})
}

// GET /ops/scanner/status
func (h *OpsHandler) ScannerStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"running": h.Scanner.Status()}})
}

// POST /ops/collection/start
func (h *OpsHandler) StartCollection(c *gin.Context) {
	go h.Collector.StartCollection()
	c.JSON(http.StatusOK, gin.H{"data": "started"})
}

// POST /ops/collection/stop
func (h *OpsHandler) StopCollection(c *gin.Context) {
	h.Collector.Stop()
	c.JSON(http.StatusOK, gin.H{"data": "stopped"})
}

// POST /ops/collection/trigger
func (h *OpsHandler) TriggerCollection(c *gin.Context) {
	if err := h.Collector.TriggerOnce(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "ok"})
} 