package aufgabe

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	appagent "agentcontrolplane/app/internal/app/agent"
	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	"agentcontrolplane/ui/bridge"
)

// NewHandler erstellt die Projekt-gebundenen Aufgabenrouten.
func NewHandler(tasks *appaufgabe.Service, projects *appprojekt.Service, archive *appprojektarchiv.Service, agents *appagent.Service, organizations *apporganisation.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{tasks: tasks, projects: projects, archive: archive, agents: agents, organizations: organizations, bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte/{projektID}/aufgaben", h.apiList)
	h.mux.HandleFunc("POST /api/organisationen/{id}/projekte/{projektID}/aufgaben", h.apiCreate)
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}", h.apiGet)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/{projektID}/aufgaben", h.pageList)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/{projektID}/aufgaben/neu", h.pageForm)
	h.mux.HandleFunc("POST /organisationen/{id}/projekte/{projektID}/aufgaben/neu", h.pageCreate)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}", h.pageGet)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.tasks.List(r.Context(), r.PathValue("id"), r.PathValue("projektID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	if tasks == nil {
		tasks = []domainaufgabe.Task{}
	}

	h.json(w, http.StatusOK, listResponse{Tasks: tasks})
}

func (h *Handler) apiCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	request, err := h.decode(r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json", FieldErrors: map[string]string{"assignee_id": "Bitte wählen Sie genau einen Agenten aus."}})
		return
	}

	task, err := h.tasks.Create(r.Context(), r.PathValue("id"), h.input(r, request))
	if err != nil {
		h.apiError(w, err)
		return
	}

	w.Header().Set("Location", h.apiURL(task))
	h.json(w, http.StatusCreated, task)
}

func (h *Handler) apiGet(w http.ResponseWriter, r *http.Request) {
	task, err := h.find(r)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, task)
}

func (h *Handler) decode(r *http.Request) (createRequest, error) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20+1))
	if err != nil || len(data) > 1<<20 {
		return createRequest{}, errors.New("JSON body too large or unreadable")
	}

	return h.decodeObject(json.NewDecoder(bytes.NewReader(data)))
}

func (h *Handler) decodeObject(decoder *json.Decoder) (createRequest, error) {
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return createRequest{}, errors.New("JSON object required")
	}

	var request createRequest
	fields := map[string]*string{"title": &request.Title, "description": &request.Description,
		"priority": &request.Priority, "assignee_id": &request.AssigneeID}
	seen := map[string]bool{}
	for decoder.More() {
		if err := h.decodePair(decoder, fields, seen); err != nil {
			return createRequest{}, err
		}
	}

	return request, h.decodeEnd(decoder)
}

func (h *Handler) decodePair(decoder *json.Decoder, fields map[string]*string, seen map[string]bool) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	key := token.(string)
	field := fields[key]
	if field == nil || seen[key] {
		return errors.New("unknown or duplicate JSON field")
	}

	seen[key] = true
	return h.decodeString(decoder, field)
}

func (h *Handler) decodeString(decoder *json.Decoder, field *string) error {
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return err
	}

	if len(raw) == 0 || raw[0] != '"' {
		return errors.New("JSON field must be a string")
	}

	return json.Unmarshal(raw, field)
}

func (h *Handler) decodeEnd(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil || token != json.Delim('}') {
		return errors.New("JSON object not closed")
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple JSON values")
	}

	return nil
}

func (h *Handler) input(r *http.Request, request createRequest) appaufgabe.CreateInput {
	source := domainaktivitaet.SourceBrowser
	if strings.HasPrefix(r.URL.Path, "/api/organisationen/") {
		source = domainaktivitaet.SourceAPI
	}

	return appaufgabe.CreateInput{ProjectID: r.PathValue("projektID"), Title: request.Title,
		Description: request.Description, Priority: request.Priority, AssigneeID: request.AssigneeID, Source: source}
}

func (h *Handler) apiURL(task domainaufgabe.Task) string {
	return "/api/organisationen/" + task.OrganizationID + "/projekte/" + task.ProjectID + "/aufgaben/" + task.ID
}

func (h *Handler) find(r *http.Request) (domainaufgabe.Task, error) {
	task, err := h.tasks.Find(r.Context(), r.PathValue("id"), r.PathValue("aufgabeID"))
	if err != nil {
		return task, err
	}

	if task.ProjectID != r.PathValue("projektID") {
		return domainaufgabe.Task{}, appaufgabe.ErrNotFound
	}

	return task, nil
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	if errors.Is(err, appaufgabe.ErrProjectArchived) {
		h.json(w, http.StatusConflict, errorResponse{Error: "project_archived"})
		return
	}

	if errors.Is(err, appaufgabe.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, appaufgabe.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	h.apiFieldError(w, err)
}

func (h *Handler) apiFieldError(w http.ResponseWriter, err error) {
	fields := h.fieldErrors(err)
	if fields == nil {
		h.json(w, http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}

	h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: fields})
}

func (h *Handler) fieldErrors(err error) map[string]string {
	if errors.Is(err, appaufgabe.ErrTitleRequired) {
		return map[string]string{"title": "Bitte geben Sie einen Titel ein."}
	}

	if errors.Is(err, appaufgabe.ErrAssigneeRequired) || errors.Is(err, appaufgabe.ErrInvalidAssignee) || errors.Is(err, appaufgabe.ErrAssigneePaused) {
		return map[string]string{"assignee_id": "Bitte wählen Sie einen aktiven Agenten dieser Organisation aus."}
	}

	if errors.Is(err, appaufgabe.ErrInvalidPriority) {
		return map[string]string{"priority": "Bitte wählen Sie eine gültige Priorität aus."}
	}

	if errors.Is(err, appaufgabe.ErrInvalidProject) {
		return map[string]string{"project_id": "Bitte wählen Sie ein gültiges Projekt aus."}
	}

	return nil
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
