package kommentar

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	"agentcontrolplane/ui/bridge"
)

// NewHandler stellt die organisations- und aufgabengebundenen Threadrouten bereit.
func NewHandler(comments *appkommentar.Service, tasks *appaufgabe.Service, projects *appprojekt.Service, organizations *apporganisation.Service, ui *bridge.Bridge, bindAddress string) *Handler {
	h := &Handler{comments: comments, tasks: tasks, projects: projects, organizations: organizations,
		bridge: ui, bindAddress: bindAddress, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare", h.apiList)
	h.mux.HandleFunc("POST /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare", h.apiCreate)
	h.mux.HandleFunc("PUT /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare/{kommentarID}", h.apiUpdate)
	h.mux.HandleFunc("GET /organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare", h.pageGet)
	h.mux.HandleFunc("POST /organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare", h.pageCreate)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

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

func (h *Handler) apiList(w http.ResponseWriter, r *http.Request) {
	if _, err := h.find(r); err != nil {
		h.apiError(w, err)
		return
	}

	comments, err := h.comments.List(r.Context(), r.PathValue("id"), r.PathValue("aufgabeID"))
	if err != nil {
		h.apiError(w, err)
		return
	}

	if comments == nil {
		comments = []domainkommentar.Comment{}
	}

	h.json(w, http.StatusOK, map[string]any{"comments": comments})
}

func (h *Handler) apiCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	input, err := h.decodeCreate(w, r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	h.createComment(w, r, input)
}

func (h *Handler) createComment(w http.ResponseWriter, r *http.Request, input appkommentar.CreateInput) {
	if _, err := h.find(r); err != nil {
		h.apiError(w, err)
		return
	}

	comment, err := h.comments.Create(r.Context(), r.PathValue("id"), r.PathValue("aufgabeID"), input)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusCreated, comment)
}

func (h *Handler) apiUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	content, err := h.decodeContent(w, r)
	if err != nil {
		h.json(w, http.StatusBadRequest, errorResponse{Error: "invalid_json"})
		return
	}

	h.updateComment(w, r, content)
}

func (h *Handler) updateComment(w http.ResponseWriter, r *http.Request, content string) {
	if _, err := h.find(r); err != nil {
		h.apiError(w, err)
		return
	}

	comment, err := h.comments.Update(r.Context(), r.PathValue("id"), r.PathValue("aufgabeID"), r.PathValue("kommentarID"), content)
	if err != nil {
		h.apiError(w, err)
		return
	}

	h.json(w, http.StatusOK, comment)
}

func (h *Handler) decodeCreate(w http.ResponseWriter, r *http.Request) (appkommentar.CreateInput, error) {
	var request createRequest
	if err := h.decodeJSON(w, r, &request); err != nil {
		return appkommentar.CreateInput{}, err
	}

	input := appkommentar.CreateInput{Content: request.Content}
	if request.Reference != nil {
		input.Reference = &domainkommentar.Reference{Type: request.Reference.Type, ID: request.Reference.ID,
			OrganizationID: request.Reference.OrganizationID}
	}

	return input, nil
}

func (h *Handler) decodeContent(w http.ResponseWriter, r *http.Request) (string, error) {
	var input updateRequest
	err := h.decodeJSON(w, r, &input)
	return input.Content, err
}

func (h *Handler) decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple JSON values")
	}

	return nil
}

func (h *Handler) apiError(w http.ResponseWriter, err error) {
	if errors.Is(err, appaufgabe.ErrNotFound) || errors.Is(err, appkommentar.ErrNotFound) {
		h.json(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}

	if errors.Is(err, appkommentar.ErrAccessDenied) || errors.Is(err, appaufgabe.ErrAccessDenied) {
		h.json(w, http.StatusForbidden, errorResponse{Error: "access_denied"})
		return
	}

	if errors.Is(err, appkommentar.ErrContentRequired) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"content": "Bitte geben Sie einen Kommentar ein."}})
		return
	}

	h.apiOtherError(w, err)
}

func (h *Handler) apiOtherError(w http.ResponseWriter, err error) {
	if errors.Is(err, appkommentar.ErrInvalidReference) {
		h.json(w, http.StatusUnprocessableEntity, errorResponse{FieldErrors: map[string]string{"reference": "Die Referenz ist ungültig."}})
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
