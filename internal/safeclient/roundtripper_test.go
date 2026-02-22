package safeclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadOnlyTransport_AllowsGET(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	client := NewHTTPClient(http.DefaultTransport)
	resp, err := client.Get(backend.URL + "/nodes")
	if err != nil {
		t.Fatalf("GET should be allowed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestReadOnlyTransport_AllowsHEAD(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	client := NewHTTPClient(http.DefaultTransport)
	resp, err := client.Head(backend.URL + "/nodes")
	if err != nil {
		t.Fatalf("HEAD should be allowed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestReadOnlyTransport_BlocksPOST(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("POST request should not reach backend")
	}))
	defer backend.Close()

	client := NewHTTPClient(http.DefaultTransport)
	_, err := client.Post(backend.URL+"/nodes", "application/json", nil)
	if err == nil {
		t.Fatal("POST should be blocked")
	}
	if got := err.Error(); !contains(got, "read-only mode") {
		t.Fatalf("error should mention read-only mode, got: %s", got)
	}
}

func TestReadOnlyTransport_BlocksPUT(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("PUT request should not reach backend")
	}))
	defer backend.Close()

	client := NewHTTPClient(http.DefaultTransport)
	req, _ := http.NewRequest(http.MethodPut, backend.URL+"/nodes/web01", nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("PUT should be blocked")
	}
	if got := err.Error(); !contains(got, "read-only mode") {
		t.Fatalf("error should mention read-only mode, got: %s", got)
	}
}

func TestReadOnlyTransport_BlocksDELETE(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("DELETE request should not reach backend")
	}))
	defer backend.Close()

	client := NewHTTPClient(http.DefaultTransport)
	req, _ := http.NewRequest(http.MethodDelete, backend.URL+"/nodes/web01", nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("DELETE should be blocked")
	}
	if got := err.Error(); !contains(got, "read-only mode") {
		t.Fatalf("error should mention read-only mode, got: %s", got)
	}
}

func TestReadOnlyTransport_BlocksPATCH(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("PATCH request should not reach backend")
	}))
	defer backend.Close()

	client := NewHTTPClient(http.DefaultTransport)
	req, _ := http.NewRequest(http.MethodPatch, backend.URL+"/nodes/web01", nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("PATCH should be blocked")
	}
	if got := err.Error(); !contains(got, "read-only mode") {
		t.Fatalf("error should mention read-only mode, got: %s", got)
	}
}

func TestReadOnlyTransport_ErrorIncludesMethod(t *testing.T) {
	client := NewHTTPClient(http.DefaultTransport)
	req, _ := http.NewRequest(http.MethodDelete, "http://localhost/nodes/web01", nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); !contains(got, "DELETE") {
		t.Fatalf("error should include method name, got: %s", got)
	}
}

func TestReadOnlyTransport_NilBase(t *testing.T) {
	// Verify that nil base defaults to http.DefaultTransport (GET should work)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	client := &http.Client{Transport: &ReadOnlyTransport{Base: nil}}
	resp, err := client.Get(backend.URL + "/test")
	if err != nil {
		t.Fatalf("GET with nil base should work: %v", err)
	}
	resp.Body.Close()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
