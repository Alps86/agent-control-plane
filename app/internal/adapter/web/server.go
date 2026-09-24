package web

import (
	"net/http"

	"agentcontrolplane/app/internal/port/system"
)

// NewServer verdrahtet die HTTP-Grenze mit der Status-Probe.
func NewServer(probe system.Probe) *Server {
	return &Server{probe: probe}
}

// Handler ist der erweiterbare Einstieg für öffentliche Routen.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if !s.probe.Status().Ready {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
