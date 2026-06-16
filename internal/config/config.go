package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	AppID       int
	AppHash     string
	Port        string
	DBPath      string
	SessionDir  string
	Workers     int
	TimeoutMins int
	Debug       bool
	CORSOrigin  string
}

func NewConfig() (*Config, error) {
	appIDStr, err := Getenv("APP_ID", "", true)
	if err != nil {
		return nil, err
	}
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return nil, errors.New("APP_ID butun son bo'lishi kerak")
	}

	appHash, err := Getenv("APP_HASH", "", true)
	if err != nil {
		return nil, err
	}

	workersStr, err := Getenv("WORKERS", "10", false)
	if err != nil {
		return nil, err
	}
	workers, err := strconv.Atoi(workersStr)
	if err != nil {
		return nil, err
	}

	timeoutStr, err := Getenv("TRANSACTION_TIMEOUT", "30", false)
	if err != nil {
		return nil, err
	}
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		AppID:       appID,
		AppHash:     appHash,
		Port:        GetenvValue("PORT", "8080"),
		DBPath:      GetenvValue("DB_PATH", "./db.sqlite3"),
		SessionDir:  GetenvValue("SESSION_DIR", "sessions"),
		Workers:     workers,
		TimeoutMins: timeout,
		Debug:       os.Getenv("DEBUG") == "true",
		CORSOrigin:  GetenvValue("CORS_ORIGIN", "*"),
	}, nil
}

func Getenv(key string, def string, is_required bool) (string, error) {
	value := os.Getenv(key)
	if is_required && value == "" {
		return "", errors.New(key + " is not set .env")
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
