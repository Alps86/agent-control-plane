package aktivitaet

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	appaktivitaet "agentcontrolplane/app/internal/app/aktivitaet"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
	"agentcontrolplane/ui/bridge"
)

// NewHandler bindet Aktivitäts-API und Browserseite an die Anwendungsgrenze.
func NewHandler(service *appaktivitaet.Service, organizations *apporganisation.Service, ui *bridge.Bridge) *Handler {
	h := &Handler{service: service, organizations: organizations, bridge: ui, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/aktivitaet", h.apiList)
	h.mux.HandleFunc("GET /organisationen/{id}/aktivitaet", h.pageList)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.List(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	if events == nil {
		events = []domainaktivitaet.Event{}
	}

	h.json(w, http.StatusOK, listResponse{Events: events})
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	if errors.Is(err, appaktivitaet.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, appaktivitaet.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	h.json(w, http.StatusInternalServerError, errorResponse{Error: "internal_error"})
}

func (h *Handler) json(w http.ResponseWriter, status int, value any) {
	var output bytes.Buffer
	if err := json.NewEncoder(&output).Encode(value); err != nil {
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}
