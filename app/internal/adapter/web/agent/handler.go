package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	appagent "agentcontrolplane/app/internal/app/agent"
	domainagent "agentcontrolplane/app/internal/domain/agent"
	"agentcontrolplane/ui/bridge"
)

// NewHandler builds the JSON and HTML routes for one agent service.
func NewHandler(service *appagent.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{service: service, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten", h.apiList)
	h.mux.HandleFunc("POST /api/organisationen/{id}/agenten", h.apiCreate)
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten/vorlagen", h.apiTemplates)
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten/teamvorlagen", h.apiTeams)
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten/{agentID}", h.apiGet)
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten/{agentID}/bereitschaft", h.apiReadiness)
	h.mux.HandleFunc("POST /api/organisationen/{id}/agenten/{agentID}/start", h.apiStart)
	h.mux.HandleFunc("GET /organisationen/{id}/agenten", h.pageList)
	h.mux.HandleFunc("POST /organisationen/{id}/agenten", h.pageCreate)
	h.mux.HandleFunc("GET /organisationen/{id}/agenten/neu", h.pageForm)
	h.mux.HandleFunc("GET /organisationen/{id}/agenten/{agentID}", h.pageGet)
	h.mux.HandleFunc("POST /organisationen/{id}/agenten/{agentID}/start", h.pageStart)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.service.List(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	if profiles == nil {
		profiles = []appagent.Profile{}
	}
	h.json(w, http.StatusOK, listResponse{Agents: profiles})
}

func (h *Handler) apiCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}
	var input appagent.CreateInput
	if err := h.decodeJSON(r, &input); err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}
	profile, err := h.service.Create(r.Context(), r.PathValue("id"), input)
	if err != nil {
		h.apiError(w, err)
		return
	}
	w.Header().Set("Location", "/api/organisationen/"+url.PathEscape(r.PathValue("id"))+"/agenten/"+url.PathEscape(profile.ID))
	h.json(w, http.StatusCreated, profile)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	profile, err := h.service.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	h.json(w, http.StatusOK, profile)
}

func (h *Handler) apiTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.service.Templates(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	if templates == nil {
		templates = []domainagent.Template{}
	}
	h.json(w, http.StatusOK, templateResponse{Templates: templates})
}

func (h *Handler) apiTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.service.TeamTemplates(r.Context(), r.PathValue("id"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	if teams == nil {
		teams = []domainagent.TeamTemplate{}
	}
	h.json(w, http.StatusOK, teamResponse{TeamTemplates: teams})
}

func (h *Handler) apiReadiness(w http.ResponseWriter, r *http.Request) {
	readiness, err := h.service.CanStart(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	h.json(w, http.StatusOK, readiness)
}

func (h *Handler) apiStart(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}
	readiness, err := h.service.CanStart(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiError(w, err)
		return
	}
	if !readiness.Ready {
		h.json(w, http.StatusConflict, errorResponse{Error: readiness.Code, Reason: h.startReason(readiness)})
		return
	}
	h.json(w, http.StatusNotImplemented, errorResponse{Error: "run_not_available"})
}

func (h *Handler) startReason(readiness domainagent.Readiness) string {
	if readiness.Code == "model_unconfigured" {
		return "Modellverbindung fehlt"
	}

	return readiness.Reason
}

func (h *Handler) decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
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
	switch {
	case errors.Is(err, appagent.ErrAccessDenied):
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
	case errors.Is(err, appagent.ErrNotFound):
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
	case errors.Is(err, appagent.ErrNameRequired):
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"name": "Bitte geben Sie einen Namen ein."}})
	case errors.Is(err, appagent.ErrNameConflict):
		h.json(w, http.StatusConflict, errorResponse{FieldErrors: map[string]string{"name": "Dieser Name ist bereits vergeben."}})
	case errors.Is(err, appagent.ErrTemplateNotFound):
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"template_id": "Vorlage nicht gefunden."}})
	case errors.Is(err, appagent.ErrExecutionKind):
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"execution_kind": "Ausführungsart passt nicht zur Vorlage."}})
	case errors.Is(err, appagent.ErrCapabilityDenied):
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"additional_capabilities": "Fachfähigkeit ist nicht zulässig."}})
	default:
		h.json(w, http.StatusInternalServerError, errorResponse{Error: "internal_error"})
	}
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
