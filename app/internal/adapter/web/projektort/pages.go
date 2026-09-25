package projektort

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	apport "agentcontrolplane/app/internal/app/projektort"
)

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	h.pageState(w, r, http.StatusOK, "", "")
}

func (h *Handler) pageBind(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageError(w, apport.ErrAccessDenied)
		return
	}

	if !h.validForm(w, r) {
		http.Error(w, "Ungültige Formulardaten", http.StatusBadRequest)
		return
	}

	location, err := h.locations.Bind(r.Context(), r.PathValue("id"), r.PathValue("projektID"), r.PostFormValue("kind"))
	if err != nil {
		h.pageError(w, err)
		return
	}
	if !location.StartReady {
		h.pageState(w, r, http.StatusConflict, "", location.Reason)
		return
	}

	http.Redirect(w, r, h.pageURL(r), http.StatusSeeOther)
}

func (h *Handler) pagePrepare(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageError(w, apport.ErrAccessDenied)
		return
	}

	result, err := h.locations.Prepare(r.Context(), r.PathValue("id"), r.PathValue("projektID"))
	if err != nil {
		h.pageError(w, err)
		return
	}

	if !result.Ready {
		status := http.StatusConflict
		if result.Code == apport.CodeAdapterUnavailable {
			status = http.StatusServiceUnavailable
		}
		h.pageState(w, r, status, "", result.Reason)
		return
	}

	h.pageState(w, r, http.StatusOK, "Startprüfung bestanden. Die Ordnerbindung ist bereit.", "")
}

func (h *Handler) validForm(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if r.ParseForm() != nil {
		return false
	}

	for key := range r.PostForm {
		if key != "kind" {
			return false
		}
	}

	return true
}

func (h *Handler) pageState(w http.ResponseWriter, r *http.Request, code int, notice, problem string) {
	data, err := h.pageData(r, notice, problem)
	if err != nil {
		h.pageError(w, err)
		return
	}

	h.render(w, code, h.templateName(r), data)
}

func (h *Handler) pageData(r *http.Request, notice, problem string) (map[string]any, error) {
	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return nil, err
	}

	project, err := h.projects.Find(r.Context(), organization.ID, r.PathValue("projektID"))
	if err != nil {
		return nil, err
	}

	location, err := h.locations.Get(r.Context(), organization.ID, project.ID)
	if err != nil {
		return nil, err
	}

	return h.view(organization.ID, organization.Name, project.ID, project.Name, location, notice, problem), nil
}

func (h *Handler) view(orgID, orgName, projectID, projectName string, location apport.Status, notice, problem string) map[string]any {
	errors := []string{}
	if problem != "" {
		errors = append(errors, problem)
	}

	return map[string]any{
		"PageTitle": "Ausführungsort · " + projectName, "Notice": notice, "Errors": errors,
		"Navigation":   []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}},
		"Organization": map[string]string{"ID": orgID, "Name": orgName},
		"View":         map[string]any{"Kind": "location", "Project": map[string]string{"ID": projectID, "Name": projectName}, "Location": h.locationView(location)},
	}
}

func (h *Handler) locationView(location apport.Status) map[string]any {
	path := ""
	if location.Path != nil {
		path = *location.Path
	}

	return map[string]any{
		"Configured": location.Configured, "Path": path, "StartReady": location.StartReady,
		"Code": location.Code, "Reason": location.Reason,
	}
}

func (h *Handler) render(w http.ResponseWriter, status int, name string, data map[string]any) {
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

func (h *Handler) templateName(r *http.Request) string {
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return "projekte/ausfuehrungsort/content"
	}

	return "projekte/ausfuehrungsort/page"
}

func (h *Handler) pageURL(r *http.Request) string {
	return "/organisationen/" + url.PathEscape(r.PathValue("id")) + "/projekte/" + url.PathEscape(r.PathValue("projektID")) + "/ausfuehrungsort"
}

func (h *Handler) pageError(w http.ResponseWriter, err error) {
	status, _ := h.publicError(err)
	http.Error(w, http.StatusText(status), status)
}
