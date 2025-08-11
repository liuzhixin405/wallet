package services

import (
	"log"
	"time"
	"wallet-backend/internal/config"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	config           *config.Config
	blockScanner     *BlockScannerService
	collectionService *CollectionService
	stopChan         chan bool
}

// NewSchedulerService 创建新的定时任务服务
func NewSchedulerService(cfg *config.Config, scanner *BlockScannerService, collector *CollectionService) *SchedulerService {
	return &SchedulerService{
		config:           cfg,
		blockScanner:     scanner,
		collectionService: collector,
		stopChan:         make(chan bool),
	}
}

// Start 启动定时任务服务
func (ss *SchedulerService) Start() {
	log.Println("Starting scheduler service...")
	
	// 启动区块扫描任务
	go ss.startBlockScanningTask()
	
	// 启动归集任务
	go ss.startCollectionTask()
	
	// 启动其他定时任务
	go ss.startOtherTasks()
}

// Stop 停止定时任务服务
func (ss *SchedulerService) Stop() {
	log.Println("Stopping scheduler service...")
	ss.stopChan <- true
}

// startBlockScanningTask 启动区块扫描定时任务
func (ss *SchedulerService) startBlockScanningTask() {
	// 参考钱包控制台的实现，不同币种有不同的扫描间隔
	tickers := map[string]*time.Ticker{
		"ETH":    time.NewTicker(500 * time.Millisecond),  // 500ms
		"USDT":   time.NewTicker(500 * time.Millisecond),  // 500ms
		"USDC":   time.NewTicker(500 * time.Millisecond),  // 500ms
		"BNB":    time.NewTicker(2 * time.Second),         // 2s
		"BUSD":   time.NewTicker(2 * time.Second),         // 2s
		"default": time.NewTicker(10 * time.Second),       // 10s
	}
	
	defer func() {
		for _, ticker := range tickers {
			ticker.Stop()
		}
	}()
	
	// 启动区块扫描服务
	if err := ss.blockScanner.StartScanning(); err != nil {
		log.Printf("Failed to start block scanning: %v", err)
	}
	
	// 监听停止信号
	<-ss.stopChan
	ss.blockScanner.StopScanning()
}

// startCollectionTask 启动归集定时任务
func (ss *SchedulerService) startCollectionTask() {
	// 归集任务间隔，参考钱包控制台
	ticker := time.NewTicker(5 * time.Minute) // 5分钟检查一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ss.stopChan:
			log.Println("Collection task stopped")
			return
		case <-ticker.C:
			if err := ss.collectionService.TriggerOnce(); err != nil {
				log.Printf("Collection task error: %v", err)
			}
		}
	}
}

// startOtherTasks 启动其他定时任务
func (ss *SchedulerService) startOtherTasks() {
	// 地址生成任务 - 每30秒检查一次
	addressTicker := time.NewTicker(30 * time.Second)
	defer addressTicker.Stop()
	
	// 交易确认任务 - 每2秒检查一次
	confirmTicker := time.NewTicker(2 * time.Second)
	defer confirmTicker.Stop()
	
	// 数据上传任务 - 每5秒检查一次
	uploadTicker := time.NewTicker(5 * time.Second)
	defer uploadTicker.Stop()
	
	for {
		select {
		case <-ss.stopChan:
			log.Println("Other tasks stopped")
			return
		case <-addressTicker.C:
			ss.processAddressGeneration()
		case <-confirmTicker.C:
			ss.processTransactionConfirmation()
		case <-uploadTicker.C:
			ss.processDataUpload()
		}
	}
}

// processAddressGeneration 处理地址生成任务
func (ss *SchedulerService) processAddressGeneration() {
	// TODO: 实现地址生成逻辑
	// 参考钱包控制台的 runNewAddress 方法
	log.Println("Processing address generation...")
}

// processTransactionConfirmation 处理交易确认任务
func (ss *SchedulerService) processTransactionConfirmation() {
	// TODO: 实现交易确认逻辑
	// 参考钱包控制台的 runUploadConfirmData 方法
	log.Println("Processing transaction confirmation...")
}

// processDataUpload 处理数据上传任务
func (ss *SchedulerService) processDataUpload() {
	// TODO: 实现数据上传逻辑
	// 参考钱包控制台的 runUploadChargeData 方法
	log.Println("Processing data upload...")
} 