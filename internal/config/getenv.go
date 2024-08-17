package config

import (
	"os"
	"strconv"
)

func GetEnv[T any](key string, fallback T) T {
	if value, exists := os.LookupEnv(key); exists {
		switch any(value).(type) {
		case string:
			return any(value).(T)
		case int:
			if valueAsInt, err := strconv.Atoi(value); err == nil {
				return any(valueAsInt).(T)
			}
		default:
			return fallback
		}
	}
	return fallback
}
