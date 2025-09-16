package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aknarts/chef-server-mcp/internal/config"
	"github.com/aknarts/chef-server-mcp/internal/version"
)

type mockNodeProvider struct {
	list []string
	err  error
}

func (m *mockNodeProvider) ListNodes() ([]string, error) { return m.list, m.err }

func newTestServer(np NodeProvider, fallback bool) *Server {
	cfg := &config.Config{Port: "0", KnifeFallback: fallback}
	return New(cfg, np)
}

func TestHealthz(t *testing.T) {
	s := newTestServer(nil, true)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "OK" {
		t.Fatalf("expected 200 OK body 'OK', got %d %q", w.Code, w.Body.String())
	}
}

func TestVersion(t *testing.T) {
	version.Version = "test-version"
	s := newTestServer(nil, true)
	r := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 || !contains(w.Body.String(), "test-version") {
		t.Fatalf("expected version in body, got %d %q", w.Code, w.Body.String())
	}
}

func TestNodes_APISuccess(t *testing.T) {
	s := newTestServer(&mockNodeProvider{list: []string{"node1", "node2"}}, true)
	r := httptest.NewRequest(http.MethodGet, "/nodes", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 || !contains(w.Body.String(), "node1") || !contains(w.Body.String(), "node2") {
		t.Fatalf("expected nodes in response: %d %q", w.Code, w.Body.String())
	}
}

func TestNodes_APIFail_KnifeSuccess(t *testing.T) {
	orig := runKnife
	runKnife = func(args ...string) (string, error) { return "k1\nk2\n", nil }
	defer func() { runKnife = orig }()
	s := newTestServer(&mockNodeProvider{err: errors.New("boom")}, true)
	r := httptest.NewRequest(http.MethodGet, "/nodes", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 || !contains(w.Body.String(), "k1") || !contains(w.Body.String(), "k2") {
		t.Fatalf("expected knife fallback nodes, got %d %q", w.Code, w.Body.String())
	}
}

func TestNodes_BothFail(t *testing.T) {
	orig := runKnife
	runKnife = func(args ...string) (string, error) { return "", errors.New("knife fail") }
	defer func() { runKnife = orig }()
	s := newTestServer(&mockNodeProvider{err: errors.New("api fail")}, true)
	r := httptest.NewRequest(http.MethodGet, "/nodes", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 500 || !contains(w.Body.String(), "knife fail") {
		t.Fatalf("expected 500 with knife fail message, got %d %q", w.Code, w.Body.String())
	}
}

func TestNodes_NoFallback_NoAPI(t *testing.T) {
	s := newTestServer(nil, false)
	r := httptest.NewRequest(http.MethodGet, "/nodes", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	// No provider and fallback off -> returns empty list 200
	if w.Code != 200 || w.Body.String() != "[]\n" {
		t.Fatalf("expected empty list, got %d %q", w.Code, w.Body.String())
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
