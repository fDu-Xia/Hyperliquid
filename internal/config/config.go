package config

import (
	"os"
	"path/filepath"
)

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
