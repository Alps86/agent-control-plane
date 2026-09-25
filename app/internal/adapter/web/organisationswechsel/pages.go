package organisationswechsel

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"strings"

	appregel "agentcontrolplane/app/internal/app/organisationsregel"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
)

func (h *Handler) pageSearch(w http.ResponseWriter, r *http.Request) {
	organizations, err := h.search(r)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	data := h.pageData("Organisation wählen", "list")
	view := data["View"].(map[string]any)
	view["Query"] = r.URL.Query().Get("q")
	view["Organizations"] = h.projectOrganizations(organizations, h.selectedOrganization(r))
	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) projectOrganizations(organizations []domainorganisation.Organization, selected string) []map[string]any {
	items := make([]map[string]any, 0, len(organizations))
	for _, organization := range organizations {
		items = append(items, map[string]any{
			"ID": organization.ID, "Name": organization.Name,
			"Description": organization.Description, "Selected": selected == organization.ID,
		})
	}

	return items
}

func (h *Handler) pageSelect(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.selectOrganization(w, r, organization.ID)
	http.Redirect(w, r, "/organisationen/"+organization.ID, http.StatusSeeOther)
}

func (h *Handler) pageRules(w http.ResponseWriter, r *http.Request) {
	data, err := h.rulesData(r, appregel.Change{})
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) rulesData(r *http.Request, change appregel.Change) (map[string]any, error) {
	snapshot, err := h.rules.Read(r.Context(), r.PathValue("id"))
	if err != nil {
		return nil, err
	}

	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return nil, err
	}

	data := h.pageData("Arbeitsregeln von "+organization.Name, "rules")
	h.populateRules(data, organization, snapshot, change)
	return data, nil
}

func (h *Handler) populateRules(data map[string]any, organization domainorganisation.Organization, snapshot appregel.Snapshot, change appregel.Change) {
	view := data["View"].(map[string]any)
	view["Organization"] = map[string]string{"ID": organization.ID, "Name": organization.Name}
	view["Revision"] = snapshot.Revision
	view["Change"] = change
	view["Areas"] = []map[string]any{
		h.area(snapshot, "status", "Aufgabenstatus", "Erlaubte Statusübergänge für Aufgaben."),
		h.area(snapshot, "assignment", "Zuweisung", "Erlaubte Übergänge bei der Zuständigkeit."),
		h.area(snapshot, "delegation", "Delegation", "Erlaubte Übergänge für delegierte Arbeit."),
	}
}

func (h *Handler) area(snapshot appregel.Snapshot, key, label, description string) map[string]any {
	rules := make([]appregel.Rule, 0)
	for _, rule := range snapshot.Rules {
		if rule.Area == key {
			rules = append(rules, rule)
		}
	}

	return map[string]any{"Label": label, "Description": description, "Rules": rules}
}

func (h *Handler) pagePreview(w http.ResponseWriter, r *http.Request) {
	change, ok := h.formChange(w, r)
	if !ok {
		h.pageFailure(w, r, http.StatusBadRequest, "Formulardaten konnten nicht gelesen werden.")
		return
	}

	data, err := h.rulesData(r, change)
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	h.showPreview(w, r, data, change)
}

func (h *Handler) showPreview(w http.ResponseWriter, r *http.Request, data map[string]any, change appregel.Change) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	preview, err := h.rules.Preview(r.Context(), r.PathValue("id"), change)
	if err != nil {
		h.pageRuleError(w, r, data, err)
		return
	}

	view := data["View"].(map[string]any)
	view["Preview"] = h.projectPreview(preview)
	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) projectPreview(preview appregel.Preview) map[string]any {
	label := map[string]string{"status": "Aufgabenstatus", "assignment": "Zuweisung", "delegation": "Delegation"}[preview.Area]
	return map[string]any{
		"AreaLabel": label, "AffectedFlows": preview.AffectedFlows,
		"Rule": preview.Rule, "Revision": preview.Revision, "Operation": preview.Operation,
	}
}

func (h *Handler) pageSave(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	change, ok := h.formChange(w, r)
	if !ok {
		h.pageFailure(w, r, http.StatusBadRequest, "Formulardaten konnten nicht gelesen werden.")
		return
	}

	change.Operation = "allow"
	h.saveForm(w, r, change)
}

func (h *Handler) pageRevoke(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	change, ok := h.formChange(w, r)
	if !ok {
		h.pageFailure(w, r, http.StatusBadRequest, "Formulardaten konnten nicht gelesen werden.")
		return
	}

	change.Operation = "revoke"
	h.saveForm(w, r, change)
}

func (h *Handler) saveForm(w http.ResponseWriter, r *http.Request, change appregel.Change) {
	_, err := h.rules.Save(r.Context(), r.PathValue("id"), change)
	if err != nil {
		h.saveFormError(w, r, change, err)
		return
	}

	http.Redirect(w, r, "/organisationen/"+r.PathValue("id")+"/arbeitsregeln", http.StatusSeeOther)
}

func (h *Handler) saveFormError(w http.ResponseWriter, r *http.Request, change appregel.Change, err error) {
	data, loadErr := h.rulesData(r, change)
	if loadErr != nil {
		h.pageError(w, r, loadErr)
		return
	}

	h.pageRuleError(w, r, data, err)
}

func (h *Handler) formChange(w http.ResponseWriter, r *http.Request) (appregel.Change, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		return appregel.Change{}, false
	}

	revision, err := strconv.ParseInt(r.PostFormValue("revision"), 10, 64)
	if err != nil {
		return appregel.Change{}, false
	}

	return appregel.Change{
		Area: r.PostFormValue("area"), From: r.PostFormValue("from"),
		To: r.PostFormValue("to"), Approver: r.PostFormValue("approver"), Revision: revision,
		Operation: r.PostFormValue("operation"),
	}, true
}

func (h *Handler) pageData(title, kind string) map[string]any {
	return map[string]any{
		"PageTitle": title, "Notice": "", "Errors": []string{},
		"View": map[string]any{"Kind": kind},
	}
}

func (h *Handler) pageRuleError(w http.ResponseWriter, r *http.Request, data map[string]any, err error) {
	status, code := h.errorStatus(err)
	if status == http.StatusNotFound || status == http.StatusForbidden {
		h.pageError(w, r, err)
		return
	}

	data["Errors"] = []string{h.ruleMessage(code)}
	h.render(w, r, status, data)
}

func (h *Handler) ruleMessage(code string) string {
	messages := map[string]string{
		"revision_conflict":  "Die Arbeitsregeln wurden inzwischen geändert. Laden Sie die Seite erneut.",
		"invalid_transition": "Dieser Übergang ist für den gewählten Bereich nicht zulässig.",
		"approver_required":  "Wählen Sie die zuständige Freigabe aus.",
		"rights_expansion":   "Die Änderung würde Rechte über die erlaubte Grenze hinaus erweitern.",
		"workflow_dsl":       "Freie Workflow-Anweisungen sind nicht zulässig.",
		"rule_not_found":     "Diese Regel ist nicht mehr gespeichert. Laden Sie die Seite erneut.",
		"invalid_operation":  "Diese Regelaktion ist nicht zulässig.",
	}
	if message, ok := messages[code]; ok {
		return message
	}

	return "Die Arbeitsregel konnte nicht verarbeitet werden."
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	status, _ := h.errorStatus(err)
	if errors.Is(err, appregel.ErrNotFound) {
		status = http.StatusNotFound
	}

	h.pageFailure(w, r, status, "Organisation nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := h.pageData("Organisation nicht verfügbar", "error")
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
		return "organisationswechsel/content"
	}

	return "organisationswechsel/page"
}
