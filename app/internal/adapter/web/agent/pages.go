package agent

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"

	appagent "agentcontrolplane/app/internal/app/agent"
	domainagent "agentcontrolplane/app/internal/domain/agent"
)

func (h *Handler) pageList(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.service.List(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, err)
		return
	}
	data := h.pageData(r.PathValue("id"), "Agentenübersicht", "list")
	items := make([]map[string]any, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, h.projectAgent(profile))
	}
	data["View"].(map[string]any)["Agents"] = items
	data["AllowedActions"] = []string{"agent.create"}
	h.render(w, http.StatusOK, data)
}

func (h *Handler) pageForm(w http.ResponseWriter, r *http.Request) {
	data, err := h.formData(r, r.URL.Query().Get("vorlage"), r.URL.Query().Get("team"), "", nil)
	if err != nil {
		h.pageError(w, err)
		return
	}
	h.render(w, http.StatusOK, data)
}

func (h *Handler) pageCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, http.StatusForbidden, "Zugriff verweigert")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		h.pageFailure(w, http.StatusBadRequest, "Formulardaten konnten nicht gelesen werden.")
		return
	}
	input := appagent.CreateInput{Name: r.PostFormValue("name"), TemplateID: r.PostFormValue("template_id"), ExecutionKind: domainagent.ExecutionKind(r.PostFormValue("execution_kind"))}
	profile, err := h.service.Create(r.Context(), r.PathValue("id"), input)
	if err != nil {
		h.pageFormError(w, r, input, err)
		return
	}
	http.Redirect(w, r, "/organisationen/"+url.PathEscape(r.PathValue("id"))+"/agenten/"+url.PathEscape(profile.ID), http.StatusSeeOther)
}

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	profile, err := h.service.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.pageError(w, err)
		return
	}

	data, err := h.profileData(r, profile)
	if err != nil {
		h.pageError(w, err)
		return
	}

	h.renderProfile(w, http.StatusOK, data, profile)
}

func (h *Handler) pageStart(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, http.StatusForbidden, "Zugriff verweigert")
		return
	}
	profile, err := h.service.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.pageError(w, err)
		return
	}

	h.pageStartProfile(w, r, profile)
}

func (h *Handler) pageStartProfile(w http.ResponseWriter, r *http.Request, profile appagent.Profile) {
	data, err := h.profileData(r, profile)
	if err != nil {
		h.pageError(w, err)
		return
	}

	if !profile.Readiness.Ready {
		data["Notice"] = h.startReason(profile.Readiness)
		h.renderProfile(w, http.StatusConflict, data, profile)
		return
	}

	data["Notice"] = "Ein Agentenlauf ist noch nicht verfügbar."
	h.renderProfile(w, http.StatusNotImplemented, data, profile)
}

func (h *Handler) formData(r *http.Request, templateID, teamID, name string, fieldErrors map[string]string) (map[string]any, error) {
	organizationID := r.PathValue("id")
	templates, teams, err := h.formCatalog(r)
	if err != nil {
		return nil, err
	}

	data := h.pageData(organizationID, "Agent anlegen", "form")
	data["AllowedActions"] = []string{"agent.create"}
	view := data["View"].(map[string]any)
	view["Name"] = name
	if fieldErrors == nil {
		fieldErrors = map[string]string{}
	}

	view["FieldErrors"] = fieldErrors
	h.formTemplates(view, templates, templateID)
	h.formTeams(view, teams, templates, teamID)
	return data, nil
}

func (h *Handler) formCatalog(r *http.Request) ([]domainagent.Template, []domainagent.TeamTemplate, error) {
	organizationID := r.PathValue("id")
	templates, err := h.service.Templates(r.Context(), organizationID)
	if err != nil {
		return nil, nil, err
	}

	teams, err := h.service.TeamTemplates(r.Context(), organizationID)
	return templates, teams, err
}

func (h *Handler) formTemplates(view map[string]any, templates []domainagent.Template, templateID string) {
	einoItems := make([]map[string]any, 0, len(templates))
	codexItems := make([]map[string]any, 0, len(templates))
	for _, template := range templates {
		item := h.projectTemplate(template)
		switch template.Kind {
		case domainagent.Eino:
			einoItems = append(einoItems, item)
		case domainagent.CodexCLI:
			codexItems = append(codexItems, item)
		}

		if template.ID == templateID {
			view["SelectedTemplate"] = item
		}
	}

	view["EinoTemplates"] = einoItems
	view["CodexTemplates"] = codexItems
}

func (h *Handler) formTeams(view map[string]any, teams []domainagent.TeamTemplate, templates []domainagent.Template, teamID string) {
	teamItems := make([]map[string]any, 0, len(teams))
	view["TeamTemplatesSelected"] = []map[string]any{}
	for _, team := range teams {
		teamItems = append(teamItems, map[string]any{"ID": team.ID, "Name": team.Name})
		if team.ID == teamID {
			view["SelectedTeam"] = map[string]any{"ID": team.ID, "Name": team.Name}
			view["TeamTemplatesSelected"] = h.teamTemplates(team, templates)
		}
	}

	view["TeamTemplates"] = teamItems
}

func (h *Handler) teamTemplates(team domainagent.TeamTemplate, templates []domainagent.Template) []map[string]any {
	selected := make([]map[string]any, 0, len(team.TemplateIDs))
	for _, template := range templates {
		for _, selectedID := range team.TemplateIDs {
			if template.ID == selectedID {
				selected = append(selected, h.projectTemplate(template))
			}
		}
	}

	return selected
}

