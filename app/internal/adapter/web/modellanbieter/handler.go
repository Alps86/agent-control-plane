package modellanbieter

import (
	"encoding/json"
	"net/http"

	"agentcontrolplane/app/internal/app/modellverbindung"
)

func New(flow modellverbindung.DeviceFlow) *Handler {
	return &Handler{flow: flow}
}

func (h *Handler) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /settings/modelle/codex/device/start", h.start)
	mux.HandleFunc("GET /settings/modelle/codex/device/status", h.status)
	mux.HandleFunc("POST /settings/modelle/codex/device/cancel", h.cancel)
	return mux
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	status, err := h.flow.Start(r.Context())
	h.respond(w, status, err)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	status, err := h.flow.Status(r.Context())
	h.respond(w, status, err)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	status, err := h.flow.Cancel(r.Context())
	h.respond(w, status, err)
}

func (h *Handler) respond(w http.ResponseWriter, status modellverbindung.Status, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(modellverbindung.Status{State: "unavailable", Reason: "transport_unavailable"})
		return
	}

	_ = json.NewEncoder(w).Encode(status)
}
