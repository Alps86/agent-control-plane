package main

import (
	"bytes"
	"net"
	"net/http"
	"strings"

	"agentcontrolplane/app/internal/adapter/web"
	"agentcontrolplane/ui/bridge"
)

func (b *Bootstrap) mountSettings(server *web.Server, ui *bridge.Bridge) {
	server.Handle("GET /settings", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.settingsPage(w, r, ui)
	}))
}

func (b *Bootstrap) settingsPage(w http.ResponseWriter, r *http.Request, ui *bridge.Bridge) {
	if !b.localSettingsRequest(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	name := "modelle/settings/page"
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		name = "modelle/settings/content"
	}

	b.renderPage(w, ui, name, map[string]any{"PageTitle": "Settings · Modellanbieter"}, http.StatusOK)
}

func (b *Bootstrap) renderPage(w http.ResponseWriter, ui *bridge.Bridge, name string, data map[string]any, code int) {
	var page bytes.Buffer
	if err := ui.Render(&page, name, data); err != nil {
		http.Error(w, "Ansicht derzeit nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_, _ = w.Write(page.Bytes())
}

func (b *Bootstrap) localSettingsRequest(r *http.Request) bool {
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || net.ParseIP(peer) == nil || !net.ParseIP(peer).IsLoopback() {
		return false
	}

	return b.localSettingsHost(r.Host)
}

func (b *Bootstrap) localSettingsHost(authority string) bool {
	host, port, err := net.SplitHostPort(authority)
	_, bindPort, bindErr := net.SplitHostPort(b.address())
	if err != nil || bindErr != nil || port != bindPort {
		return false
	}

	address := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || address != nil && address.IsLoopback()
}
