package processor

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"HyperliquidNodeMonitor/internal/config"
	"HyperliquidNodeMonitor/internal/models"
	"HyperliquidNodeMonitor/internal/storage"

	"github.com/fsnotify/fsnotify"
)

// RawProcessor processes raw blockchain data files
type RawProcessor struct {
	config  *config.Config
	storage *storage.InfluxDBStorage
}

// NewRawProcessor creates a new raw data processor
func NewRawProcessor(cfg *config.Config, store *storage.InfluxDBStorage) *RawProcessor {
	return &RawProcessor{
		config:  cfg,
		storage: store,
	}
}

// ProcessExistingData processes all existing data files
func (p *RawProcessor) ProcessExistingData() {
	// Process trade data
	p.processDirData(p.config.TradesBasePath, p.processTrade)

	// Process order status data
	p.processDirData(p.config.OrdersBasePath, p.processOrderStatus)

	// Ensure data is written
	p.storage.Flush()
}

// SetupFileWatcher sets up a file watcher to process new files
func (p *RawProcessor) SetupFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	// Add directories to watch
	watchDirs := []string{p.config.TradesBasePath, p.config.OrdersBasePath}

	// Create done channel
	done := make(chan bool)

	// Asynchronously process watch events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					// If a new date directory was created, add it to the watch
					if p.isDateDir(filepath.Base(event.Name)) {
						watcher.Add(event.Name)
						log.Printf("Added new date directory to watch: %s", event.Name)
						continue
					}

					// Process newly created files
					if p.isDataFile(event.Name) {
						log.Printf("Found new data file: %s", event.Name)

						// Determine data type and process
						if strings.Contains(event.Name, "node_trades") {
							p.processFileWithRetry(event.Name, p.processTrade)
						} else if strings.Contains(event.Name, "node_order_statuses") {
							p.processFileWithRetry(event.Name, p.processOrderStatus)
						}

						// Ensure data is written
						p.storage.Flush()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("File watch error: %v", err)
			}
		}
	}()

	// Initialize watching directories and subdirectories
	for _, dir := range watchDirs {
		err = p.addDirAndSubdirs(watcher, dir)
		if err != nil {
			log.Printf("Failed to add directory to watch %s: %v", dir, err)
		}
	}

	<-done
}

// processDirData processes all data in a directory
func (p *RawProcessor) processDirData(basePath string, processFunc func(string)) {
	// Get date directories
	dateDirs, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	for _, dateDir := range dateDirs {
		if !dateDir.IsDir() || !p.isDateDir(dateDir.Name()) {
			continue
		}

		datePath := filepath.Join(basePath, dateDir.Name())
		hourDirs, err := os.ReadDir(datePath)
		if err != nil {
			log.Printf("Failed to read hour directory: %v", err)
			continue
		}

		for _, hourDir := range hourDirs {
			if !hourDir.IsDir() && !p.isHourFile(hourDir.Name()) {
				continue
			}

			var dataPath string
			if hourDir.IsDir() {
				// node_trades structure
				hourPath := filepath.Join(datePath, hourDir.Name())
				files, err := os.ReadDir(hourPath)
				if err != nil {
					log.Printf("Failed to read files: %v", err)
					continue
				}

				for _, file := range files {
					if file.IsDir() {
						continue
					}
					dataPath = filepath.Join(hourPath, file.Name())
					p.processFile(dataPath, processFunc)
				}
			} else {
				// node_order_statuses structure
				dataPath = filepath.Join(datePath, hourDir.Name())
				p.processFile(dataPath, processFunc)
			}
		}
	}
}

// processFile processes a single data file
func (p *RawProcessor) processFile(filePath string, processFunc func(string)) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", filePath, err)
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		processFunc(line)
	}
}

// processTrade processes a trade data line
func (p *RawProcessor) processTrade(line string) {
	var trade models.Trade
	if err := json.Unmarshal([]byte(line), &trade); err != nil {
		log.Printf("Failed to parse trade data: %v", err)
		return
	}

	err := p.storage.StoreTrade(&trade)
	if err != nil {
		log.Printf("Failed to store trade: %v", err)
	}
}

// processOrderStatus processes an order status data line
func (p *RawProcessor) processOrderStatus(line string) {
	var orderStatus models.OrderStatus
	if err := json.Unmarshal([]byte(line), &orderStatus); err != nil {
		log.Printf("Failed to parse order status data: %v", err)
		return
	}

	err := p.storage.StoreOrderStatus(&orderStatus)
	if err != nil {
		log.Printf("Failed to store order status: %v", err)
	}
}

// processFileWithRetry processes a file with retry logic
func (p *RawProcessor) processFileWithRetry(filePath string, processFunc func(string)) {
	maxRetries := 5
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("File does not exist, retrying after delay: %s", filePath)
			time.Sleep(retryDelay)
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read file, retrying after delay %s: %v", filePath, err)
			time.Sleep(retryDelay)
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			processFunc(line)
		}
		return
	}
	log.Printf("Failed to process file, maximum retry count reached: %s", filePath)
}

// addDirAndSubdirs adds a directory and its subdirectories to the watcher
func (p *RawProcessor) addDirAndSubdirs(watcher *fsnotify.Watcher, dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	return err
}

// Helper functions

// isDateDir checks if a name is a date directory
func (p *RawProcessor) isDateDir(name string) bool {
	// Date format: YYYYMMDD, e.g., 20250328
	if len(name) != 8 {
		return false
	}
	_, err := strconv.Atoi(name)
	return err == nil
}

// isHourFile checks if a name is an hour file
func (p *RawProcessor) isHourFile(name string) bool {
	// Hour format: H or HH, e.g., 9 or 10
	_, err := strconv.Atoi(name)
	return err == nil
}

// isDataFile checks if a path is a data file
func (p *RawProcessor) isDataFile(path string) bool {
	// Check if file is a data file
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
