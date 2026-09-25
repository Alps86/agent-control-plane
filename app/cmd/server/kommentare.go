package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webkommentar "agentcontrolplane/app/internal/adapter/web/kommentar"
	appaufgabe "agentcontrolplane/app/internal/app/aufgabe"
	appkommentar "agentcontrolplane/app/internal/app/kommentar"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	"agentcontrolplane/ui/bridge"
)

// mountComments bindet den Betreiber-Thread; Artefakte brauchen bis ART-01 einen Resolver.
func (b *Bootstrap) mountComments(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, tasks *appaufgabe.Service, ui *bridge.Bridge) {
	projects := appprojekt.NewService(db, organizations)
	comments := appkommentar.NewService(db, apporganisation.NewLocalIdentity(), nil)
	handler := webkommentar.NewHandler(comments, tasks, projects, organizations, ui, b.address())
	for _, pattern := range []string{
		"GET /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare",
		"POST /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare",
		"PUT /api/organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare/{kommentarID}",
		"GET /organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare",
		"POST /organisationen/{id}/projekte/{projektID}/aufgaben/{aufgabeID}/kommentare",
	} {
		server.Handle(pattern, handler)
	}
}
