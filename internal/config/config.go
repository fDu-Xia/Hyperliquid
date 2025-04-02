package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file %v", err)
	}
}

// Config holds application configuration parameters
type Config struct {
	InfluxDBURL    string
	InfluxDBToken  string
	InfluxDBOrg    string
	InfluxDBBucket string
	TradesBasePath string
	OrdersBasePath string
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		InfluxDBURL:    "http://localhost:8086",
		InfluxDBToken:  "71OmLJvT2hdHhJYOkOl-apGBjleoog75OEWnjT8s292v0BVTjw-GMYnT0o1IipYOWXnoHvQy18BZ73PossmXQg==",
		InfluxDBOrg:    "hyperliquid",
		InfluxDBBucket: "hyperliquid",
		TradesBasePath: filepath.Join(os.Getenv("HOME"), "hl/data/node_trades/hourly"),
		OrdersBasePath: filepath.Join(os.Getenv("HOME"), "hl/data/node_order_statuses/hourly"),
	}
}

func NewConfig() *Config {
	return &Config{
		InfluxDBURL:    os.Getenv("INFLUXDB_URL"),
		InfluxDBToken:  os.Getenv("INFLUXDB_TOKEN"),
		InfluxDBOrg:    os.Getenv("INFLUXDB_ORG"),
		InfluxDBBucket: os.Getenv("INFLUXDB_BUCKET"),
		TradesBasePath: os.Getenv("TRADES_BASE_PATH"),
		OrdersBasePath: os.Getenv("ORDERS_BASE_PATH"),
	}
}