func (h *Handler) pageFormError(w http.ResponseWriter, r *http.Request, input appagent.CreateInput, err error) {
	status, fieldErrors, known := h.formFieldError(err)
	if !known {
		h.pageError(w, err)
		return
	}

	data, loadErr := h.formData(r, input.TemplateID, "", input.Name, fieldErrors)
	if loadErr != nil {
		h.pageError(w, loadErr)
		return
	}

	h.render(w, status, data)
}

func (h *Handler) formFieldError(err error) (int, map[string]string, bool) {
	switch {
	case errors.Is(err, appagent.ErrNameRequired):
		return http.StatusUnprocessableEntity, map[string]string{"name": "Bitte geben Sie einen Namen ein."}, true
	case errors.Is(err, appagent.ErrNameConflict):
		return http.StatusConflict, map[string]string{"name": "Dieser Name ist bereits vergeben."}, true
	case errors.Is(err, appagent.ErrTemplateNotFound):
		return http.StatusUnprocessableEntity, map[string]string{"template_id": "Vorlage nicht gefunden."}, true
	case errors.Is(err, appagent.ErrExecutionKind):
		return http.StatusUnprocessableEntity, map[string]string{"template_id": "Ausführungsart passt nicht zur Vorlage."}, true
	case errors.Is(err, appagent.ErrCapabilityDenied):
		return http.StatusUnprocessableEntity, map[string]string{"template_id": "Fachfähigkeit ist nicht zulässig."}, true
	default:
		return 0, nil, false
	}
}

func (h *Handler) detailData(organizationID string, profile appagent.Profile) map[string]any {
	data := h.pageData(organizationID, profile.Name, "detail")
	data["View"].(map[string]any)["Agent"] = h.projectAgent(profile)
	return data
}

func (h *Handler) profileData(r *http.Request, profile appagent.Profile) (map[string]any, error) {
	organizationID := r.PathValue("id")
	organization, err := h.service.Organization(r.Context(), organizationID)
	if err != nil {
		return nil, err
	}

	data := h.detailData(organizationID, profile)
	data["View"].(map[string]any)["OrganizationName"] = organization.Name
	return data, nil
}

func (h *Handler) projectAgent(profile appagent.Profile) map[string]any {
	return map[string]any{
		"ID": profile.ID, "OrganizationID": profile.OrganizationID,
		"TemplateID": profile.TemplateID, "Name": profile.Name, "Role": profile.Role,
		"Instructions": profile.Instructions, "Capabilities": profile.Capabilities,
		"ExecutionKind": profile.ExecutionKind, "ExecutionKindLabel": h.executionKindLabel(profile.ExecutionKind),
		"ModelStatus": h.modelStatus(profile.ExecutionKind), "ReadinessLabel": h.readinessLabel(profile.Readiness),
		"ReadinessCode": profile.Readiness.Code, "ReadinessReady": profile.Readiness.Ready,
		"ReadinessReason": h.startReason(profile.Readiness),
	}
}

func (h *Handler) modelStatus(kind domainagent.ExecutionKind) string {
	if kind == domainagent.Eino {
		return "Modell nicht verbunden"
	}

	return ""
}

func (h *Handler) readinessLabel(readiness domainagent.Readiness) string {
	if readiness.Ready {
		return "Startbereit"
	}

	return "Nicht startbereit"
}

func (h *Handler) projectTemplate(template domainagent.Template) map[string]any {
	return map[string]any{
		"ID": template.ID, "Name": template.Name, "Role": template.Role,
		"Instructions": template.Instructions, "ExecutionKind": template.Kind,
		"ExecutionKindLabel": h.executionKindLabel(template.Kind), "Capabilities": template.Capabilities,
	}
}

func (h *Handler) executionKindLabel(kind domainagent.ExecutionKind) string {
	if kind == domainagent.Eino {
		return "Eino"
	}

	if kind == domainagent.CodexCLI {
		return "Codex CLI"
	}

	return string(kind)
}

func (h *Handler) pageData(organizationID, title, kind string) map[string]any {
	return map[string]any{
		"PageTitle":  title,
		"Navigation": []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}, {"Key": "agenten", "Label": "Agenten", "Href": "/organisationen/" + url.PathEscape(organizationID) + "/agenten", "Icon": "◇"}},
		"Notice":     "", "Errors": []string{}, "AllowedActions": []string{},
		"View": map[string]any{"Kind": kind, "OrganizationID": organizationID},
	}
}

func (h *Handler) pageError(w http.ResponseWriter, err error) {
	if errors.Is(err, appagent.ErrAccessDenied) {
		h.pageFailure(w, http.StatusForbidden, "Zugriff verweigert")
		return
	}
	if errors.Is(err, appagent.ErrNotFound) {
		h.pageFailure(w, http.StatusNotFound, "Nicht gefunden")
		return
	}
	h.pageFailure(w, http.StatusInternalServerError, "Die Seite ist derzeit nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, status int, message string) {
	data := h.pageData("", "Agenten", "error")
	data["Errors"] = []string{message}
	h.render(w, status, data)
}

func (h *Handler) render(w http.ResponseWriter, status int, data map[string]any) {
	h.renderNamed(w, status, "agenten/page", data)
}

func (h *Handler) renderProfile(w http.ResponseWriter, status int, data map[string]any, profile appagent.Profile) {
	name := "agenten/page"
	if profile.ExecutionKind == domainagent.CodexCLI {
		name = "agenten/codex/page"
	}
	h.renderNamed(w, status, name, data)
}

func (h *Handler) renderNamed(w http.ResponseWriter, status int, name string, data map[string]any) {
	if h.bridge == nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}
	var output bytes.Buffer
	if err := h.bridge.Render(&output, name, data); err != nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}
