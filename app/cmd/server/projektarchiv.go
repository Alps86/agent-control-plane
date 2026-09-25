package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webprojektarchiv "agentcontrolplane/app/internal/adapter/web/projektarchiv"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	appziel "agentcontrolplane/app/internal/app/ziel"
	"agentcontrolplane/ui/bridge"
)

// mountProjectArchive ergänzt Statusaktionen und Archivansicht.
func (b *Bootstrap) mountProjectArchive(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, ui *bridge.Bridge) {
	archive := appprojektarchiv.NewService(db, organizations)
	handler := webprojektarchiv.NewHandler(archive, organizations, appziel.NewService(db, organizations), ui, b.address())
	for _, pattern := range []string{
		"GET /api/organisationen/{id}/projekte/archiv",
		"POST /api/organisationen/{id}/projekte/{projektID}/archivieren",
		"POST /api/organisationen/{id}/projekte/{projektID}/wiederherstellen",
		"GET /api/organisationen/{id}/projekte/{projektID}/status",
		"GET /organisationen/{id}/projekte/archiv",
		"POST /organisationen/{id}/projekte/{projektID}/archivieren",
		"POST /organisationen/{id}/projekte/{projektID}/restaurieren",
	} {
		server.Handle(pattern, handler)
	}
}
