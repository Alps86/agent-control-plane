package modellfreigabe

import (
	"bytes"
	"net/http"
	"strings"

	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
)

func (h *Handler) pageShow(w http.ResponseWriter, r *http.Request) {
	h.showPage(w, r, "")
}

func (h *Handler) pageGrantOrganization(w http.ResponseWriter, r *http.Request) {
	h.pageOrganizationWrite(w, r, true)
}

func (h *Handler) pageRevokeOrganization(w http.ResponseWriter, r *http.Request) {
	h.pageOrganizationWrite(w, r, false)
}

func (h *Handler) pageOrganizationWrite(w http.ResponseWriter, r *http.Request, grant bool) {
	var err error
	if grant {
		err = h.service.GrantOrganization(r.Context(), r.PathValue("id"))
	}
	if !grant {
		err = h.service.RevokeOrganization(r.Context(), r.PathValue("id"))
	}
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.showPage(w, r, "Freigabe gespeichert")
}

func (h *Handler) pageGrantAgent(w http.ResponseWriter, r *http.Request) {
	h.pageAgentWrite(w, r, true)
}

func (h *Handler) pageRevokeAgent(w http.ResponseWriter, r *http.Request) {
	h.pageAgentWrite(w, r, false)
}

func (h *Handler) pageAgentWrite(w http.ResponseWriter, r *http.Request, grant bool) {
	var err error
	if grant {
		err = h.service.GrantAgent(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	}
	if !grant {
		err = h.service.RevokeAgent(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	}
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.showPage(w, r, "Freigabe gespeichert")
}

func (h *Handler) showPage(w http.ResponseWriter, r *http.Request, notice string) {
	view, err := h.service.Overview(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, h.pageData(view, notice))
}

func (h *Handler) pageData(view appfreigabe.Overview, notice string) map[string]any {
	agents := make([]map[string]any, 0, len(view.Agents))
	for _, agent := range view.Agents {
		agents = append(agents, map[string]any{"ID": agent.ID, "Name": agent.Name, "Granted": agent.Status.AgentGranted})
	}

	return map[string]any{"PageTitle": "OpenRouter-Freigaben · " + view.OrganizationName,
		"Notice": notice, "Errors": []string{}, "View": map[string]any{
			"Reference": view.Status.Reference, "ConnectionStatus": view.Status.ConnectionStatus,
			"Organization": map[string]any{"ID": view.OrganizationID, "Name": view.OrganizationName, "Granted": view.Status.OrganizationGranted},
			"Agents":       agents,
		}}
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	status, _, message := h.failure(err)
	h.pageFailure(w, r, status, message)
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := map[string]any{"PageTitle": "OpenRouter-Freigaben", "Notice": "", "Errors": []string{message},
		"View": map[string]any{"Reference": "", "ConnectionStatus": "nicht verfügbar",
			"Organization": map[string]any{"ID": "", "Name": "Organisation", "Granted": false}, "Agents": []any{}}}
	h.render(w, r, status, data)
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, data map[string]any) {
	if h.renderer == nil {
		http.Error(w, "Ansicht derzeit nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	var output bytes.Buffer
	if h.renderer.Render(&output, h.templateName(r), data) != nil {
		http.Error(w, "Ansicht derzeit nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}

func (h *Handler) templateName(r *http.Request) string {
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return "modelle/freigabe/content"
	}

	return "modelle/freigabe/page"
}
