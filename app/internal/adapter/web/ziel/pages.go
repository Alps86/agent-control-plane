package ziel

import (
	"bytes"
	"errors"
	"net/http"

	appziel "agentcontrolplane/app/internal/app/ziel"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
)

func (h *Handler) pageList(w http.ResponseWriter, r *http.Request) {
	data, err := h.treePageData(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) pageForm(w http.ResponseWriter, r *http.Request) {
	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, h.formData(organization, "", ""))
}

func (h *Handler) pageCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		h.pageFailure(w, r, http.StatusBadRequest, "Formulardaten konnten nicht gelesen werden.")
		return
	}

	h.createForm(w, r, r.PostFormValue("name"))
}

func (h *Handler) createForm(w http.ResponseWriter, r *http.Request, name string) {
	goal, err := h.goals.Create(r.Context(), r.PathValue("id"), name)
	if err != nil {
		h.formError(w, r, err, name)
		return
	}

	http.Redirect(w, r, "/organisationen/"+goal.OrganizationID+"/ziele", http.StatusSeeOther)
}

func (h *Handler) formError(w http.ResponseWriter, r *http.Request, err error, name string) {
	if !errors.Is(err, appziel.ErrNameRequired) {
		h.pageError(w, r, err)
		return
	}

	organization, getErr := h.organizations.Get(r.Context(), r.PathValue("id"))
	if getErr != nil {
		h.pageError(w, r, getErr)
		return
	}

	h.render(w, r, http.StatusUnprocessableEntity, h.formData(organization, name, "Bitte geben Sie einen Namen ein."))
}

func (h *Handler) pageData(organization domainorganisation.Organization, kind string) map[string]any {
	return map[string]any{
		"PageTitle":    "Ziele von " + organization.Name,
		"Navigation":   []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}},
		"Organization": map[string]string{"ID": organization.ID, "Name": organization.Name},
		"Errors":       []string{},
		"View":         map[string]any{"Kind": kind},
	}
}

func (h *Handler) formData(organization domainorganisation.Organization, name, fieldError string) map[string]any {
	data := h.pageData(organization, "form")
	view := data["View"].(map[string]any)
	view["Name"] = name
	view["FieldErrors"] = map[string]string{}
	if fieldError != "" {
		view["FieldErrors"] = map[string]string{"name": fieldError}
	}

	data["PageTitle"] = "Ziel anlegen"
	return data
}

func (h *Handler) projectGoals(goals []domainziel.Goal) []map[string]string {
	items := make([]map[string]string, 0, len(goals))
	for _, goal := range goals {
		items = append(items, map[string]string{"ID": goal.ID, "Name": goal.Name})
	}

	return items
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, appziel.ErrAccessDenied) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if errors.Is(err, appziel.ErrNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Organisation nicht gefunden")
		return
	}

	if errors.Is(err, appziel.ErrGoalNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Ziel nicht gefunden")
		return
	}

	h.pageFailure(w, r, http.StatusInternalServerError, "Die Seite ist derzeit nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := h.pageData(domainorganisation.Organization{}, "error")
	data["PageTitle"] = "Ziele"
	data["Errors"] = []string{message}
	h.render(w, r, status, data)
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, data map[string]any) {
	if h.bridge == nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	var output bytes.Buffer
	if err := h.bridge.Render(&output, h.templateName(r), data); err != nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}

func (h *Handler) templateName(r *http.Request) string {
	if r.Header.Get("HX-Request") == "true" {
		return "ziele/content"
	}

	return "ziele/page"
}
