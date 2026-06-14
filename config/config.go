package config

import (
	"os"
	"strconv"

	"coupon-service/internal/constants"
)

type Config struct {
	DBPath        string
	Port          string
	TickerMinutes int
}

func Load() *Config {
	dbPath := os.Getenv("COUPON_DB_PATH")
	if dbPath == "" {
		dbPath = constants.DefaultDBPath
	}

	port := os.Getenv("COUPON_PORT")
	if port == "" {
		port = constants.DefaultPort
	}

	tickerMinutes := constants.DefaultTickerMinutes
	if tm := os.Getenv("COUPON_TICKER_MINUTES"); tm != "" {
		if v, err := strconv.Atoi(tm); err == nil && v > 0 {
			tickerMinutes = v
		}
	}

	return &Config{
		DBPath:        dbPath,
		Port:          port,
		TickerMinutes: tickerMinutes,
	}
}
