package verbindungsstatus

import (
	"encoding/json"
	"net/http"

	app "agentcontrolplane/app/internal/app/verbindungsstatus"
	"agentcontrolplane/ui/bridge"
)

func NewHandler(service *app.Service, ui *bridge.Bridge) *Handler {
	handler := &Handler{service: service, ui: ui, mux: http.NewServeMux()}
	handler.mux.HandleFunc("GET /api/settings/modellanbieter/status", handler.apiGet)
	handler.mux.HandleFunc("GET /settings/modellanbieter/status", handler.pageGet)
	return handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(h.service.Get(r.Context()))
}
