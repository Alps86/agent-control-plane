package projektort

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	apport "agentcontrolplane/app/internal/app/projektort"
	"agentcontrolplane/ui/bridge"
)

// NewHandler bindet die vier öffentlichen Projektort-Routen.
func NewHandler(locations *apport.Service, projects *appprojekt.Service, organizations *apporganisation.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{locations: locations, projects: projects, organizations: organizations, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte/{projektID}/ausfuehrungsort", h.apiGet)
	h.mux.HandleFunc("PUT /api/organisationen/{id}/projekte/{projektID}/ausfuehrungsort", h.apiBind)
	h.mux.HandleFunc("POST /api/organisationen/{id}/projekte/{projektID}/ausfuehrungsort/startpruefung", h.apiPrepare)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/{projektID}/ausfuehrungsort", h.pageGet)
	h.mux.HandleFunc("POST /organisationen/{id}/projekte/{projektID}/ausfuehrungsort", h.pageBind)
	h.mux.HandleFunc("POST /organisationen/{id}/projekte/{projektID}/ausfuehrungsort/startpruefung", h.pagePrepare)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	status, err := h.locations.Get(r.Context(), r.PathValue("id"), r.PathValue("projektID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, status)
}

func (h *Handler) apiBind(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	request, err := h.decode(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	status, err := h.locations.Bind(r.Context(), r.PathValue("id"), r.PathValue("projektID"), request.Kind)
	if err != nil {
		h.apiError(w, err)
		return
	}

	code := http.StatusOK
	if !status.StartReady {
		code = http.StatusConflict
	}

	h.json(w, code, status)
}

func (h *Handler) apiPrepare(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	result, err := h.locations.Prepare(r.Context(), r.PathValue("id"), r.PathValue("projektID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	code := http.StatusOK
	if !result.Ready {
		code = http.StatusConflict
	}
	if result.Code == apport.CodeAdapterUnavailable {
		code = http.StatusServiceUnavailable
	}

	h.json(w, code, result)
}

func (h *Handler) decode(r *http.Request) (bindRequest, error) {
	var request bindRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return request, errors.New("multiple JSON values")
	}

	return request, nil
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	status, response := h.publicError(err)
	h.json(w, status, response)
}

func (h *Handler) publicError(err error) (int, errorResponse) {
	if errors.Is(err, apport.ErrKindInvalid) {
		return http.StatusUnprocessableEntity, errorResponse{Error: "invalid_kind", Reason: err.Error()}
	}

	if errors.Is(err, apport.ErrNotFound) || errors.Is(err, appprojekt.ErrNotFound) || errors.Is(err, apporganisation.ErrNotFound) {
		return http.StatusNotFound, errorResponse{Error: "not_found"}
	}

	if errors.Is(err, apport.ErrAccessDenied) || errors.Is(err, appprojekt.ErrAccessDenied) || errors.Is(err, apporganisation.ErrAccessDenied) {
		return http.StatusForbidden, errorResponse{Error: "access_denied"}
	}

	return http.StatusInternalServerError, errorResponse{Error: "internal_error"}
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
