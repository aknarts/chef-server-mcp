package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds environment configuration for the MCP server.
// ChefServerURL should be the root Chef server URL WITHOUT an organization segment.
// Organizations are supplied via CHEF_ORG_ALIASES env var as comma-separated alias=org pairs (e.g. "qa=qa1,prod=fireamp_classic").
// Optional CHEF_DEFAULT_ORG can set the default alias when multiple aliases are configured.
type Config struct {
	ChefUser        string
	ChefKeyPath     string
	ChefServerURL   string            // root URL (no /organizations/<org>)
	OrgAliases      map[string]string // alias -> real organization name
	DefaultOrgAlias string            // optional default alias
	Debug           bool              // enable verbose debug logging (CHEF_DEBUG=1/true/yes)
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() *Config {
	cfg := &Config{
		ChefUser:      os.Getenv("CHEF_USER"),
		ChefKeyPath:   os.Getenv("CHEF_KEY_PATH"),
		ChefServerURL: strings.TrimSuffix(os.Getenv("CHEF_SERVER_URL"), "/"),
		OrgAliases:    map[string]string{},
	}

	aliasSpec := os.Getenv("CHEF_ORG_ALIASES")
	if aliasSpec != "" {
		cfg.OrgAliases = ParseOrgAliases(aliasSpec)
	}
	cfg.DefaultOrgAlias = os.Getenv("CHEF_DEFAULT_ORG")

	if v := strings.ToLower(strings.TrimSpace(os.Getenv("CHEF_DEBUG"))); v == "1" || v == "true" || v == "yes" {
		cfg.Debug = true
	}

	return cfg
}

// ParseOrgAliases parses a comma-separated alias=org specification.
// Whitespace around aliases or orgs is trimmed. Empty entries are ignored.
func ParseOrgAliases(spec string) map[string]string {
	res := make(map[string]string)
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue // silently ignore malformed segment
		}
		alias := strings.TrimSpace(kv[0])
		org := strings.TrimSpace(kv[1])
		if alias == "" || org == "" {
			continue
		}
		res[alias] = org
	}
	return res
}

// Validate performs basic sanity checks and returns an error if misconfigured.
func (c *Config) Validate() error {
	if c.ChefUser == "" || c.ChefKeyPath == "" || c.ChefServerURL == "" {
		return fmt.Errorf("missing one of CHEF_USER, CHEF_KEY_PATH, CHEF_SERVER_URL")
	}
	if strings.Contains(c.ChefServerURL, "/organizations/") {
		return fmt.Errorf("CHEF_SERVER_URL must not include an organization path; provide root URL only and use CHEF_ORG_ALIASES for per-org access")
	}
	if len(c.OrgAliases) == 0 {
		// Single implicit org usage still requires specifying an alias mapping for clarity.
		// We allow zero aliases meaning user must always specify an org later if multi support demanded.
		return nil
	}
	if c.DefaultOrgAlias != "" {
		if _, ok := c.OrgAliases[c.DefaultOrgAlias]; !ok {
			return fmt.Errorf("CHEF_DEFAULT_ORG '%s' not present in CHEF_ORG_ALIASES", c.DefaultOrgAlias)
		}
	} else if len(c.OrgAliases) == 1 {
		// auto-assign sole alias as default
		for a := range c.OrgAliases {
			c.DefaultOrgAlias = a
		}
	}
	return nil
}
