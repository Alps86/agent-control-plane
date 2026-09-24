package datenbereich

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"

	appdatenbereich "agentcontrolplane/app/internal/app/datenbereich"
	domaindatenbereich "agentcontrolplane/app/internal/domain/datenbereich"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
)

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	data, err := h.pageData(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) pageData(r *http.Request) (map[string]any, error) {
	organizationID, agentID := r.PathValue("id"), r.PathValue("agentID")
	agent, err := h.service.Agent(r.Context(), organizationID, agentID)
	if err != nil {
		return nil, err
	}

	projects, err := h.service.Projects(r.Context(), organizationID, agentID)
	if err != nil {
		return nil, err
	}

	scopes, err := h.service.List(r.Context(), organizationID, agentID)
	if err != nil {
		return nil, err
	}

	items, summary := h.projectItems(projects, scopes)
	return map[string]any{"PageTitle": "Datenbereich von " + agent.Name, "Navigation": h.navigation(organizationID), "Notice": "", "Errors": []string{}, "AllowedActions": []string{"agent.scope.update"}, "View": map[string]any{"Kind": "scope", "OrganizationID": organizationID, "Agent": map[string]string{"ID": agent.ID, "Name": agent.Name}, "Projects": items, "HasGrants": len(scopes) > 0, "GrantSummary": strings.Join(summary, ", ")}}, nil
}

func (h *Handler) navigation(organizationID string) []map[string]string {
	return []map[string]string{
		{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"},
		{"Key": "agenten", "Label": "Agenten", "Href": "/organisationen/" + url.PathEscape(organizationID) + "/agenten", "Icon": "◇"},
	}
}

func (h *Handler) projectItems(projects []domainprojekt.Project, scopes []domaindatenbereich.Scope) ([]map[string]any, []string) {
	items := make([]map[string]any, 0, len(projects))
	summary := []string{}
	for _, project := range projects {
		item := h.projectItem(project, scopes)
		items = append(items, item)
		if item["CanRead"].(bool) || item["CanWrite"].(bool) {
			summary = append(summary, project.Name+" "+item["AccessLabel"].(string))
		}
	}

	return items, summary
}

func (h *Handler) projectItem(project domainprojekt.Project, scopes []domaindatenbereich.Scope) map[string]any {
	read, write := false, false
	for _, scope := range scopes {
		if scope.ProjectID == project.ID {
			read, write = scope.CanRead, scope.CanWrite
		}
	}

	return map[string]any{"ID": project.ID, "Name": project.Name, "CanRead": read, "CanWrite": write, "AccessLabel": h.accessLabel(read, write)}
}

func (h *Handler) accessLabel(read, write bool) string {
	if read && write {
		return "mit Lesen und Schreiben"
	}

	if read {
		return "mit Lesen und ohne Schreiben"
	}

	if write {
		return "ohne Lesen und mit Schreiben"
	}

	return "ohne Lesen und Schreiben"
}

func (h *Handler) pageReplace(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	inputs, err := h.formInputs(w, r)
	if err != nil {
		h.pageFailure(w, r, http.StatusBadRequest, "Ungültige Formulardaten")
		return
	}

	h.saveForm(w, r, inputs)
}

func (h *Handler) saveForm(w http.ResponseWriter, r *http.Request, inputs []appdatenbereich.ScopeInput) {
	_, err := h.service.Replace(r.Context(), r.PathValue("id"), r.PathValue("agentID"), inputs)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	http.Redirect(w, r, h.pageURL(r), http.StatusSeeOther)
}

func (h *Handler) formInputs(w http.ResponseWriter, r *http.Request) ([]appdatenbereich.ScopeInput, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	for key := range r.PostForm {
		if key != "read_project_ids" && key != "write_project_ids" {
			return nil, errors.New("unknown form field")
		}
	}

	return h.scopeInputs(r.PostForm), nil
}

func (h *Handler) scopeInputs(values url.Values) []appdatenbereich.ScopeInput {
	byProject := map[string]*appdatenbereich.ScopeInput{}
	for _, id := range values["read_project_ids"] {
		input := h.scopeInput(byProject, id)
		input.CanRead = true
	}

	for _, id := range values["write_project_ids"] {
		input := h.scopeInput(byProject, id)
		input.CanWrite = true
	}

	return h.orderedInputs(byProject)
}

func (h *Handler) scopeInput(inputs map[string]*appdatenbereich.ScopeInput, id string) *appdatenbereich.ScopeInput {
	if inputs[id] == nil {
		inputs[id] = &appdatenbereich.ScopeInput{ProjectID: id}
	}

	return inputs[id]
}

func (h *Handler) orderedInputs(inputs map[string]*appdatenbereich.ScopeInput) []appdatenbereich.ScopeInput {
	ids := make([]string, 0, len(inputs))
	for id := range inputs {
		ids = append(ids, id)
	}

	sort.Strings(ids)
	result := make([]appdatenbereich.ScopeInput, 0, len(ids))
	for _, id := range ids {
		result = append(result, *inputs[id])
	}

	return result
}

func (h *Handler) pageURL(r *http.Request) string {
	return "/organisationen/" + url.PathEscape(r.PathValue("id")) + "/agenten/" + url.PathEscape(r.PathValue("agentID")) + "/datenbereich"
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, appdatenbereich.ErrNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Nicht gefunden")
		return
	}

	if errors.Is(err, appdatenbereich.ErrAccessDenied) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if errors.Is(err, appdatenbereich.ErrInvalidScope) {
		h.pageFailure(w, r, http.StatusUnprocessableEntity, "Ungültiger Datenbereich")
		return
	}

	h.pageFailure(w, r, http.StatusInternalServerError, "Die Seite ist derzeit nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := map[string]any{"PageTitle": "Datenbereich", "Navigation": []any{}, "Notice": "", "Errors": []string{message}, "AllowedActions": []string{}, "View": map[string]any{"Kind": "error"}}
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
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return "agenten/datenbereich/content"
	}

	return "agenten/datenbereich/page"
}
