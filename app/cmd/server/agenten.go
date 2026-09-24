package main

import (
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	appagent "agentcontrolplane/app/internal/app/agent"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
)

// mountAgents bindet die organisationsgebundenen Agentenrouten ein.
func (b *Bootstrap) mountAgents(server *web.Server, db *sqlite.Database, ui *bridge.Bridge) {
	service := appagent.NewService(db, apporganisation.NewLocalIdentity(), sqlite.NewOrganizationStore(db))
	agents := webagent.NewHandler(service, ui, b.address())
	for _, pattern := range []string{
		"/api/organisationen/{id}/agenten",
		"/api/organisationen/{id}/agenten/",
		"/organisationen/{id}/agenten",
		"/organisationen/{id}/agenten/",
	} {
		server.Handle(pattern, agents)
	}
}
