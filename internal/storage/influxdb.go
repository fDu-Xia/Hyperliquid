package storage

import (
	"log"
	"strconv"
	"time"

	"HyperliquidNodeMonitor/internal/config"
	"HyperliquidNodeMonitor/internal/models"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// InfluxDBStorage handles connections and operations with InfluxDB
type InfluxDBStorage struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	config   *config.Config
}

// NewInfluxDBStorage creates a new InfluxDB storage handler
func NewInfluxDBStorage(cfg *config.Config) *InfluxDBStorage {
	client := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	writeAPI := client.WriteAPI(cfg.InfluxDBOrg, cfg.InfluxDBBucket)

	return &InfluxDBStorage{
		client:   client,
		writeAPI: writeAPI,
		config:   cfg,
	}
}

// Close closes the InfluxDB client
func (s *InfluxDBStorage) Close() {
	s.client.Close()
}

// Flush ensures all data is written to InfluxDB
func (s *InfluxDBStorage) Flush() {
	s.writeAPI.Flush()
}

// StoreTrade stores a trade record in InfluxDB
func (s *InfluxDBStorage) StoreTrade(trade *models.Trade) error {
	// Parse time
	timeWithZone := trade.Time + "Z"
	t, err := time.Parse(time.RFC3339Nano, timeWithZone)
	if err != nil {
		log.Printf("Failed to parse time: %v", err)
		return err
	}

	// Parse price and size
	price, _ := strconv.ParseFloat(trade.Price, 64)
	size, _ := strconv.ParseFloat(trade.Size, 64)

	// Create point data
	p := influxdb2.NewPoint(
		"trades",
		map[string]string{
			"coin":               trade.Coin,
			"side":               trade.Side,
			"hash":               trade.Hash,
			"trade_dir_override": trade.TradeDirections,
		},
		map[string]interface{}{
			"price":      price,
			"size":       size,
			"user_count": len(trade.SideInfo),
		},
		t,
	)

	// Add trade party information
	for i, info := range trade.SideInfo {
		startPos, _ := strconv.ParseFloat(info.StartPos, 64)

		// Create a separate point for each trading party
		userPoint := influxdb2.NewPoint(
			"trade_side_info",
			map[string]string{
				"coin": trade.Coin,
			},
			map[string]interface{}{
				"side":       trade.Side,
				"hash":       trade.Hash,
				"user":       info.User,
				"order_id":   strconv.FormatInt(info.OrderID, 10),
				"start_pos":  startPos,
				"has_twap":   info.TwapID != nil,
				"has_cloid":  info.ClOrderID != nil,
				"side_index": i,
			},
			t,
		)
		s.writeAPI.WritePoint(userPoint)
	}

	s.writeAPI.WritePoint(p)
	return nil
}

// StoreOrderStatus stores an order status in InfluxDB
func (s *InfluxDBStorage) StoreOrderStatus(orderStatus *models.OrderStatus) error {
	// Parse time
	timeWithZone := orderStatus.Time + "Z"
	t, err := time.Parse(time.RFC3339Nano, timeWithZone)
	if err != nil {
		log.Printf("Failed to parse time: %v", err)
		return err
	}

	// Parse price and size
	limitPrice, _ := strconv.ParseFloat(orderStatus.Order.LimitPrice, 64)
	size, _ := strconv.ParseFloat(orderStatus.Order.Size, 64)
	origSize, _ := strconv.ParseFloat(orderStatus.Order.OrigSize, 64)
	triggerPrice, _ := strconv.ParseFloat(orderStatus.Order.TriggerPrice, 64)

	// Create point data
	p := influxdb2.NewPoint(
		"order_statuses",
		map[string]string{
			"coin": orderStatus.Order.Coin,
		},
		map[string]interface{}{
			"user":             orderStatus.User,
			"status":           orderStatus.Status,
			"side":             orderStatus.Order.Side,
			"order_type":       orderStatus.Order.OrderType,
			"time_in_force":    orderStatus.Order.TimeInForce,
			"order_id":         orderStatus.Order.OrderID,
			"limit_price":      limitPrice,
			"size":             size,
			"orig_size":        origSize,
			"timestamp":        orderStatus.Order.Timestamp,
			"is_trigger":       orderStatus.Order.IsTrigger,
			"trigger_price":    triggerPrice,
			"is_position_tpsl": orderStatus.Order.IsPositionTpsl,
			"reduce_only":      orderStatus.Order.ReduceOnly,
			"children":         orderStatus.Order.Children,
			"cloid":            orderStatus.Order.ClOrderID,
		},
		t,
	)

	s.writeAPI.WritePoint(p)
	return nil
}
