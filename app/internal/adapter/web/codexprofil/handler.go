package codexprofil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	runtimeprofil "agentcontrolplane/app/internal/adapter/runtime/codexprofil"
	appagent "agentcontrolplane/app/internal/app/agent"
	appcodexprofil "agentcontrolplane/app/internal/app/codexprofil"
	domainprofil "agentcontrolplane/app/internal/domain/codexprofil"
	"agentcontrolplane/ui/bridge"
)

// NewHandler erzeugt die öffentlichen JSON- und HTML-Routen.
func NewHandler(profiles *appcodexprofil.Service, agents *appagent.Service, ui *bridge.Bridge, bindAddress string, runner Runner) *Handler {
	h := &Handler{profiles: profiles, agents: agents, bridge: ui, bindAddress: bindAddress, runner: runner, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/agenten/{agentID}/codex-profil", h.apiGet)
	h.mux.HandleFunc("PUT /api/organisationen/{id}/agenten/{agentID}/codex-profil", h.apiSave)
	h.mux.HandleFunc("POST /api/organisationen/{id}/agenten/{agentID}/codex-profil/pruefen", h.apiCheck)
	h.mux.HandleFunc("GET /organisationen/{id}/agenten/{agentID}/codex-profil", h.pageGet)
	h.mux.HandleFunc("POST /organisationen/{id}/agenten/{agentID}/codex-profil", h.pageSave)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	profile, err := h.profiles.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, profile)
}

func (h *Handler) apiSave(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	input, err := h.decode(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	profile, err := h.profiles.Save(r.Context(), r.PathValue("id"), r.PathValue("agentID"), input)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, profile)
}

func (h *Handler) decode(r *http.Request) (domainprofil.Input, error) {
	body, err := h.readBody(r)
	if err != nil {
		return domainprofil.Input{}, err
	}

	return h.parseSave(body)
}

func (h *Handler) parseSave(body []byte) (domainprofil.Input, error) {
	var request saveRequest

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return domainprofil.Input{}, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return domainprofil.Input{}, errors.New("multiple JSON values")
	}

	if request.WorkspaceEnabled == nil || request.WriteEnabled == nil {
		return domainprofil.Input{}, errors.New("both booleans required")
	}

	return domainprofil.Input{WorkspaceEnabled: *request.WorkspaceEnabled, WriteEnabled: *request.WriteEnabled}, nil
}

func (h *Handler) readBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 {
		return nil, errors.New("JSON body too large")
	}

	return body, nil
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	status, response := h.publicError(err)
	h.json(w, status, response)
}

func (h *Handler) publicError(err error) (int, errorResponse) {
	if errors.Is(err, appcodexprofil.ErrAccessDenied) {
		return http.StatusForbidden, errorResponse{Error: "access_denied"}
	}

	if errors.Is(err, appcodexprofil.ErrNotFound) {
		return http.StatusNotFound, errorResponse{Error: "not_found"}
	}

	if errors.Is(err, appcodexprofil.ErrWriteWithoutWorkspace) {
		return http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"write_enabled": "Schreiben benötigt einen aktivierten Arbeitsbereich."}}
	}

	if status, response, matched := h.actionError(err); matched {
		return status, response
	}

	if errors.Is(err, appcodexprofil.ErrExecutionKind) || errors.Is(err, runtimeprofil.ErrIntegrity) {
		return http.StatusConflict, errorResponse{Error: "profile_conflict"}
	}

	return http.StatusInternalServerError, errorResponse{Error: "internal_error"}
}

func (h *Handler) actionError(err error) (int, errorResponse, bool) {
	if errors.Is(err, appcodexprofil.ErrWorkspaceDisabled) {
		return http.StatusConflict, errorResponse{Error: "workspace_disabled"}, true
	}

	if errors.Is(err, appcodexprofil.ErrWriteDisabled) {
		return http.StatusForbidden, errorResponse{Error: "write_denied"}, true
	}

	if errors.Is(err, appcodexprofil.ErrActionDenied) {
		return http.StatusForbidden, errorResponse{Error: "action_denied"}, true
	}

	return 0, errorResponse{}, false
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
