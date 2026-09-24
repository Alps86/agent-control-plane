package codexprofil

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"

	appagent "agentcontrolplane/app/internal/app/agent"
	appcodexprofil "agentcontrolplane/app/internal/app/codexprofil"
	domainprofil "agentcontrolplane/app/internal/domain/codexprofil"
)

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	profile, err := h.profiles.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.pageError(w, err)
		return
	}

	h.pageProfile(w, r, http.StatusOK, profile, "", "")
}

func (h *Handler) pageSave(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageError(w, appcodexprofil.ErrAccessDenied)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ungültige Formulardaten", http.StatusBadRequest)
		return
	}

	input := domainprofil.Input{WorkspaceEnabled: r.PostForm.Has("workspace_enabled"), WriteEnabled: r.PostForm.Has("write_enabled")}
	h.pageSaveInput(w, r, input)
}

func (h *Handler) pageSaveInput(w http.ResponseWriter, r *http.Request, input domainprofil.Input) {
	_, err := h.profiles.Save(r.Context(), r.PathValue("id"), r.PathValue("agentID"), input)
	if errors.Is(err, appcodexprofil.ErrWriteWithoutWorkspace) {
		h.pageWriteError(w, r)
		return
	}

	if err != nil {
		h.pageError(w, err)
		return
	}

	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		h.pageGet(w, r)
		return
	}

	http.Redirect(w, r, h.pageURL(r), http.StatusSeeOther)
}

func (h *Handler) pageWriteError(w http.ResponseWriter, r *http.Request) {
	profile, err := h.profiles.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.pageError(w, err)
		return
	}

	h.pageProfile(w, r, http.StatusUnprocessableEntity, profile, "", "Schreiben benötigt einen aktivierten Arbeitsbereich.")
}

func (h *Handler) pageProfile(w http.ResponseWriter, r *http.Request, status int, profile domainprofil.Profile, notice, fieldError string) {
	data, err := h.pageData(r, profile, notice, fieldError)
	if err != nil {
		h.pageError(w, err)
		return
	}

	name := "agenten/codexprofil/page"
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		name = "agenten/codexprofil/content"
	}

	h.render(w, status, name, data)
}

func (h *Handler) pageData(r *http.Request, profile domainprofil.Profile, notice, fieldError string) (map[string]any, error) {
	if h.agents == nil {
		return nil, appagent.ErrAccessDenied
	}

	agent, err := h.agents.Get(r.Context(), profile.OrganizationID, profile.AgentID)
	if err != nil {
		return nil, err
	}

	organization, err := h.agents.Organization(r.Context(), profile.OrganizationID)
	if err != nil {
		return nil, err
	}

	return h.view(profile, agent.Name, organization.Name, agent.Readiness.Reason, notice, fieldError), nil
}

func (h *Handler) view(profile domainprofil.Profile, agentName, organizationName, reason, notice, fieldError string) map[string]any {
	return map[string]any{
		"PageTitle": agentName + " · Codex-Arbeitsprofil", "Notice": notice,
		"Navigation": []map[string]string{{"Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}, {"Label": "Agenten", "Href": "/organisationen/" + url.PathEscape(profile.OrganizationID) + "/agenten", "Icon": "◇"}},
		"View": map[string]any{
			"OrganizationID": profile.OrganizationID, "AgentID": profile.AgentID,
			"AgentName": agentName, "OrganizationName": organizationName,
			"WorkspaceEnabled": profile.WorkspaceEnabled, "WriteEnabled": profile.WriteEnabled,
			"ReadinessReason": reason, "WriteError": fieldError,
		},
	}
}

func (h *Handler) pageURL(r *http.Request) string {
	return "/organisationen/" + url.PathEscape(r.PathValue("id")) + "/agenten/" + url.PathEscape(r.PathValue("agentID")) + "/codex-profil"
}

func (h *Handler) pageError(w http.ResponseWriter, err error) {
	status, _ := h.publicError(err)
	if errors.Is(err, appagent.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, appagent.ErrAccessDenied) {
		status = http.StatusForbidden
	}

	http.Error(w, http.StatusText(status), status)
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
