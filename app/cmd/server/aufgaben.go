package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webaufgabe "agentcontrolplane/app/internal/adapter/web/aufgabe"
	appagent "agentcontrolplane/app/internal/app/agent"
	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"
	"agentcontrolplane/ui/bridge"
)

// mountTasks bindet Aufgaben-API und Aufgabenoberfläche an den Server.
func (b *Bootstrap) mountTasks(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, ui *bridge.Bridge) *appaufgabe.Service {
	projects := appprojekt.NewService(db, organizations)
	agents := appagent.NewService(db, apporganisation.NewLocalIdentity(), sqlite.NewOrganizationStore(db))
	activities := b.mountActivity(server, db, organizations, ui)
	tasks := appaufgabe.NewService(db, projects, agents, activities, db)
	handler := webaufgabe.NewHandler(tasks, projects, appprojektarchiv.NewService(db, organizations), agents, organizations, ui, b.address())
	for _, pattern := range []string{
		"GET /api/organisationen/{id}/projekte/{projektID}/aufgaben",
		"POST /api/organisationen/{id}/projekte/{projektID}/aufgaben",
		"GET /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}",
		"GET /organisationen/{id}/projekte/{projektID}/aufgaben",
		"GET /organisationen/{id}/projekte/{projektID}/aufgaben/neu",
		"POST /organisationen/{id}/projekte/{projektID}/aufgaben/neu",
		"GET /organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}",
	} {
		server.Handle(pattern, handler)
	}

	return tasks
}
