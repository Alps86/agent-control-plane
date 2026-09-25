package modellwahl

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"

	app "agentcontrolplane/app/internal/app/modellwahl"
	domain "agentcontrolplane/app/internal/domain/modellwahl"
)

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.Get(r.Context(), r.PathValue("id"), r.PathValue("agentID"))
	if err != nil {
		h.pageError(w, err)
		return
	}

	h.render(w, http.StatusOK, h.pageData(r.PathValue("id"), view, nil))
}

func (h *Handler) pagePost(w http.ResponseWriter, r *http.Request) {
	if !h.localWrite(r) {
		h.pageFailure(w, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		h.pageFailure(w, http.StatusBadRequest, "Ungültige Formulardaten")
		return
	}

	selection := domain.Selection{Provider: r.PostFormValue("provider"),
		ConnectionReference: r.PostFormValue("connection_reference"), Model: r.PostFormValue("model")}
	h.savePage(w, r, selection)
}

func (h *Handler) savePage(w http.ResponseWriter, r *http.Request, selection domain.Selection) {
	_, err := h.service.Put(r.Context(), r.PathValue("id"), r.PathValue("agentID"), selection)
	if err != nil {
		h.pageError(w, err)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func (h *Handler) pageData(organizationID string, view app.View, fieldErrors map[string]string) map[string]any {
	items := make([]map[string]any, 0, len(view.Providers))
	for _, provider := range view.Providers {
		items = append(items, h.providerItem(provider))
	}

	return map[string]any{"PageTitle": "Modellwahl von " + view.Agent.Name,
		"Navigation": h.navigation(organizationID), "Notice": "", "Errors": []string{},
		"AllowedActions": []string{"agent.model.select"}, "View": map[string]any{
			"Kind": "model-choice", "OrganizationID": organizationID,
			"Agent":     map[string]any{"ID": view.Agent.ID, "Name": view.Agent.Name, "ExecutionKindLabel": "Eino"},
			"Selection": map[string]any{"Provider": view.Selection.Provider, "ConnectionReference": view.Selection.ConnectionReference, "Model": view.Selection.Model},
			"Providers": items, "FieldErrors": fieldErrors}}
}

func (h *Handler) navigation(organizationID string) []map[string]string {
	return []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"},
		{"Key": "agenten", "Label": "Agenten", "Href": "/organisationen/" + url.PathEscape(organizationID) + "/agenten", "Icon": "◇"}}
}

func (h *Handler) providerItem(provider domain.Provider) map[string]any {
	models := make([]map[string]any, 0, len(provider.Models))
	for _, model := range provider.Models {
		models = append(models, h.modelItem(model))
	}

	return map[string]any{"ID": provider.ID, "Name": provider.Name, "Selectable": provider.Selectable,
		"Reason": provider.Reason, "AuthType": provider.AuthType, "Connections": provider.Connections, "Models": models}
}

func (h *Handler) modelItem(model domain.Model) map[string]any {
	capabilities := make([]map[string]any, 0, 5)
	for _, name := range []string{"text", "tools", "audio_input", "audio_output", "realtime"} {
		capabilities = append(capabilities, h.capabilityItem(name, model.Capabilities[name]))
	}

	return map[string]any{"ID": model.ID, "Label": h.modelLabel(model), "Source": model.Source,
		"ObservedAt": model.ObservedAt, "CheckStatus": model.CheckStatus,
		"AccountVerified": model.AccountVerified, "RouteVerified": model.RouteVerified,
		"Capabilities": capabilities}
}

func (h *Handler) modelLabel(model domain.Model) string {
	if model.Label != "" {
		return model.Label
	}

	return model.ID
}

func (h *Handler) capabilityItem(name string, capability domain.Capability) map[string]any {
	status := capability.Status
	if status == "" {
		status = "Nicht nachgewiesen"
	}

	return map[string]any{"Name": name, "Status": status, "Source": capability.Source, "CheckedAt": capability.CheckedAt}
}

func (h *Handler) pageError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	if errors.Is(err, app.ErrNotFound) {
		h.pageFailure(w, http.StatusNotFound, "Nicht gefunden")
		return
	}

	if errors.Is(err, app.ErrCatalog) || errors.Is(err, app.ErrExecutionKind) {
		h.pageFailure(w, http.StatusUnprocessableEntity, "Modellroute nicht im Katalog")
		return
	}

	if h.grantError(err) {
		h.pageFailure(w, http.StatusForbidden, h.publicGrantReason(err))
		return
	}

	h.pageFailure(w, http.StatusInternalServerError, "Die Modellwahl ist derzeit nicht verfügbar")
}

func (h *Handler) publicGrantReason(err error) string {
	if errors.Is(err, app.ErrAccessDenied) {
		return "Zugriff verweigert"
	}

	return "Modellverbindung nicht freigegeben"
}

func (h *Handler) pageFailure(w http.ResponseWriter, status int, message string) {
	h.render(w, status, map[string]any{"PageTitle": "Modellwahl", "Errors": []string{message},
		"View": map[string]any{"Kind": "error"}})
}

func (h *Handler) render(w http.ResponseWriter, status int, data map[string]any) {
	if h.ui == nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	var output bytes.Buffer
	if err := h.ui.Render(&output, "agenten/modellwahl/page", data); err != nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}
