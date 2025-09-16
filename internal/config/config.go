package config

import "os"

// Config holds environment configuration for the MCP server.
// Knife fallback removed; all operations require Chef API credentials.
type Config struct {
	ChefUser      string
	ChefKeyPath   string
	ChefServerURL string
}

func LoadFromEnv() *Config {
	return &Config{
		ChefUser:      os.Getenv("CHEF_USER"),
		ChefKeyPath:   os.Getenv("CHEF_KEY_PATH"),
		ChefServerURL: os.Getenv("CHEF_SERVER_URL"),
	}
}
