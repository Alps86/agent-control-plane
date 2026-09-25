package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webberichtsweg "agentcontrolplane/app/internal/adapter/web/berichtsweg"
	appberichtsweg "agentcontrolplane/app/internal/app/berichtsweg"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
)

// mountReportingLines exposes the persistent organization chart.
func (b *Bootstrap) mountReportingLines(server *web.Server, db *sqlite.Database, ui *bridge.Bridge) {
	service := appberichtsweg.NewService(db, db, sqlite.NewOrganizationStore(db), apporganisation.NewLocalIdentity())
	handler := webberichtsweg.NewHandler(service, ui, b.address())
	for _, pattern := range []string{
		"/api/organisationen/{id}/berichtswege",
		"/api/organisationen/{id}/agenten/{agentID}/berichtsweg",
		"/organisationen/{id}/berichtswege",
		"/organisationen/{id}/agenten/{agentID}/berichtsweg",
	} {
		server.Handle(pattern, handler)
	}
}
