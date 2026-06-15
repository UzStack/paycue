package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Workers    int
	WebhookURL string
	Port       string
	WatchID    int64
	Limit      int64
	APIKey     string
}

func NewConfig() (*Config, error) {
	workers_str, err := Getenv("WORKERS", "10", false)
	if err != nil {
		return nil, err
	}
	workers, err := strconv.Atoi(workers_str)
	if err != nil {
		return nil, err
	}

	watchStr, err := Getenv("WATCH_ID", "", true)
	if err != nil {
		return nil, err
	}
	watchID, err := strconv.ParseInt(watchStr, 10, 64)
	if err != nil {
		return nil, err
	}

	limitStr, err := Getenv("LIMIT", "100", false)
	if err != nil {
		return nil, err
	}
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return nil, err
	}

	apiKey, err := Getenv("API_KEY", "", true)
	if err != nil {
		return nil, err
	}

	return &Config{
		Workers:    workers,
		WebhookURL: GetenvValue("WEBHOOK_URL", ""),
		Port:       GetenvValue("PORT", "8080"),
		WatchID:    watchID,
		Limit:      limit,
		APIKey:     apiKey,
	}, nil
}

func Getenv(key string, def string, is_required bool) (string, error) {
	value := os.Getenv(key)
	if is_required && value == "" {
		return "", errors.New(key + "is not set .env")
	}
	if value == "" {
		return def, nil
	}
	return value, nil
}

func GetenvValue(key string, def string) string {
	value, err := Getenv(key, def, false)
	if err != nil || value == "" {
		return def
	}
	return value
}
