package configenv

import (
	"fmt"
	"os"
	"strconv"
)

func String(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}

	return val
}

func Int(key string, fallback int) (int, error) {
	valStr, ok := os.LookupEnv(key)
	if !ok || valStr == "" {
		return fallback, nil
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %w", key, err)
	}

	return val, nil
}

func RequiredString(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return "", fmt.Errorf("missing required environment variable %s", key)
	}

	return val, nil
}
