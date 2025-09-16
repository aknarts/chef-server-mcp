package config

import (
	"os"
	"strings"
)

type Config struct {
	Port          string
	ChefUser      string
	ChefKeyPath   string
	ChefServerURL string
	KnifeFallback bool
}

func LoadFromEnv() *Config {
	cfg := &Config{
		Port:          getenvDefault("PORT", "8080"),
		ChefUser:      os.Getenv("CHEF_USER"),
		ChefKeyPath:   os.Getenv("CHEF_KEY_PATH"),
		ChefServerURL: os.Getenv("CHEF_SERVER_URL"),
		KnifeFallback: parseBoolDefault(os.Getenv("KNIFE_FALLBACK"), true),
	}
	return cfg
}

func getenvDefault(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func parseBoolDefault(v string, d bool) bool {
	if v == "" {
		return d
	}
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
