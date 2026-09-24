package main

import (
	"os"

	"agentcontrolplane/app/internal/adapter/agent/codexcli"
	appagent "agentcontrolplane/app/internal/app/agent"
	domainagent "agentcontrolplane/app/internal/domain/agent"
)

// registerCodexAgent ergänzt die getrennte Vorlage und den gesperrten CLI-Status.
func (b *Bootstrap) registerCodexAgent(service *appagent.Service) {
	service.RegisterTemplate(codexcli.NewTemplate())
	service.RegisterProbe(domainagent.CodexCLI, codexcli.NewProbe(os.Getenv("APP_CODEX_CLI_REFERENCE")))
}
