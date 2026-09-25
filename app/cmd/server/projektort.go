package main

import (
	runtimeprojektort "agentcontrolplane/app/internal/adapter/runtime/projektort"
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webprojektort "agentcontrolplane/app/internal/adapter/web/projektort"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	appprojektort "agentcontrolplane/app/internal/app/projektort"
	"agentcontrolplane/ui/bridge"
)

func (b *Bootstrap) mountProjectLocations(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, ui *bridge.Bridge) {
	projects := appprojekt.NewService(db, organizations)
	locator := runtimeprojektort.NewLocator(b.path())
	locations := appprojektort.NewService(db, projects, locator, runtimeprojektort.NewStartProbe(locator))
	handler := webprojektort.NewHandler(locations, projects, organizations, ui, b.address())
	for _, pattern := range []string{
		"/api/organisationen/{id}/projekte/{projektID}/ausfuehrungsort",
		"/api/organisationen/{id}/projekte/{projektID}/ausfuehrungsort/startpruefung",
		"/organisationen/{id}/projekte/{projektID}/ausfuehrungsort",
		"/organisationen/{id}/projekte/{projektID}/ausfuehrungsort/startpruefung",
	} {
		server.Handle(pattern, handler)
	}
}
