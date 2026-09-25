package kommentar

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"

	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
)

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	h.threadPage(w, r, http.StatusOK, "", nil)
}

func (h *Handler) pageCreate(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	content, err := h.formContent(w, r)
	if err != nil {
		h.pageFailure(w, r, http.StatusBadRequest, "Formulardaten konnten nicht gelesen werden")
		return
	}

	h.savePageComment(w, r, content)
}

func (h *Handler) formContent(w http.ResponseWriter, r *http.Request) (string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		return "", err
	}

	if len(r.PostForm) != 1 || len(r.PostForm["content"]) != 1 {
		return "", errors.New("unexpected form fields")
	}

	return r.PostForm["content"][0], nil
}

func (h *Handler) savePageComment(w http.ResponseWriter, r *http.Request, content string) {
	if _, err := h.find(r); err != nil {
		h.pageError(w, r, err)
		return
	}

	_, err := h.comments.Create(r.Context(), r.PathValue("id"), r.PathValue("aufgabeID"), appkommentar.CreateInput{Content: content})
	if err != nil {
		h.createPageError(w, r, content, err)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func (h *Handler) createPageError(w http.ResponseWriter, r *http.Request, content string, err error) {
	if errors.Is(err, appkommentar.ErrContentRequired) {
		h.threadPage(w, r, http.StatusUnprocessableEntity, content, map[string]string{"content": "Bitte geben Sie einen Kommentar ein."})
		return
	}

	h.pageError(w, r, err)
}

func (h *Handler) threadPage(w http.ResponseWriter, r *http.Request, status int, content string, fields map[string]string) {
	organization, project, task, err := h.pageContext(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	comments, err := h.comments.List(r.Context(), organization.ID, task.ID)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, status, h.threadData(r, organization, project, task, comments, content, fields))
}

func (h *Handler) pageContext(r *http.Request) (domainorganisation.Organization, domainprojekt.Project, domainaufgabe.Task, error) {
	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return domainorganisation.Organization{}, domainprojekt.Project{}, domainaufgabe.Task{}, err
	}

	project, err := h.projects.Find(r.Context(), organization.ID, r.PathValue("projektID"))
	if err != nil {
		return organization, project, domainaufgabe.Task{}, err
	}

	task, err := h.find(r)
	return organization, project, task, err
}

func (h *Handler) threadData(r *http.Request, organization domainorganisation.Organization, project domainprojekt.Project, task domainaufgabe.Task, comments []domainkommentar.Comment, content string, fields map[string]string) map[string]any {
	data := h.pageData(organization, "thread")
	data["PageTitle"] = "Kommentare zu " + task.Title
	data["View"] = map[string]any{"Kind": "thread", "Project": project,
		"Task":     map[string]string{"ID": task.ID, "Title": task.Title},
		"Comments": h.commentViews(r, comments), "Content": content, "FieldErrors": fields}
	return data
}

func (h *Handler) commentViews(r *http.Request, comments []domainkommentar.Comment) []map[string]any {
	views := make([]map[string]any, 0, len(comments))
	for _, comment := range comments {
		views = append(views, map[string]any{"ID": comment.ID, "Content": comment.Content,
			"SourceLabel": h.sourceLabel(comment), "Timestamp": comment.CreatedAt,
			"References": h.referenceViews(r, comment)})
	}

	return views
}

func (h *Handler) sourceLabel(comment domainkommentar.Comment) string {
	if comment.SourceName != "" {
		return comment.SourceName
	}

	if comment.SourceKind == domainkommentar.SourceOperator {
		return "Betreiberin"
	}

	return "Agent"
}

func (h *Handler) referenceViews(r *http.Request, comment domainkommentar.Comment) []map[string]string {
	if comment.Reference == nil {
		return nil
	}

	reference := comment.Reference
	view := map[string]string{"ID": reference.ID, "OrganizationID": comment.OrganizationID,
		"TypeLabel": h.referenceLabel(reference.Type)}
	if reference.Type == domainkommentar.ReferenceArtifact {
		view["DisplayName"], view["Link"] = reference.DisplayName, h.safeArtifactLink(reference.Link)
	}

	if reference.Type == domainkommentar.ReferenceTask {
		view["Link"] = h.taskReferenceLink(r, comment.OrganizationID, reference.ID)
	}

	return []map[string]string{view}
}

func (h *Handler) referenceLabel(kind string) string {
	if kind == domainkommentar.ReferenceArtifact {
		return "Artefakt"
	}

	return "Aufgabe"
}

func (h *Handler) taskReferenceLink(r *http.Request, orgID, taskID string) string {
	task, err := h.tasks.Find(r.Context(), orgID, taskID)
	if err != nil {
		return ""
	}

	return "/organisationen/" + url.PathEscape(orgID) + "/projekte/" + url.PathEscape(task.ProjectID) + "/aufgaben/" + url.PathEscape(task.ID)
}

func (h *Handler) safeArtifactLink(link string) string {
	parsed, err := url.Parse(link)
	if err != nil || !strings.HasPrefix(link, "/") || strings.HasPrefix(link, "//") || strings.Contains(link, "\\") {
		return ""
	}

	if parsed.IsAbs() || parsed.Host != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return ""
	}

	return link
}

func (h *Handler) pageData(organization domainorganisation.Organization, kind string) map[string]any {
	return map[string]any{"PageTitle": "Aufgabenkommentare", "Organization": map[string]string{"ID": organization.ID, "Name": organization.Name},
		"Navigation": []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}},
		"Errors":     []string{}, "View": map[string]any{"Kind": kind}}
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, appaufgabe.ErrNotFound) || errors.Is(err, appkommentar.ErrNotFound) || errors.Is(err, appprojekt.ErrNotFound) || errors.Is(err, apporganisation.ErrNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Aufgabe oder Projekt nicht gefunden")
		return
	}

	if errors.Is(err, appkommentar.ErrAccessDenied) || errors.Is(err, appaufgabe.ErrAccessDenied) || errors.Is(err, appprojekt.ErrAccessDenied) || errors.Is(err, apporganisation.ErrAccessDenied) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	h.pageFailure(w, r, http.StatusInternalServerError, "Die Seite ist derzeit nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := h.pageData(domainorganisation.Organization{}, "error")
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
		return "kommentare/content"
	}

	return "kommentare/page"
}
