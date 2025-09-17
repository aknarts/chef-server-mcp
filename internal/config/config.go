package config

import (
	"encoding/json"
	"os"
	"strings"
)

// Config holds environment configuration for the MCP server.
// Knife fallback removed; all operations require Chef API credentials.
type Config struct {
	ChefUser      string
	ChefKeyPath   string
	ChefServerURL string            // Base Chef server URL without organization
	DefaultOrg    string            // Default organization to use if none specified
	OrgAliases    map[string]string // Organization aliases mapping
}

func LoadFromEnv() *Config {
	cfg := &Config{
		ChefUser:      os.Getenv("CHEF_USER"),
		ChefKeyPath:   os.Getenv("CHEF_KEY_PATH"),
		ChefServerURL: os.Getenv("CHEF_SERVER_URL"),
		DefaultOrg:    os.Getenv("CHEF_DEFAULT_ORG"),
		OrgAliases:    make(map[string]string),
	}

	// Backward compatibility: if CHEF_SERVER_URL includes "/organizations/<org>",
	// extract the org and set it as DefaultOrg (if not already set), and trim the base URL.
	// Examples:
	//   https://chef.example.com/organizations/acme -> base=https://chef.example.com, org=acme
	//   https://chef.example.com/organizations/acme/ -> base=https://chef.example.com, org=acme
	if idx := strings.Index(cfg.ChefServerURL, "/organizations/"); idx != -1 {
		base := strings.TrimRight(cfg.ChefServerURL[:idx], "/")
		rest := cfg.ChefServerURL[idx+len("/organizations/"):]
		// rest may contain trailing slash or further path; take first segment as org
		if slash := strings.Index(rest, "/"); slash != -1 {
			rest = rest[:slash]
		}
		org := strings.TrimSpace(rest)
		if org != "" {
			if cfg.DefaultOrg == "" {
				cfg.DefaultOrg = org
			}
			cfg.ChefServerURL = base
		}
	}

	// Load organization aliases from environment variable (JSON format)
	if aliasesJSON := os.Getenv("CHEF_ORG_ALIASES"); aliasesJSON != "" {
		if err := json.Unmarshal([]byte(aliasesJSON), &cfg.OrgAliases); err != nil {
			// If JSON parsing fails, try simple key=value format
			cfg.OrgAliases = parseSimpleAliases(aliasesJSON)
		}
	}

	return cfg
}

// parseSimpleAliases parses aliases in format "alias1=org1,alias2=org2"
func parseSimpleAliases(aliasStr string) map[string]string {
	aliases := make(map[string]string)
	pairs := strings.Split(aliasStr, ",")
	for _, pair := range pairs {
		if kv := strings.SplitN(strings.TrimSpace(pair), "=", 2); len(kv) == 2 {
			aliases[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return aliases
}

// ResolveOrganization resolves an organization name or alias to the actual organization name
func (c *Config) ResolveOrganization(orgInput string) string {
	if orgInput == "" {
		return c.DefaultOrg
	}

	// Check if it's an alias
	if actualOrg, exists := c.OrgAliases[orgInput]; exists {
		return actualOrg
	}

	// Return as-is if not an alias
	return orgInput
}
