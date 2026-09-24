package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webdatenbereich "agentcontrolplane/app/internal/adapter/web/datenbereich"
	appdatenbereich "agentcontrolplane/app/internal/app/datenbereich"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
)

// mountDataScope binds the operator's agent scope editor to the server.
func (b *Bootstrap) mountDataScope(server *web.Server, db *sqlite.Database, ui *bridge.Bridge) {
	service := appdatenbereich.NewService(db, apporganisation.NewLocalIdentity(), nil)
	handler := webdatenbereich.NewHandler(service, ui, b.address())
	server.Handle("/api/organisationen/{id}/agenten/{agentID}/datenbereich", handler)
	server.Handle("/organisationen/{id}/agenten/{agentID}/datenbereich", handler)
}
