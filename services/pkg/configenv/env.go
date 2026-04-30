package configenv

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func String(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	val = strings.TrimSpace(val)
	if val == "" {
		return fallback
	}

	return val
}

func Int(key string, fallback int) (int, error) {
	valStr, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	valStr = strings.TrimSpace(valStr)
	if valStr == "" {
		return fallback, nil
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %w", key, err)
	}

	return val, nil
}

func Bool(key string, fallback bool) (bool, error) {
	valStr, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	valStr = strings.TrimSpace(valStr)
	if valStr == "" {
		return fallback, nil
	}

	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value for %s: %w", key, err)
	}

	return val, nil
}

func RequiredString(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("missing required environment variable %s", key)
	}

	val = strings.TrimSpace(val)
	if val == "" {
		return "", fmt.Errorf("missing required environment variable %s", key)
	}

	return val, nil
}
