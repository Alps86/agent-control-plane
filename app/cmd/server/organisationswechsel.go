package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webwechsel "agentcontrolplane/app/internal/adapter/web/organisationswechsel"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	appregel "agentcontrolplane/app/internal/app/organisationsregel"
	"agentcontrolplane/ui/bridge"
)

func (b *Bootstrap) mountOrganizationSwitch(server *web.Server, db *sqlite.Database, organizations *apporganisation.Service, ui *bridge.Bridge) {
	rules := appregel.NewService(sqlite.NewOrganizationStore(db), apporganisation.NewLocalIdentity(), db)
	handler := webwechsel.NewHandler(organizations, rules, ui, b.address())
	for _, pattern := range []string{
		"GET /api/organisationswechsel", "POST /api/organisationen/{id}/wechsel",
		"GET /api/organisationen/{id}/arbeitsregeln", "PUT /api/organisationen/{id}/arbeitsregeln",
		"DELETE /api/organisationen/{id}/arbeitsregeln",
		"POST /api/organisationen/{id}/arbeitsregeln/vorschau", "GET /organisationswechsel",
		"POST /organisationswechsel/{id}", "GET /organisationen/{id}/arbeitsregeln",
		"POST /organisationen/{id}/arbeitsregeln", "POST /organisationen/{id}/arbeitsregeln/vorschau",
		"POST /organisationen/{id}/arbeitsregeln/widerruf",
	} {
		server.Handle(pattern, handler)
	}
}
