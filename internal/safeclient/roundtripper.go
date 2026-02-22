// Package safeclient provides an HTTP transport layer that enforces read-only
// access to Chef Server by blocking mutating HTTP methods (POST, PUT, DELETE, PATCH).
//
// This is the strongest runtime guard in the defense-in-depth strategy:
// even if a developer accidentally adds a write method to the ChefAPI wrapper,
// the underlying HTTP request will be rejected before reaching Chef Server.
package safeclient

import (
	"fmt"
	"net/http"
)

// allowedMethods contains HTTP methods considered safe (non-mutating).
var allowedMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

// ReadOnlyTransport wraps an http.RoundTripper and rejects any HTTP request
// that uses a mutating method (POST, PUT, DELETE, PATCH, etc.).
type ReadOnlyTransport struct {
	// Base is the underlying transport to use for allowed requests.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper
}

// RoundTrip implements http.RoundTripper. It returns an error for any
// non-GET/HEAD/OPTIONS request, preventing mutations to Chef Server.
func (t *ReadOnlyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !allowedMethods[req.Method] {
		return nil, fmt.Errorf("read-only mode: blocked %s request to %s", req.Method, req.URL.Path)
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// NewHTTPClient returns an *http.Client that blocks mutating HTTP methods.
// Pass nil for base to use http.DefaultTransport.
func NewHTTPClient(base http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: &ReadOnlyTransport{Base: base},
	}
}
