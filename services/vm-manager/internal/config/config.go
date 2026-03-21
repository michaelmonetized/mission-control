package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port          int
	FlyAPIToken   string
	FlyApp        string
	MaxVMs        int
	AllowOrigin   string
	MetricsPath   string
	LogLevel      string
}

func Load() *Config {
	cfg := &Config{
		Port:        getEnvInt("VM_MANAGER_PORT", 8080),
		FlyAPIToken: getEnv("FLY_API_TOKEN", ""),
		FlyApp:      getEnv("FLY_APP", "mission-control-vms"),
		MaxVMs:      getEnvInt("MAX_VMS_PER_ORG", 100),
		AllowOrigin: getEnv("ALLOW_ORIGIN", "*"),
		MetricsPath: getEnv("METRICS_PATH", "/tmp/vm-manager-metrics.json"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	if cfg.FlyAPIToken == "" {
		fmt.Fprintf(os.Stderr, "⚠️  FLY_API_TOKEN not set; Fly.io operations will fail\n")
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
