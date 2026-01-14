package config

import (
	"log"
	"os"
	"strconv"
)

func GetEnv(key, value string) string {
	envVal := os.Getenv(key)
	if envVal == "" {
		return value
	}
	return envVal
}

func GetEnvInt(key string, value int) int {
	envVal := os.Getenv(key)
	if envVal == "" {
		return value
	}
	val, err := strconv.Atoi(envVal)
	if err != nil {
		log.Printf("Invalid integer for %s: %s, using default %d", key, envVal, value)
		return value
	}
	return val
}
