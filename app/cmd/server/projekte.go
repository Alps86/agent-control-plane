package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webprojekt "agentcontrolplane/app/internal/adapter/web/projekt"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	appziel "agentcontrolplane/app/internal/app/ziel"
	"agentcontrolplane/ui/bridge"
)

func (b *Bootstrap) mountProjects(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, ui *bridge.Bridge) {
	tasks := b.mountTasks(server, db, organizations, ui)
	projects := webprojekt.NewHandler(appprojekt.NewService(db, organizations), organizations, appziel.NewService(db, organizations), ui, b.address())
	projects.SetArchive(appprojektarchiv.NewService(db, organizations))
	projects.SetTasks(tasks)
	for _, pattern := range []string{
		"GET /api/organisationen/{id}/projekte",
		"POST /api/organisationen/{id}/projekte",
		"GET /api/organisationen/{id}/projekte/{projektID}",
		"GET /organisationen/{id}/projekte",
		"GET /organisationen/{id}/projekte/neu",
		"POST /organisationen/{id}/projekte",
		"GET /organisationen/{id}/projekte/{projektID}",
	} {
		server.Handle(pattern, projects)
	}
}
