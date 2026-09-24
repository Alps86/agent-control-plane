package organisation

import (
	"bytes"
	"errors"
	"net/http"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
)

func (h *Handler) pageList(w http.ResponseWriter, r *http.Request) {
	organizations, err := h.service.List(r.Context())
	if err != nil {
		h.pageError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(organizations))
	for _, organization := range organizations {
		items = append(items, h.project(organization))
	}

	data := h.pageData("Organisationsübersicht", "list")
	data["AllowedActions"] = []string{"organisation.create"}
	data["View"].(map[string]any)["Organizations"] = items
	h.render(w, http.StatusOK, data)
}

func (h *Handler) pageForm(w http.ResponseWriter, r *http.Request) {
	if _, err := h.service.List(r.Context()); err != nil {
		h.pageError(w, err)
		return
	}

	h.render(w, http.StatusOK, h.formData("", "", ""))
}

func (h *Handler) pageCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		h.render(w, http.StatusBadRequest, h.formData("", "", "Formulardaten konnten nicht gelesen werden."))
		return
	}

	organization, err := h.service.Create(r.Context(), r.PostFormValue("name"), r.PostFormValue("description"))
	if err != nil {
		h.formError(w, err, r.PostFormValue("name"), r.PostFormValue("description"))
		return
	}

	http.Redirect(w, r, "/organisationen/"+organization.ID, http.StatusSeeOther)
}

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	organization, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, err)
		return
	}

	data := h.pageData(organization.Name, "detail")
	data["View"].(map[string]any)["Organization"] = h.project(organization)
	h.render(w, http.StatusOK, data)
}

func (h *Handler) project(organization domainorganisation.Organization) map[string]any {
	return map[string]any{
		"ID":                    organization.ID,
		"Name":                  organization.Name,
		"Description":           organization.Description,
		"WorkTransitions":       organization.WorkflowPolicy.WorkTransitions,
		"DelegationTransitions": organization.WorkflowPolicy.DelegationTransitions,
	}
}

func (h *Handler) pageData(title, kind string) map[string]any {
	return map[string]any{
		"PageTitle":      title,
		"Navigation":     []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}},
		"Notice":         "",
		"Errors":         []string{},
		"AllowedActions": []string{},
		"View":           map[string]any{"Kind": kind},
	}
}

func (h *Handler) formData(name, description, fieldError string) map[string]any {
	data := h.pageData("Organisation anlegen", "form")
	data["AllowedActions"] = []string{"organisation.create"}
	view := data["View"].(map[string]any)
	view["Name"] = name
	view["Description"] = description
	view["FieldErrors"] = map[string]string{}
	if fieldError != "" {
		view["FieldErrors"] = map[string]string{"name": fieldError}
	}

	return data
}

func (h *Handler) formError(w http.ResponseWriter, err error, name, description string) {
	if errors.Is(err, apporganisation.ErrNameRequired) {
		h.render(w, http.StatusUnprocessableEntity, h.formData(name, description, "Bitte geben Sie einen Namen ein."))
		return
	}

	if errors.Is(err, apporganisation.ErrNameConflict) {
		h.render(w, http.StatusConflict, h.formData(name, description, "Dieser Name ist bereits vergeben."))
		return
	}

	h.pageError(w, err)
}

func (h *Handler) pageError(w http.ResponseWriter, err error) {
	if errors.Is(err, apporganisation.ErrAccessDenied) {
		h.pageFailure(w, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if errors.Is(err, apporganisation.ErrNotFound) {
		h.pageFailure(w, http.StatusNotFound, "Organisation nicht gefunden")
		return
	}

	h.pageFailure(w, http.StatusInternalServerError, "Die Seite ist derzeit nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, status int, message string) {
	data := h.pageData("Organisationen", "error")
	data["Errors"] = []string{message}
	h.render(w, status, data)
}

func (h *Handler) render(w http.ResponseWriter, status int, data map[string]any) {
	if h.bridge == nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	var output bytes.Buffer
	if err := h.bridge.Render(&output, "organisation/page", data); err != nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}
