package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webziel "agentcontrolplane/app/internal/adapter/web/ziel"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appziel "agentcontrolplane/app/internal/app/ziel"
	"agentcontrolplane/ui/bridge"
)

func (b *Bootstrap) mountGoals(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, ui *bridge.Bridge) {
	goals := webziel.NewHandler(appziel.NewService(db, organizations), organizations, ui, b.address())
	for _, pattern := range []string{
		"GET /api/organisationen/{id}/ziele",
		"POST /api/organisationen/{id}/ziele",
		"GET /organisationen/{id}/ziele",
		"POST /organisationen/{id}/ziele",
		"GET /organisationen/{id}/ziele/neu",
	} {
		server.Handle(pattern, goals)
	}
}
