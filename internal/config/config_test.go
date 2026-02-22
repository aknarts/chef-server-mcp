package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv_Basic(t *testing.T) {
	os.Setenv("CHEF_USER", "testuser")
	os.Setenv("CHEF_KEY_PATH", "/tmp/test.pem")
	os.Setenv("CHEF_SERVER_URL", "https://chef.example.com")
	os.Setenv("CHEF_DEFAULT_ORG", "acme")
	os.Setenv("CHEF_ORG_ALIASES", "")
	defer func() {
		os.Unsetenv("CHEF_USER")
		os.Unsetenv("CHEF_KEY_PATH")
		os.Unsetenv("CHEF_SERVER_URL")
		os.Unsetenv("CHEF_DEFAULT_ORG")
		os.Unsetenv("CHEF_ORG_ALIASES")
	}()

	cfg := LoadFromEnv()
	if cfg.ChefUser != "testuser" {
		t.Errorf("expected ChefUser=testuser, got %s", cfg.ChefUser)
	}
	if cfg.ChefKeyPath != "/tmp/test.pem" {
		t.Errorf("expected ChefKeyPath=/tmp/test.pem, got %s", cfg.ChefKeyPath)
	}
	if cfg.ChefServerURL != "https://chef.example.com" {
		t.Errorf("expected ChefServerURL=https://chef.example.com, got %s", cfg.ChefServerURL)
	}
	if cfg.DefaultOrg != "acme" {
		t.Errorf("expected DefaultOrg=acme, got %s", cfg.DefaultOrg)
	}
}

func TestLoadFromEnv_BackwardCompatURL(t *testing.T) {
	os.Setenv("CHEF_USER", "testuser")
	os.Setenv("CHEF_KEY_PATH", "/tmp/test.pem")
	os.Setenv("CHEF_SERVER_URL", "https://chef.example.com/organizations/acme")
	os.Setenv("CHEF_DEFAULT_ORG", "")
	os.Setenv("CHEF_ORG_ALIASES", "")
	defer func() {
		os.Unsetenv("CHEF_USER")
		os.Unsetenv("CHEF_KEY_PATH")
		os.Unsetenv("CHEF_SERVER_URL")
		os.Unsetenv("CHEF_DEFAULT_ORG")
		os.Unsetenv("CHEF_ORG_ALIASES")
	}()

	cfg := LoadFromEnv()
	if cfg.ChefServerURL != "https://chef.example.com" {
		t.Errorf("expected base URL without org, got %s", cfg.ChefServerURL)
	}
	if cfg.DefaultOrg != "acme" {
		t.Errorf("expected extracted org=acme, got %s", cfg.DefaultOrg)
	}
}

func TestLoadFromEnv_BackwardCompatURL_WithTrailingSlash(t *testing.T) {
	os.Setenv("CHEF_USER", "testuser")
	os.Setenv("CHEF_KEY_PATH", "/tmp/test.pem")
	os.Setenv("CHEF_SERVER_URL", "https://chef.example.com/organizations/acme/")
	os.Setenv("CHEF_DEFAULT_ORG", "")
	os.Setenv("CHEF_ORG_ALIASES", "")
	defer func() {
		os.Unsetenv("CHEF_USER")
		os.Unsetenv("CHEF_KEY_PATH")
		os.Unsetenv("CHEF_SERVER_URL")
		os.Unsetenv("CHEF_DEFAULT_ORG")
		os.Unsetenv("CHEF_ORG_ALIASES")
	}()

	cfg := LoadFromEnv()
	if cfg.ChefServerURL != "https://chef.example.com" {
		t.Errorf("expected base URL without org, got %s", cfg.ChefServerURL)
	}
	if cfg.DefaultOrg != "acme" {
		t.Errorf("expected extracted org=acme, got %s", cfg.DefaultOrg)
	}
}

