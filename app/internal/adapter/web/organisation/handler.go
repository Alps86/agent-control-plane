package organisation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	"agentcontrolplane/ui/bridge"
)

// NewHandler erstellt die öffentlichen JSON- und HTML-Routen.
func NewHandler(service *apporganisation.Service, ui *bridge.Bridge) *Handler {
	h := &Handler{service: service, bridge: ui, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen", h.apiList)
	h.mux.HandleFunc("POST /api/organisationen", h.apiCreate)
	h.mux.HandleFunc("GET /api/organisationen/{id}", h.apiGet)
	h.mux.HandleFunc("GET /organisationen", h.pageList)
	h.mux.HandleFunc("POST /organisationen", h.pageCreate)
	h.mux.HandleFunc("GET /organisationen/neu", h.pageForm)
	h.mux.HandleFunc("GET /organisationen/{id}", h.pageGet)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	organizations, err := h.service.List(r.Context())
	if err != nil {
		h.apiError(w, err)
		return
	}

	if organizations == nil {
		organizations = []domainorganisation.Organization{}
	}

	h.json(w, http.StatusOK, listResponse{Organizations: organizations})
}

func (h *Handler) apiCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	request, err := h.decode(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	organization, err := h.service.Create(r.Context(), request.Name, request.Description)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.apiCreated(w, organization)
}

func (h *Handler) apiCreated(w http.ResponseWriter, organization domainorganisation.Organization) {
	w.Header().Set("Location", "/api/organisationen/"+organization.ID)
	h.json(w, http.StatusCreated, organization)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	organization, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, organization)
}

func (h *Handler) decode(r *http.Request) (createRequest, error) {
	var request createRequest
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
	if errors.Is(err, apporganisation.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, apporganisation.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	h.apiFieldError(w, err)
}

func (h *Handler) apiFieldError(w http.ResponseWriter, err error) {
	if errors.Is(err, apporganisation.ErrNameRequired) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"name": "Bitte geben Sie einen Namen ein."}})
		return
	}

	if errors.Is(err, apporganisation.ErrNameConflict) {
		h.json(w, http.StatusConflict, errorResponse{FieldErrors: map[string]string{"name": "Dieser Name ist bereits vergeben."}})
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
