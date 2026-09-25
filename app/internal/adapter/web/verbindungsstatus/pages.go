package verbindungsstatus

import (
	"bytes"
	"net/http"
	"strings"

	app "agentcontrolplane/app/internal/app/verbindungsstatus"
)

func (h *Handler) pageGet(w http.ResponseWriter, r *http.Request) {
	if h.ui == nil {
		http.Error(w, "Ansicht derzeit nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	var output bytes.Buffer
	if h.ui.Render(&output, h.template(r), h.pageData(h.service.Get(r.Context()))) != nil {
		http.Error(w, "Ansicht derzeit nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = output.WriteTo(w)
}

func (h *Handler) template(r *http.Request) string {
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return "modelle/status/content"
	}

	return "modelle/status/page"
}

func (h *Handler) pageData(view app.View) map[string]any {
	providers := make([]map[string]any, 0, len(view.Providers))
	for _, provider := range view.Providers {
		providers = append(providers, h.providerData(provider))
	}

	return map[string]any{"PageTitle": "Settings · Modellanbieter", "View": map[string]any{"Providers": providers}}
}

func (h *Handler) providerData(provider app.Provider) map[string]any {
	return map[string]any{"ID": provider.ID, "Name": provider.Name,
		"AuthMethod": h.authMethod(provider.AuthType), "Status": provider.Status,
		"StatusDetail": provider.StatusDetail, "Reference": provider.ConnectionReference,
		"Ready": provider.Ready, "ManageURL": h.manageURL(provider.ID)}
}

func (h *Handler) authMethod(kind string) string {
	if kind == "device_code" {
		return "Gerätecode"
	}

	if kind == "api_key" {
		return "API-Schlüssel"
	}

	return kind
}

func (h *Handler) manageURL(id string) string {
	if id == "codex-abo" {
		return "/settings/modelle/codex"
	}

	if id == "openrouter" {
		return "/settings/modellanbieter/openrouter"
	}

	return ""
}
