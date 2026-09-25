package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webaktivitaet "agentcontrolplane/app/internal/adapter/web/aktivitaet"
	appaktivitaet "agentcontrolplane/app/internal/app/aktivitaet"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
)

// mountActivity bindet den organisationsgebundenen Aktivitätsverlauf an den Server.
func (b *Bootstrap) mountActivity(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, ui *bridge.Bridge) *appaktivitaet.Service {
	service := appaktivitaet.NewService(db, organizations, apporganisation.NewLocalIdentity())
	handler := webaktivitaet.NewHandler(service, organizations, ui)
	server.Handle("GET /api/organisationen/{id}/aktivitaet", handler)
	server.Handle("GET /api/organisationen/{id}/aktivitaet/export.csv", handler)
	server.Handle("GET /organisationen/{id}/aktivitaet", handler)
	return service
}
