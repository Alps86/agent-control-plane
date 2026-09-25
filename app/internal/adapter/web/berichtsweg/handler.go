package berichtsweg

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	appberichtsweg "agentcontrolplane/app/internal/app/berichtsweg"
	"agentcontrolplane/ui/bridge"
)

// NewHandler registers the chart API and HTML routes.
func NewHandler(service *appberichtsweg.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{service: service, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/berichtswege", h.apiGet)
	h.mux.HandleFunc("PUT /api/organisationen/{id}/agenten/{agentID}/berichtsweg", h.apiAssign)
	h.mux.HandleFunc("GET /organisationen/{id}/berichtswege", h.pageGet)
	h.mux.HandleFunc("POST /organisationen/{id}/agenten/{agentID}/berichtsweg", h.pageAssign)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	chart, err := h.service.Chart(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, chart)
}

func (h *Handler) apiAssign(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	var input assignRequest
	if err := h.decodeJSON(w, r, &input); err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	chart, err := h.service.Assign(r.Context(), r.PathValue("id"), r.PathValue("agentID"), input.ParentID)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, chart)
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
	status, reason := h.errorStatus(err)
	h.json(w, status, errorResponse{Error: reason})
}

func (h *Handler) errorStatus(err error) (int, string) {
	if errors.Is(err, appberichtsweg.ErrAccessDenied) || errors.Is(err, appberichtsweg.ErrNotFound) {
		return http.StatusNotFound, "not_found"
	}

	if errors.Is(err, appberichtsweg.ErrAgentNotFound) {
		return http.StatusNotFound, "agent_not_found"
	}

	if errors.Is(err, appberichtsweg.ErrSelfParent) {
		return http.StatusUnprocessableEntity, "self_parent"
	}

	if errors.Is(err, appberichtsweg.ErrCycle) {
		return http.StatusUnprocessableEntity, "cycle"
	}

	return http.StatusInternalServerError, "internal_error"
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
