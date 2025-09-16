package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/aknarts/chef-server-mcp/internal/config"
	"github.com/aknarts/chef-server-mcp/internal/knife"
	"github.com/aknarts/chef-server-mcp/internal/nodes"
	"github.com/aknarts/chef-server-mcp/internal/version"
)

// NodeProvider abstracts listing nodes (Chef API or mock).
type NodeProvider interface {
	ListNodes() ([]string, error)
}

// runKnife is overridable in tests.
var runKnife = knife.RunKnifeCommand

// Server encapsulates HTTP server state.
type Server struct {
	cfg   *config.Config
	chef  NodeProvider
	mux   *http.ServeMux
	httpS *http.Server
}

// New creates a new Server.
func New(cfg *config.Config, chefClient NodeProvider) *Server {
	s := &Server{cfg: cfg, chef: chefClient, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/version", s.handleVersion)
	s.mux.HandleFunc("/nodes", s.handleNodes)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"version": version.Version})
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodesList, err := nodes.ListNodes(s.chef, s.cfg.KnifeFallback, runKnife)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nodesList); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("encode error: " + err.Error()))
	}
}

func splitLines(s string) []string {
	// moved to nodes package, keep stub if still referenced elsewhere
	return strings.Split(strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n")), "\n")
}

func errString(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}

// Start launches the HTTP server (blocking).
func (s *Server) Start() error {
	s.httpS = &http.Server{Addr: ":" + s.cfg.Port, Handler: s.mux}
	log.Printf("listening on :%s", s.cfg.Port)
	err := s.httpS.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpS == nil {
		return nil
	}
	return s.httpS.Shutdown(ctx)
}

// Handler exposes mux for tests.
func (s *Server) Handler() http.Handler { return s.mux }
