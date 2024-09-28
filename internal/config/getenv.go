package config

import (
	"os"
	"strconv"
)

func GetEnv(key string, fallback any) any {
	if value, exists := os.LookupEnv(key); exists {
		switch any(value).(type) {
		case string:
			return value
		case int:
			if valueAsInt, err := strconv.Atoi(value); err == nil {
				return valueAsInt
			}
		case bool:
			if valueAsBool, err := strconv.ParseBool(value); err == nil {
				return valueAsBool
			}
		default:
			return fallback
		}
	}
	return fallback
}
