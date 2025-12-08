package config

import "os"

func GetEnv(key, value string) string {
	envVal := os.Getenv(key)
	if envVal == "" {
		return value
	}
	return envVal
}
