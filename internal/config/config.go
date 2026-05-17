package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBHost       string
	DBPort       int
	DBName       string
	DBUser       string
	DBPass       string
	RedisAddr    string
	CacheTTL     time.Duration
	SessionTTL   time.Duration
	SQLLogQuery      bool
	SQLLogTime       bool
	SQLSlowThreshold time.Duration
}

func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	slowMS, err := strconv.Atoi(getEnv("SQL_SLOW_THRESHOLD_MS", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid SQL_SLOW_THRESHOLD_MS: %w", err)
	}

	cacheTTL, err := time.ParseDuration(getEnv("CACHE_TTL", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid CACHE_TTL: %w", err)
	}

	sessionTTL, err := time.ParseDuration(getEnv("SESSION_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TTL: %w", err)
	}

	return &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      dbPort,
		DBName:      mustEnv("DB_NAME"),
		DBUser:      mustEnv("DB_USER"),
		DBPass:      mustEnv("DB_PASS"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		CacheTTL:    cacheTTL,
		SessionTTL:  sessionTTL,
		SQLLogQuery:      parseBool(getEnv("SQL_LOG_QUERY", "false")),
		SQLLogTime:       parseBool(getEnv("SQL_LOG_TIME", "false")),
		SQLSlowThreshold: time.Duration(slowMS) * time.Millisecond,
	}, nil
}

func parseBool(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env variable %s is not set", key))
	}
	return v
}
