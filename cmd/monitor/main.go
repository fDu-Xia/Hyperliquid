package main

import (
	"HyperliquidNodeMonitor/internal/config"
	"HyperliquidNodeMonitor/internal/processor"
	"HyperliquidNodeMonitor/internal/storage"
)

func main() {
	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize InfluxDB storage
	store := storage.NewInfluxDBStorage(cfg)
	defer store.Close()

	// Initialize processor
	proc := processor.NewRawProcessor(cfg, store)

	// Process existing data
	proc.ProcessExistingData()

	// Set up file watcher for new data
	proc.SetupFileWatcher()
}
