package pkg

import (
	"os"
	"strings"
)

func GetEnv(key, defaultValue string) string {
	if data, err := os.ReadFile("/run/secrets/" + key); err == nil {
		return strings.TrimSpace(string(data))
	}
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
