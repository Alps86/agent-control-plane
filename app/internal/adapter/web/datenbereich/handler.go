package datenbereich

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	appdatenbereich "agentcontrolplane/app/internal/app/datenbereich"
	domaindatenbereich "agentcontrolplane/app/internal/domain/datenbereich"
	"agentcontrolplane/ui/bridge"
)

// NewHandler creates the operator's JSON and HTML data scope routes.
func NewHandler(service *appdatenbereich.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{service: service, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten/{agentID}/datenbereich", h.apiGet)
	h.mux.HandleFunc("PUT /api/organisationen/{id}/agenten/{agentID}/datenbereich", h.apiReplace)
	h.mux.HandleFunc("GET /organisationen/{id}/agenten/{agentID}/datenbereich", h.pageGet)
	h.mux.HandleFunc("POST /organisationen/{id}/agenten/{agentID}/datenbereich", h.pageReplace)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	scopes, err := h.service.List(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.writeScopes(w, scopes)
}

func (h *Handler) apiReplace(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	var request replaceRequest
	if err := h.decodeJSON(w, r, &request); err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	h.replace(w, r, request.Scopes)
}

func (h *Handler) replace(w http.ResponseWriter, r *http.Request, inputs []appdatenbereich.ScopeInput) {
	scopes, err := h.service.Replace(r.Context(), r.PathValue("id"), r.PathValue("agentID"), inputs)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.writeScopes(w, scopes)
}

func (h *Handler) writeScopes(w http.ResponseWriter, scopes []domaindatenbereich.Scope) {
	if scopes == nil {
		scopes = []domaindatenbereich.Scope{}
	}

	h.json(w, http.StatusOK, scopeResponse{Scopes: scopes})
}

func (h *Handler) decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple JSON values")
	}

	return nil
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	if errors.Is(err, appdatenbereich.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, appdatenbereich.ErrInvalidScope) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{Error: "invalid_scope"})
		return
	}

	if errors.Is(err, appdatenbereich.ErrNotFound) {
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
