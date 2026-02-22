package chefapi

import (
	"sync"
	"testing"
)

func TestChefReaderInterface_Compile(t *testing.T) {
	// This test verifies at compile time that ChefAPI implements ChefReader.
	// If someone adds a method to ChefReader without implementing it on ChefAPI,
	// this file won't compile.
	var _ ChefReader = (*ChefAPI)(nil)
}

func TestNewChefAPI_InlineKey(t *testing.T) {
	key := "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----"
	api, err := NewChefAPI("testuser", key, "https://chef.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if api.keyMaterial != key {
		t.Error("expected inline key to be used directly")
	}
}

func TestNewChefAPI_FileKey(t *testing.T) {
	_, err := NewChefAPI("testuser", "/nonexistent/path.pem", "https://chef.example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent key file")
	}
}

func TestNewChefAPI_URLTrailingSlash(t *testing.T) {
	key := "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----"
	api, err := NewChefAPI("testuser", key, "https://chef.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if api.baseURL != "https://chef.example.com/" {
		t.Errorf("expected trailing slash, got %s", api.baseURL)
	}
}

func TestNewChefAPI_URLAlreadyHasTrailingSlash(t *testing.T) {
	key := "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----"
	api, err := NewChefAPI("testuser", key, "https://chef.example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if api.baseURL != "https://chef.example.com/" {
		t.Errorf("expected single trailing slash, got %s", api.baseURL)
	}
}

func TestGetClientForOrg_EmptyOrg(t *testing.T) {
	key := "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----"
	api, _ := NewChefAPI("testuser", key, "https://chef.example.com")
	_, err := api.getClientForOrg("")
	if err == nil {
		t.Fatal("expected error for empty organization")
	}
}

func TestGetClientForOrg_ConcurrentAccess(t *testing.T) {
	key := "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----"
	api, _ := NewChefAPI("testuser", key, "https://chef.example.com")

	// This test exercises the mutex protection. With -race flag, it will
	// catch data races on the clients map.
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(org string) {
			defer wg.Done()
			// These will fail to actually connect, but the cache logic
			// and mutex should work correctly under concurrent access.
			_, _ = api.getClientForOrg(org)
		}("org" + string(rune('A'+i%5)))
	}
	wg.Wait()
}
