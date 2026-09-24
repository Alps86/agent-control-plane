package projekt

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appziel "agentcontrolplane/app/internal/app/ziel"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	"agentcontrolplane/ui/bridge"
)

// NewHandler erstellt die Projekt-Routen für API und Browser.
func NewHandler(projects *appprojekt.Service, organizations *apporganisation.Service, goals *appziel.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{projects: projects, organizations: organizations, goals: goals, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte", h.apiList)
	h.mux.HandleFunc("POST /api/organisationen/{id}/projekte", h.apiCreate)
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte/{projektID}", h.apiGet)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte", h.pageList)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/neu", h.pageForm)
	h.mux.HandleFunc("POST /organisationen/{id}/projekte", h.pageCreate)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/{projektID}", h.pageGet)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.List(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	if projects == nil {
		projects = []domainprojekt.Project{}
	}

	h.json(w, http.StatusOK, listResponse{Projects: projects})
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

	project, err := h.projects.Create(r.Context(), r.PathValue("id"), request.Name, request.Description, request.GoalID)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.apiCreated(w, project)
}

func (h *Handler) apiCreated(w http.ResponseWriter, project domainprojekt.Project) {
	w.Header().Set("Location", "/api/organisationen/"+project.OrganizationID+"/projekte/"+project.ID)
	h.json(w, http.StatusCreated, project)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	project, err := h.projects.Find(r.Context(), r.PathValue("id"), r.PathValue("projektID"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	h.json(w, http.StatusOK, detailResponse{Project: project, Tasks: []any{}})
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
	if errors.Is(err, appprojekt.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, appprojekt.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	h.apiFieldError(w, err)
}

func (h *Handler) apiFieldError(w http.ResponseWriter, err error) {
	if errors.Is(err, appprojekt.ErrNameRequired) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"name": "Bitte geben Sie einen Namen ein."}})
		return
	}

	if errors.Is(err, appprojekt.ErrGoalRequired) || errors.Is(err, appprojekt.ErrInvalidGoal) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"goal_id": "Bitte wählen Sie ein Ziel dieser Organisation aus."}})
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
