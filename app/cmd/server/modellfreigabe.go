package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webfreigabe "agentcontrolplane/app/internal/adapter/web/modellfreigabe"
	appfreigabe "agentcontrolplane/app/internal/app/modellfreigabe"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	portfreigabe "agentcontrolplane/app/internal/port/modellfreigabe"
	"agentcontrolplane/ui/bridge"
)

// mountModelGrants bindet die Freigaben an dieselbe zentrale Verbindung wie Settings.
func (b *Bootstrap) mountModelGrants(server *web.Server, db *sqlite.Database, ui *bridge.Bridge, connection portfreigabe.ConnectionStatus) *appfreigabe.Service {
	service := appfreigabe.NewService(sqlite.NewModellfreigabeStore(db), sqlite.NewOrganizationStore(db), db, apporganisation.NewLocalIdentity(), connection)
	handler := webfreigabe.NewHandler(service, ui, b.address())
	for _, pattern := range []string{
		"/organisationen/{id}/modellfreigabe/openrouter", "/organisationen/{id}/modellfreigabe/openrouter/",
		"/api/organisationen/{id}/modellfreigabe/openrouter", "/api/organisationen/{id}/modellfreigabe/openrouter/",
	} {
		server.Handle(pattern, handler)
	}
	return service
}
