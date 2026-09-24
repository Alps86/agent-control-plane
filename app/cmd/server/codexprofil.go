package main

import (
	"os"

	runtimecli "agentcontrolplane/app/internal/adapter/runtime/codexcli"
	"agentcontrolplane/app/internal/adapter/runtime/codexprofil"
	"agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/adapter/web"
	webprofil "agentcontrolplane/app/internal/adapter/web/codexprofil"
	appagent "agentcontrolplane/app/internal/app/agent"
	appprofil "agentcontrolplane/app/internal/app/codexprofil"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/ui/bridge"
)

// mountCodexProfile bindet das serverseitige CLI-Arbeitsprofil ein.
func (b *Bootstrap) mountCodexProfile(server *web.Server, db *sqlite.Database, ui *bridge.Bridge) {
	organizations := sqlite.NewOrganizationStore(db)
	identity := apporganisation.NewLocalIdentity()
	agents := appagent.NewService(db, identity, organizations)
	b.registerCodexAgent(agents)
	profiles := appprofil.NewService(db, db, identity, organizations, codexprofil.NewLocator(b.path()))
	runner := runtimecli.NewRunner(b.codexRuntimeConfig(), profiles, codexprofil.NewBrokerFactory(profiles))
	handler := webprofil.NewHandler(profiles, agents, ui, b.address(), runner)
	for _, pattern := range []string{
		"/api/organisationen/{id}/agenten/{agentID}/codex-profil",
		"/api/organisationen/{id}/agenten/{agentID}/codex-profil/pruefen",
		"/organisationen/{id}/agenten/{agentID}/codex-profil",
	} {
		server.Handle(pattern, handler)
	}
}

func (b *Bootstrap) codexRuntimeConfig() runtimecli.Config {
	return runtimecli.Config{
		DatabasePath: b.path(), CLIPath: os.Getenv("APP_CODEX_CLI_PATH"),
		CLISHA256:          os.Getenv("APP_CODEX_CLI_SHA256"),
		BwrapPath:          os.Getenv("APP_CODEX_BWRAP_PATH"),
		BwrapSHA256:        os.Getenv("APP_CODEX_BWRAP_SHA256"),
		ActionServerPath:   os.Getenv("APP_CODEX_ACTION_SERVER_PATH"),
		ActionServerSHA256: os.Getenv("APP_CODEX_ACTION_SERVER_SHA256"),
		Scenario:           os.Getenv("APP_CODEX_RUNTIME_SCENARIO"),
	}
}
