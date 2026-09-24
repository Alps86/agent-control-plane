package web

import (
	"fmt"
	"net/http"

	"agentcontrolplane/app/internal/port/system"
)

// NewServer verdrahtet die HTTP-Grenze mit der Status-Probe.
func NewServer(probe system.Probe, store SchemaVersioner) *Server {
	server := &Server{probe: probe, mux: http.NewServeMux(), store: store}
	server.mux.HandleFunc("GET /health", server.health)
	return server
}

// Handler ist der erweiterbare Einstieg für öffentliche Routen.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// Handle ergänzt eine öffentliche Route vor dem Serverstart.
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if !s.probe.Status().Ready {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}

	version, err := s.store.SchemaVersion(r.Context())
	if err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"schema_version":%d}`, version)
}
