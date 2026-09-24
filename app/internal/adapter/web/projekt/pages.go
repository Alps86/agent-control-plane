package projekt

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appziel "agentcontrolplane/app/internal/app/ziel"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
)

func (h *Handler) pageList(w http.ResponseWriter, r *http.Request) {
	organization, goals, err := h.organizationGoals(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	projects, err := h.projects.List(r.Context(), organization.ID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	data := h.pageData(organization, "list")
	data["View"].(map[string]any)["Projects"] = h.projectList(projects, goals)
	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) pageForm(w http.ResponseWriter, r *http.Request) {
	organization, goals, err := h.organizationGoals(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, h.formData(organization, goals, createRequest{}, nil))
}

func (h *Handler) pageCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	request, message := h.formRequest(w, r)
	if message != "" {
		h.pageFailure(w, r, http.StatusBadRequest, message)
		return
	}

	project, err := h.projects.Create(r.Context(), r.PathValue("id"), request.Name, request.Description, request.GoalID)
	if err != nil {
		h.formError(w, r, err, request)
		return
	}

	http.Redirect(w, r, "/organisationen/"+project.OrganizationID+"/projekte", http.StatusSeeOther)
}

func (h *Handler) formRequest(w http.ResponseWriter, r *http.Request) (createRequest, string) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		return createRequest{}, "Formulardaten konnten nicht gelesen werden."
	}

	for key := range r.PostForm {
		if key != "name" && key != "description" && key != "goal_id" {
			return createRequest{}, "Unbekanntes Formularfeld"
		}
	}

	return createRequest{Name: r.PostFormValue("name"), Description: r.PostFormValue("description"), GoalID: r.PostFormValue("goal_id")}, ""
}

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	organization, goals, err := h.organizationGoals(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	project, err := h.projects.Find(r.Context(), organization.ID, r.PathValue("projektID"))
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	data := h.pageData(organization, "detail")
	data["PageTitle"] = project.Name
	view := data["View"].(map[string]any)
	view["Project"] = h.projectView(project, h.goalName(goals, project.GoalID))
	view["Tasks"] = []map[string]string{}
	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) organizationGoals(r *http.Request) (domainorganisation.Organization, []domainziel.Goal, error) {
	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return domainorganisation.Organization{}, nil, err
	}

	goals, err := h.goals.List(r.Context(), organization.ID)
	return organization, goals, err
}

func (h *Handler) formError(w http.ResponseWriter, r *http.Request, err error, request createRequest) {
	fieldErrors := h.projectFieldErrors(err)
	if fieldErrors == nil {
		h.pageError(w, r, err)
		return
	}

	organization, goals, contextErr := h.organizationGoals(r)
	if contextErr != nil {
		h.pageError(w, r, contextErr)
		return
	}

	h.render(w, r, http.StatusUnprocessableEntity, h.formData(organization, goals, request, fieldErrors))
}

func (h *Handler) projectFieldErrors(err error) map[string]string {
	if errors.Is(err, appprojekt.ErrNameRequired) {
		return map[string]string{"name": "Bitte geben Sie einen Namen ein."}
	}

	if errors.Is(err, appprojekt.ErrGoalRequired) || errors.Is(err, appprojekt.ErrInvalidGoal) {
		return map[string]string{"goal_id": "Bitte wählen Sie ein Ziel dieser Organisation aus."}
	}

	return nil
}

func (h *Handler) pageData(organization domainorganisation.Organization, kind string) map[string]any {
	return map[string]any{
		"PageTitle":    "Projekte von " + organization.Name,
		"Navigation":   []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}},
		"Organization": map[string]string{"ID": organization.ID, "Name": organization.Name},
		"Errors":       []string{},
		"View":         map[string]any{"Kind": kind},
	}
}

func (h *Handler) formData(organization domainorganisation.Organization, goals []domainziel.Goal, request createRequest, fieldErrors map[string]string) map[string]any {
	data := h.pageData(organization, "form")
	data["PageTitle"] = "Projekt anlegen"
	if fieldErrors == nil {
		fieldErrors = map[string]string{}
	}

	view := data["View"].(map[string]any)
	view["Name"] = request.Name
	view["Description"] = request.Description
	view["GoalID"] = request.GoalID
	view["Goals"] = h.goalOptions(goals, request.GoalID)
	view["FieldErrors"] = fieldErrors
	return data
}

func (h *Handler) goalOptions(goals []domainziel.Goal, selectedID string) []map[string]any {
	options := make([]map[string]any, 0, len(goals))
	for _, goal := range goals {
		options = append(options, map[string]any{"ID": goal.ID, "Name": goal.Name, "Selected": goal.ID == selectedID})
	}

	return options
}

func (h *Handler) projectList(projects []domainprojekt.Project, goals []domainziel.Goal) []map[string]string {
	items := make([]map[string]string, 0, len(projects))
	for _, project := range projects {
		items = append(items, h.projectView(project, h.goalName(goals, project.GoalID)))
	}

	return items
}

func (h *Handler) projectView(project domainprojekt.Project, goalName string) map[string]string {
	return map[string]string{"ID": project.ID, "Name": project.Name, "Description": project.Description, "GoalID": project.GoalID, "GoalName": goalName}
}

func (h *Handler) goalName(goals []domainziel.Goal, goalID string) string {
	for _, goal := range goals {
		if goal.ID == goalID {
			return goal.Name
		}
	}

	return "Ziel nicht verfügbar"
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, appprojekt.ErrAccessDenied) || errors.Is(err, apporganisation.ErrAccessDenied) || errors.Is(err, appziel.ErrAccessDenied) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if errors.Is(err, appprojekt.ErrNotFound) || errors.Is(err, apporganisation.ErrNotFound) || errors.Is(err, appziel.ErrNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Projekt oder Organisation nicht gefunden")
		return
	}

	h.pageFailure(w, r, http.StatusInternalServerError, "Die Seite ist derzeit nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := h.pageData(domainorganisation.Organization{}, "error")
	data["PageTitle"] = "Projekte"
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
		return "projekte/content"
	}

	return "projekte/page"
}
