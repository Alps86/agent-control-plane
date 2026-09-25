package modellwahl

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	app "agentcontrolplane/app/internal/app/modellwahl"
	domain "agentcontrolplane/app/internal/domain/modellwahl"
	port "agentcontrolplane/app/internal/port/modellwahl"
	"agentcontrolplane/ui/bridge"
)

// NewHandler registers public API and HTML routes on one local handler.
func NewHandler(service *app.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{service: service, ui: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten/{agentID}/modellwahl", h.apiGet)
	h.mux.HandleFunc("PUT /api/organisationen/{id}/agenten/{agentID}/modellwahl", h.apiPut)
	h.mux.HandleFunc("GET /organisationen/{id}/agenten/{agentID}/modellwahl", h.pageGet)
	h.mux.HandleFunc("POST /organisationen/{id}/agenten/{agentID}/modellwahl", h.pagePost)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, view)
}

func (h *Handler) apiPut(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	selection, err := h.readSelection(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	view, err := h.service.Put(r.Context(), r.PathValue("id"), r.PathValue("agentID"), selection)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, view)
}

func (h *Handler) readSelection(r *http.Request) (domain.Selection, error) {
	var selection domain.Selection
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&selection); err != nil {
		return domain.Selection{}, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return domain.Selection{}, errors.New("multiple JSON values")
	}

	return selection, nil
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	if errors.Is(err, app.ErrNotFound) || errors.Is(err, port.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	if errors.Is(err, app.ErrCatalog) || errors.Is(err, app.ErrExecutionKind) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{Error: "invalid_model_route", Reason: err.Error()})
		return
	}

	if h.grantError(err) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "model_connection_denied", Reason: err.Error()})
		return
	}

	h.json(w, http.StatusInternalServerError, errorResponse{Error: "internal_error"})
}

func (h *Handler) grantError(err error) bool {
	return errors.Is(err, app.ErrGrant) || errors.Is(err, app.ErrAccessDenied) ||
		errors.Is(err, port.ErrOrganizationGrant) || errors.Is(err, port.ErrAgentGrant) ||
		errors.Is(err, port.ErrConnectionUnavailable)
}

func (h *Handler) json(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
