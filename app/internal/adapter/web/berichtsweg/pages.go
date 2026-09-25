package berichtsweg

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"

	appberichtsweg "agentcontrolplane/app/internal/app/berichtsweg"
	domainberichtsweg "agentcontrolplane/app/internal/domain/berichtsweg"
)

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	chart, err := h.service.Chart(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageFailure(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, h.pageData(chart))
}

func (h *Handler) pageAssign(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, appberichtsweg.ErrAccessDenied)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil || len(r.PostForm) != 1 || len(r.PostForm["parent_id"]) != 1 {
		h.renderFormError(w, r, http.StatusBadRequest, "Ungültige Formulardaten")
		return
	}

	_, err := h.service.Assign(r.Context(), r.PathValue("id"), r.PathValue("agentID"), r.PostFormValue("parent_id"))
	if err != nil {
		h.pageAssignError(w, r, err)
		return
	}

	http.Redirect(w, r, h.chartURL(r)+"#agent-"+url.PathEscape(r.PathValue("agentID")), http.StatusSeeOther)
}

func (h *Handler) pageAssignError(w http.ResponseWriter, r *http.Request, err error) {
	status, reason := h.errorStatus(err)
	if errors.Is(err, appberichtsweg.ErrAgentNotFound) {
		h.renderFormError(w, r, status, h.message(reason))
		return
	}

	if status == http.StatusNotFound || status == http.StatusInternalServerError {
		h.pageFailure(w, r, err)
		return
	}

	h.renderFormError(w, r, status, h.message(reason))
}

func (h *Handler) renderFormError(w http.ResponseWriter, r *http.Request, status int, message string) {
	chart, err := h.service.Chart(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageFailure(w, r, err)
		return
	}

	data := h.pageData(chart)
	data["Errors"] = []string{message}
	h.render(w, r, status, data)
}

func (h *Handler) message(reason string) string {
	if reason == "self_parent" {
		return "Selbstbezug: Ein Agent kann sich nicht selbst als Vorgesetzten haben."
	}

	if reason == "cycle" {
		return "Nachfahre und Zyklus: Ein Nachfahre kann kein Vorgesetzter werden."
	}

	if reason == "agent_not_found" {
		return "Fremde Organisation oder unbekannter Agent: Die gewählte Vorgesetzte gehört nicht zu dieser Organisation."
	}

	return "Der Berichtsweg konnte nicht gespeichert werden."
}

func (h *Handler) pageData(chart domainberichtsweg.Chart) map[string]any {
	roots, agents := h.nodes(chart.OrganizationID, chart.Lines)
	return map[string]any{
		"PageTitle": "Organigramm", "Notice": "", "Errors": []string{},
		"View": map[string]any{"Kind": "chart", "OrganizationID": chart.OrganizationID, "Roots": roots, "Agents": agents},
	}
}

func (h *Handler) nodes(organizationID string, lines []domainberichtsweg.Line) ([]*chartNode, []*chartNode) {
	byID := make(map[string]*chartNode, len(lines))
	agents := make([]*chartNode, 0, len(lines))
	for _, line := range lines {
		profileURL := "/organisationen/" + url.PathEscape(organizationID) + "/agenten/" + url.PathEscape(line.AgentID)
		node := &chartNode{ID: line.AgentID, Name: line.Name, ParentID: line.ParentID, ExecutionKind: h.kindLabel(string(line.ExecutionKind)), ProfileURL: profileURL}
		byID[line.AgentID] = node
		agents = append(agents, node)
	}

	return h.attachParents(agents, byID), agents
}

func (h *Handler) attachParents(agents []*chartNode, byID map[string]*chartNode) []*chartNode {
	roots := make([]*chartNode, 0, len(agents))
	for _, node := range agents {
		parent := byID[node.ParentID]
		if parent == nil || parent == node {
			roots = append(roots, node)
			continue
		}

		node.ParentName = parent.Name
		parent.Children = append(parent.Children, node)
	}

	return roots
}

func (h *Handler) kindLabel(kind string) string {
	if kind == "eino" {
		return "Eino"
	}

	if kind == "codex_cli" || kind == "codex" {
		return "Codex CLI"
	}

	return kind
}

func (h *Handler) chartURL(r *http.Request) string {
	return "/organisationen/" + url.PathEscape(r.PathValue("id")) + "/berichtswege"
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, err error) {
	status, _ := h.errorStatus(err)
	message := "Das Organigramm ist derzeit nicht verfügbar."
	if errors.Is(err, appberichtsweg.ErrAccessDenied) || errors.Is(err, appberichtsweg.ErrNotFound) {
		message = "Organisation nicht gefunden"
	}

	data := map[string]any{"PageTitle": "Organigramm", "Errors": []string{message}, "View": map[string]any{"Kind": "error"}}
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
		return "berichtswege/content"
	}

	return "berichtswege/page"
}
