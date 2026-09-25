package projektarchiv

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainprojektarchiv "agentcontrolplane/app/internal/domain/projektarchiv"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
)

func (h *Handler) pageList(w http.ResponseWriter, r *http.Request) {
	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	projects, err := h.archive.ListArchived(r.Context(), organization.ID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	goals, err := h.goals.List(r.Context(), organization.ID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, h.pageData(organization, h.projectViews(projects, goals), ""))
}

func (h *Handler) pageArchive(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageError(w, r, apporganisation.ErrAccessDenied)
		return
	}

	if err := h.archive.Archive(r.Context(), r.PathValue("id"), r.PathValue("projektID")); err != nil {
		h.pageError(w, r, err)
		return
	}

	http.Redirect(w, r, "/organisationen/"+r.PathValue("id")+"/projekte", http.StatusSeeOther)
}

func (h *Handler) pageRestore(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageError(w, r, apporganisation.ErrAccessDenied)
		return
	}

	if err := h.archive.Restore(r.Context(), r.PathValue("id"), r.PathValue("projektID")); err != nil {
		h.pageError(w, r, err)
		return
	}

	http.Redirect(w, r, "/organisationen/"+r.PathValue("id")+"/projekte", http.StatusSeeOther)
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	status, message := http.StatusInternalServerError, "Projektarchiv ist derzeit nicht verfügbar"
	if errors.Is(err, apporganisation.ErrAccessDenied) {
		status, message = http.StatusForbidden, "Zugriff verweigert"
	}

	if errors.Is(err, apporganisation.ErrNotFound) || errors.Is(err, appprojektarchiv.ErrNotFound) {
		status, message = http.StatusNotFound, "Projekt oder Organisation nicht gefunden"
	}

	if errors.Is(err, appprojektarchiv.ErrActiveRun) {
		status, message = http.StatusConflict, "Projekt mit aktivem Lauf kann nicht archiviert werden"
	}

	if errors.Is(err, appprojektarchiv.ErrAlreadyArchived) || errors.Is(err, appprojektarchiv.ErrNotArchived) {
		status, message = http.StatusConflict, "Projektstatus erlaubt diese Aktion nicht"
	}

	h.render(w, r, status, h.pageData(domainorganisation.Organization{}, nil, message))
}

func (h *Handler) pageData(organization domainorganisation.Organization, projects []map[string]string, message string) map[string]any {
	kind := "list"
	if message != "" {
		kind = "error"
	}

	return map[string]any{
		"PageTitle":    "Projektarchiv",
		"Organization": map[string]string{"ID": organization.ID, "Name": organization.Name},
		"Navigation":   []map[string]string{{"Key": "projekte", "Label": "Projekte", "Href": "/organisationen/" + organization.ID + "/projekte", "Icon": "▦"}},
		"Errors":       h.messages(message),
		"View":         map[string]any{"Kind": kind, "Projects": projects},
	}
}

func (h *Handler) projectViews(projects []domainprojektarchiv.ArchivedProject, goals []domainziel.Goal) []map[string]string {
	views := make([]map[string]string, 0, len(projects))
	for _, project := range projects {
		views = append(views, map[string]string{"ID": project.ID, "Name": project.Name,
			"Description": project.Description, "GoalID": project.GoalID,
			"GoalName": h.goalName(goals, project.GoalID), "Status": "archived"})
	}

	return views
}

func (h *Handler) goalName(goals []domainziel.Goal, goalID string) string {
	for _, goal := range goals {
		if goal.ID == goalID {
			return goal.Name
		}
	}

	return "Ziel nicht verfügbar"
}

func (h *Handler) messages(message string) []string {
	if message != "" {
		return []string{message}
	}

	return []string{}
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
		return "projektarchiv/content"
	}

	return "projektarchiv/page"
}
