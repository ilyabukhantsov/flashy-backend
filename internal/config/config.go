// Package config читає налаштування з змінних середовища.
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr             string
	SessionTTL       time.Duration
	ReminderInterval time.Duration
}

func Load() Config {
	return Config{
		Addr:             getenv("ADDR", ":8080"),
		SessionTTL:       time.Duration(getint("SESSION_TTL_HOURS", 720)) * time.Hour,
		ReminderInterval: time.Duration(getint("REMINDER_INTERVAL_MINUTES", 60)) * time.Minute,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return def
}
