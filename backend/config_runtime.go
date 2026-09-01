package backend

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	lookupEnv = os.Getenv
}

var lookupEnv = func(key string) string { return "" }

func getenv(key string) string {
	return strings.TrimSpace(lookupEnv(key))
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envInt64(key string, fallback int64) int64 {
	value := getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envString(key, fallback string) string {
	if value := getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
