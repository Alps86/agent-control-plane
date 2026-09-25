package modellfreigabe

import (
	"errors"
	"net/http"
	"strings"

	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
)

func NewHandler(service *appfreigabe.Service, renderer Renderer, bindAddress string) *Handler {
	h := &Handler{service: service, renderer: renderer, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /organisationen/{id}/modellfreigabe/openrouter", h.pageShow)
	h.mux.HandleFunc("POST /organisationen/{id}/modellfreigabe/openrouter/organisation", h.pageGrantOrganization)
	h.mux.HandleFunc("POST /organisationen/{id}/modellfreigabe/openrouter/organisation/entziehen", h.pageRevokeOrganization)
	h.mux.HandleFunc("POST /organisationen/{id}/modellfreigabe/openrouter/agenten/{agentID}", h.pageGrantAgent)
	h.mux.HandleFunc("POST /organisationen/{id}/modellfreigabe/openrouter/agenten/{agentID}/entziehen", h.pageRevokeAgent)
	h.mux.HandleFunc("GET /api/organisationen/{id}/modellfreigabe/openrouter", h.apiShow)
	h.mux.HandleFunc("POST /api/organisationen/{id}/modellfreigabe/openrouter/organisation", h.apiGrantOrganization)
	h.mux.HandleFunc("DELETE /api/organisationen/{id}/modellfreigabe/openrouter/organisation", h.apiRevokeOrganization)
	h.mux.HandleFunc("GET /api/organisationen/{id}/modellfreigabe/openrouter/agenten/{agentID}", h.apiAgent)
	h.mux.HandleFunc("POST /api/organisationen/{id}/modellfreigabe/openrouter/agenten/{agentID}", h.apiGrantAgent)
	h.mux.HandleFunc("DELETE /api/organisationen/{id}/modellfreigabe/openrouter/agenten/{agentID}", h.apiRevokeAgent)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Referrer-Policy", "same-origin")
	if !h.localRequest(r) {
		h.deny(w, r)
		return
	}

	if r.Method != http.MethodGet && !h.sameOrigin(r) {
		h.deny(w, r)
		return
	}

	h.mux.ServeHTTP(w, r)
}

func (h *Handler) deny(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
}

func (h *Handler) failure(err error) (int, string, string) {
	if errors.Is(err, appfreigabe.ErrAccessDenied) {
		return http.StatusForbidden, "access_denied", "Zugriff verweigert"
	}

	if errors.Is(err, appfreigabe.ErrNotFound) {
		return http.StatusNotFound, "not_found", "Organisation oder Eino-Agent nicht gefunden"
	}

	if errors.Is(err, appfreigabe.ErrConnection) {
		return http.StatusConflict, "connection_unavailable", "OpenRouter-Verbindung nicht einsatzbereit"
	}

	if errors.Is(err, appfreigabe.ErrOrganizationGrantMissing) {
		return http.StatusConflict, "organization_grant_missing", "Organisationsfreigabe fehlt"
	}

	return http.StatusServiceUnavailable, "unavailable", "Freigaben derzeit nicht verfügbar"
}
