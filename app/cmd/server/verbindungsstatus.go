package main

import (
	"net/http"

	"agentcontrolplane/app/internal/adapter/web"
	webstatus "agentcontrolplane/app/internal/adapter/web/verbindungsstatus"
	appstatus "agentcontrolplane/app/internal/app/verbindungsstatus"
	"agentcontrolplane/ui/bridge"
)

func (b *Bootstrap) mountConnectionStatus(server *web.Server, ui *bridge.Bridge, service *appstatus.Service) {
	handler := webstatus.NewHandler(service, ui)
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !b.localSettingsRequest(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		handler.ServeHTTP(w, r)
	})
	server.Handle("GET /api/settings/modellanbieter/status", protected)
	server.Handle("GET /settings/modellanbieter/status", protected)
}