func TestLoadFromEnv_ExplicitOrgNotOverridden(t *testing.T) {
	os.Setenv("CHEF_USER", "testuser")
	os.Setenv("CHEF_KEY_PATH", "/tmp/test.pem")
	os.Setenv("CHEF_SERVER_URL", "https://chef.example.com/organizations/url-org")
	os.Setenv("CHEF_DEFAULT_ORG", "explicit-org")
	os.Setenv("CHEF_ORG_ALIASES", "")
	defer func() {
		os.Unsetenv("CHEF_USER")
		os.Unsetenv("CHEF_KEY_PATH")
		os.Unsetenv("CHEF_SERVER_URL")
		os.Unsetenv("CHEF_DEFAULT_ORG")
		os.Unsetenv("CHEF_ORG_ALIASES")
	}()

	cfg := LoadFromEnv()
	if cfg.DefaultOrg != "explicit-org" {
		t.Errorf("explicit CHEF_DEFAULT_ORG should not be overridden, got %s", cfg.DefaultOrg)
	}
}

func TestLoadFromEnv_JSONAliases(t *testing.T) {
	os.Setenv("CHEF_USER", "testuser")
	os.Setenv("CHEF_KEY_PATH", "/tmp/test.pem")
	os.Setenv("CHEF_SERVER_URL", "https://chef.example.com")
	os.Setenv("CHEF_DEFAULT_ORG", "")
	os.Setenv("CHEF_ORG_ALIASES", `{"qa":"qa1","prod":"production"}`)
	defer func() {
		os.Unsetenv("CHEF_USER")
		os.Unsetenv("CHEF_KEY_PATH")
		os.Unsetenv("CHEF_SERVER_URL")
		os.Unsetenv("CHEF_DEFAULT_ORG")
		os.Unsetenv("CHEF_ORG_ALIASES")
	}()

	cfg := LoadFromEnv()
	if cfg.OrgAliases["qa"] != "qa1" {
		t.Errorf("expected alias qa=qa1, got %s", cfg.OrgAliases["qa"])
	}
	if cfg.OrgAliases["prod"] != "production" {
		t.Errorf("expected alias prod=production, got %s", cfg.OrgAliases["prod"])
	}
}

func TestLoadFromEnv_SimpleAliases(t *testing.T) {
	os.Setenv("CHEF_USER", "testuser")
	os.Setenv("CHEF_KEY_PATH", "/tmp/test.pem")
	os.Setenv("CHEF_SERVER_URL", "https://chef.example.com")
	os.Setenv("CHEF_DEFAULT_ORG", "")
	os.Setenv("CHEF_ORG_ALIASES", "qa=qa1,prod=production")
	defer func() {
		os.Unsetenv("CHEF_USER")
		os.Unsetenv("CHEF_KEY_PATH")
		os.Unsetenv("CHEF_SERVER_URL")
		os.Unsetenv("CHEF_DEFAULT_ORG")
		os.Unsetenv("CHEF_ORG_ALIASES")
	}()

	cfg := LoadFromEnv()
	if cfg.OrgAliases["qa"] != "qa1" {
		t.Errorf("expected alias qa=qa1, got %s", cfg.OrgAliases["qa"])
	}
	if cfg.OrgAliases["prod"] != "production" {
		t.Errorf("expected alias prod=production, got %s", cfg.OrgAliases["prod"])
	}
}

func TestResolveOrganization_Default(t *testing.T) {
	cfg := &Config{DefaultOrg: "default-org", OrgAliases: map[string]string{}}
	if got := cfg.ResolveOrganization(""); got != "default-org" {
		t.Errorf("expected default-org, got %s", got)
	}
}

func TestResolveOrganization_Explicit(t *testing.T) {
	cfg := &Config{DefaultOrg: "default-org", OrgAliases: map[string]string{}}
	if got := cfg.ResolveOrganization("explicit"); got != "explicit" {
		t.Errorf("expected explicit, got %s", got)
	}
}

func TestResolveOrganization_Alias(t *testing.T) {
	cfg := &Config{
		DefaultOrg: "default-org",
		OrgAliases: map[string]string{"qa": "qa-actual"},
	}
	if got := cfg.ResolveOrganization("qa"); got != "qa-actual" {
		t.Errorf("expected qa-actual, got %s", got)
	}
}

func TestResolveOrganization_NonAlias(t *testing.T) {
	cfg := &Config{
		DefaultOrg: "default-org",
		OrgAliases: map[string]string{"qa": "qa-actual"},
	}
	if got := cfg.ResolveOrganization("unknown"); got != "unknown" {
		t.Errorf("expected unknown passed through, got %s", got)
	}
}
