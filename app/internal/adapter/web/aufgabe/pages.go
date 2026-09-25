package aufgabe

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"

	appagent "agentcontrolplane/app/internal/app/agent"
	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	domainagent "agentcontrolplane/app/internal/domain/agent"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	domainprojektarchiv "agentcontrolplane/app/internal/domain/projektarchiv"
)

func (h *Handler) pageList(w http.ResponseWriter, r *http.Request) {
	organization, project, err := h.context(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	tasks, err := h.tasks.List(r.Context(), organization.ID, project.ID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.renderTaskList(w, r, organization, project, tasks)
}

func (h *Handler) renderTaskList(w http.ResponseWriter, r *http.Request, organization domainorganisation.Organization, project domainprojekt.Project, tasks []domainaufgabe.Task) {
	status, err := h.archive.Status(r.Context(), organization.ID, project.ID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	data := h.pageData(organization, project, "list")
	data["View"].(map[string]any)["ProjectStatus"] = status
	data["View"].(map[string]any)["Tasks"] = h.taskList(tasks)
	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) pageForm(w http.ResponseWriter, r *http.Request) {
	organization, project, err := h.context(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	if h.rejectArchived(w, r, organization.ID, project.ID) {
		return
	}

	h.renderForm(w, r, http.StatusOK, organization, project, createRequest{}, nil)
}

func (h *Handler) pageCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	request, fields, message := h.formRequest(w, r)
	if message != "" {
		h.pageFailure(w, r, http.StatusBadRequest, message)
		return
	}

	h.createFormTask(w, r, request, fields)
}

func (h *Handler) createFormTask(w http.ResponseWriter, r *http.Request, request createRequest, fields map[string]string) {
	organization, project, err := h.context(r)
	if err != nil {
		h.formContextError(w, r, request, err)
		return
	}
	if h.rejectArchived(w, r, organization.ID, project.ID) {
		return
	}

	fields = h.projectFields(r, project, fields)
	if fields != nil {
		h.renderForm(w, r, http.StatusUnprocessableEntity, organization, project, request, fields)
		return
	}

	h.saveFormTask(w, r, organization, project, request)
}

func (h *Handler) projectFields(r *http.Request, project domainprojekt.Project, fields map[string]string) map[string]string {
	if r.PostFormValue("project_id") == project.ID && r.PostFormValue("project_name") == project.Name {
		return fields
	}

	if fields == nil {
		fields = map[string]string{}
	}

	fields["project_id"] = "Bitte wählen Sie ein gültiges Projekt aus."
	return fields
}

func (h *Handler) formContextError(w http.ResponseWriter, r *http.Request, request createRequest, err error) {
	if !errors.Is(err, appprojekt.ErrNotFound) {
		h.pageError(w, r, err)
		return
	}

	organization, orgErr := h.organizations.Get(r.Context(), r.PathValue("id"))
	if orgErr != nil {
		h.pageError(w, r, orgErr)
		return
	}

	h.renderForm(w, r, http.StatusUnprocessableEntity, organization, domainprojekt.Project{}, request,
		map[string]string{"project_id": "Bitte wählen Sie ein gültiges Projekt aus."})
}

func (h *Handler) saveFormTask(w http.ResponseWriter, r *http.Request, organization domainorganisation.Organization, project domainprojekt.Project, request createRequest) {
	task, err := h.tasks.Create(r.Context(), organization.ID, h.input(r, request))
	if err != nil {
		h.formError(w, r, organization, project, request, err)
		return
	}

	path := "/organisationen/" + url.PathEscape(organization.ID) + "/projekte/" + url.PathEscape(project.ID) + "/aufgaben/" + url.PathEscape(task.ID)
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func (h *Handler) formRequest(w http.ResponseWriter, r *http.Request) (createRequest, map[string]string, string) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		return createRequest{}, nil, "Formulardaten konnten nicht gelesen werden."
	}

	for key, values := range r.PostForm {
		if !h.allowedField(key) {
			return createRequest{}, nil, "Unbekanntes Formularfeld"
		}

		if len(values) != 1 {
			return h.duplicateField(r, key)
		}
	}

	return h.requestFromForm(r), nil, ""
}

func (h *Handler) duplicateField(r *http.Request, key string) (createRequest, map[string]string, string) {
	if key == "assignee_id" {
		return h.requestFromForm(r), map[string]string{"assignee_id": "Bitte wählen Sie genau einen Agenten aus."}, ""
	}

	if key == "project_id" {
		return h.requestFromForm(r), map[string]string{"project_id": "Bitte wählen Sie ein gültiges Projekt aus."}, ""
	}

	return createRequest{}, nil, "Formularfeld wurde mehrfach übermittelt."
}

func (h *Handler) allowedField(key string) bool {
	return key == "title" || key == "description" || key == "priority" || key == "assignee_id" || key == "project_id" || key == "project_name"
}

func (h *Handler) requestFromForm(r *http.Request) createRequest {
	return createRequest{Title: r.PostFormValue("title"), Description: r.PostFormValue("description"), Priority: r.PostFormValue("priority"), AssigneeID: r.PostFormValue("assignee_id")}
}

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	organization, project, err := h.context(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	task, err := h.find(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.renderDetail(w, r, organization, project, task)
}

func (h *Handler) renderDetail(w http.ResponseWriter, r *http.Request, organization domainorganisation.Organization, project domainprojekt.Project, task domainaufgabe.Task) {
	profile, err := h.agents.Get(r.Context(), organization.ID, task.AssigneeID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	data := h.pageData(organization, project, "detail")
	data["PageTitle"] = task.Title
	data["View"].(map[string]any)["Task"] = h.taskView(task, profile.Name, h.kindLabel(profile.ExecutionKind))
	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) context(r *http.Request) (domainorganisation.Organization, domainprojekt.Project, error) {
	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return domainorganisation.Organization{}, domainprojekt.Project{}, err
	}

	project, err := h.projects.Find(r.Context(), organization.ID, r.PathValue("projektID"))
	return organization, project, err
}

func (h *Handler) formError(w http.ResponseWriter, r *http.Request, organization domainorganisation.Organization, project domainprojekt.Project, request createRequest, err error) {
	if errors.Is(err, appaufgabe.ErrProjectArchived) {
		h.pageFailure(w, r, http.StatusConflict, "Archiviertes Projekt nimmt keine neue Arbeit an")
		return
	}

	fields := h.fieldErrors(err)
	if fields == nil {
		h.pageError(w, r, err)
		return
	}

	h.renderForm(w, r, http.StatusUnprocessableEntity, organization, project, request, fields)
}

func (h *Handler) renderForm(w http.ResponseWriter, r *http.Request, status int, organization domainorganisation.Organization, project domainprojekt.Project, request createRequest, fields map[string]string) {
	profiles, err := h.agents.List(r.Context(), organization.ID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	data := h.formData(organization, project, request, profiles, fields)
	h.render(w, r, status, data)
}

func (h *Handler) formData(organization domainorganisation.Organization, project domainprojekt.Project, request createRequest, profiles []appagent.Profile, fields map[string]string) map[string]any {
	data := h.pageData(organization, project, "form")
	data["PageTitle"] = "Aufgabe anlegen"
	view := data["View"].(map[string]any)
	view["Title"], view["Description"], view["Priority"] = request.Title, request.Description, request.Priority
	view["FieldErrors"], view["Agents"] = fields, h.agentOptions(profiles, request.AssigneeID)
	view["Priorities"] = h.priorityOptions(request.Priority)
	return data
}

func (h *Handler) agentOptions(profiles []appagent.Profile, selected string) []map[string]any {
	options := make([]map[string]any, 0, len(profiles))
	for _, profile := range profiles {
		if profile.Status != domainagent.StatusActive {
			continue
		}

		options = append(options, map[string]any{"ID": profile.ID, "Name": profile.Name, "ExecutionKindLabel": h.kindLabel(profile.ExecutionKind), "Selected": profile.ID == selected})
	}

	return options
}

func (h *Handler) priorityOptions(selected string) []map[string]any {
	if selected == "" {
		selected = "normal"
	}

	options := make([]map[string]any, 0, 4)
	for _, value := range []string{"low", "normal", "high", "urgent"} {
		options = append(options, map[string]any{"Value": value, "Label": h.priorityLabel(value), "Selected": value == selected})
	}

	return options
}

func (h *Handler) taskList(tasks []domainaufgabe.Task) []map[string]string {
	items := make([]map[string]string, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, h.taskView(task, "", ""))
	}

	return items
}

func (h *Handler) taskView(task domainaufgabe.Task, assigneeName, kind string) map[string]string {
	return map[string]string{"ID": task.ID, "Title": task.Title, "Description": task.Description,
		"PriorityLabel": h.priorityLabel(task.Priority), "StatusLabel": h.statusLabel(task.Status),
		"AssigneeName": assigneeName, "ExecutionKindLabel": kind}
}

func (h *Handler) priorityLabel(value string) string {
	labels := map[string]string{"low": "Niedrig", "normal": "Normal", "high": "Hoch", "urgent": "Dringend"}
	return labels[value]
}

func (h *Handler) statusLabel(value string) string {
	if value == "open" {
		return "Offen"
	}

	return value
}

func (h *Handler) kindLabel(kind domainagent.ExecutionKind) string {
	if kind == domainagent.CodexCLI {
		return "Codex CLI"
	}

	return "Eino"
}

func (h *Handler) pageData(organization domainorganisation.Organization, project domainprojekt.Project, kind string) map[string]any {
	return map[string]any{"PageTitle": "Aufgaben", "Organization": map[string]string{"ID": organization.ID, "Name": organization.Name},
		"Navigation": []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}},
		"Errors":     []string{}, "View": map[string]any{"Kind": kind, "Project": project}}
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, appprojektarchiv.ErrNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Aufgabe oder Projekt nicht gefunden")
		return
	}

	if errors.Is(err, appaufgabe.ErrAccessDenied) || errors.Is(err, appprojekt.ErrAccessDenied) || errors.Is(err, apporganisation.ErrAccessDenied) || errors.Is(err, appagent.ErrAccessDenied) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if errors.Is(err, appaufgabe.ErrNotFound) || errors.Is(err, appprojekt.ErrNotFound) || errors.Is(err, apporganisation.ErrNotFound) || errors.Is(err, appagent.ErrNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Aufgabe oder Projekt nicht gefunden")
		return
	}

	h.pageFailure(w, r, http.StatusInternalServerError, "Die Seite ist derzeit nicht verfügbar")
}

func (h *Handler) rejectArchived(w http.ResponseWriter, r *http.Request, organizationID, projectID string) bool {
	status, err := h.archive.Status(r.Context(), organizationID, projectID)
	if err != nil {
		h.pageError(w, r, err)
		return true
	}

	if status == domainprojektarchiv.StatusArchived {
		h.pageFailure(w, r, http.StatusConflict, "Archiviertes Projekt nimmt keine neue Arbeit an")
		return true
	}

	return false
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := h.pageData(domainorganisation.Organization{}, domainprojekt.Project{}, "error")
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
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return "aufgaben/content"
	}

	return "aufgaben/page"
}
