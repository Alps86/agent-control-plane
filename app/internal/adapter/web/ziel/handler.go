package ziel

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appziel "agentcontrolplane/app/internal/app/ziel"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
	"agentcontrolplane/ui/bridge"
)

// NewHandler erstellt die Zielrouten für API und Browser.
func NewHandler(goals *appziel.Service, organizations *apporganisation.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{goals: goals, organizations: organizations, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/ziele", h.apiList)
	h.mux.HandleFunc("POST /api/organisationen/{id}/ziele", h.apiCreate)
	h.mux.HandleFunc("GET /api/organisationen/{id}/zielbaum", h.apiTree)
	h.mux.HandleFunc("POST /api/organisationen/{id}/ziele/{zielID}/kinder", h.apiCreateChild)
	h.mux.HandleFunc("POST /api/organisationen/{id}/ziele/{zielID}/status", h.apiStatus)
	h.mux.HandleFunc("POST /api/organisationen/{id}/ziele/{zielID}/verschieben", h.apiMove)
	h.mux.HandleFunc("GET /organisationen/{id}/ziele", h.pageList)
	h.mux.HandleFunc("GET /organisationen/{id}/ziele/neu", h.pageForm)
	h.mux.HandleFunc("POST /organisationen/{id}/ziele", h.pageCreate)
	h.mux.HandleFunc("POST /organisationen/{id}/ziele/{zielID}/kinder", h.pageCreateChild)
	h.mux.HandleFunc("POST /organisationen/{id}/ziele/{zielID}/status", h.pageStatus)
	h.mux.HandleFunc("POST /organisationen/{id}/ziele/{zielID}/verschieben", h.pageMove)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	goals, err := h.goals.List(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	if goals == nil {
		goals = []domainziel.Goal{}
	}

	h.json(w, http.StatusOK, listResponse{Goals: goals})
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

	goal, err := h.goals.Create(r.Context(), r.PathValue("id"), request.Name)
	if err != nil {
		h.apiError(w, err)
		return
	}

	w.Header().Set("Location", "/api/organisationen/"+goal.OrganizationID+"/ziele")
	h.json(w, http.StatusCreated, goal)
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
	if errors.Is(err, appziel.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, appziel.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	h.apiFieldError(w, err)
}

func (h *Handler) apiFieldError(w http.ResponseWriter, err error) {
	if errors.Is(err, appziel.ErrNameRequired) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"name": "Bitte geben Sie einen Namen ein."}})
		return
	}

	if errors.Is(err, appziel.ErrInvalidParent) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"parent_goal_id": "Bitte wählen Sie ein Ziel dieser Organisation aus."}})
		return
	}

	if errors.Is(err, appziel.ErrGoalCycle) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"parent_goal_id": "Diese Zielkante würde einen Kreis bilden."}})
		return
	}

	if errors.Is(err, appziel.ErrInvalidStatus) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"status": "Bitte wählen Sie einen gültigen Status aus."}})
		return
	}

	if errors.Is(err, appziel.ErrGoalNotFound) {
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
