package codexcli

import domainagent "agentcontrolplane/app/internal/domain/agent"

// NewTemplate beschreibt Codex CLI ohne vorab freigegebene Fachfähigkeit.
func NewTemplate() domainagent.Template {
	return domainagent.Template{
		ID:           "codex-cli",
		Name:         "Codex CLI",
		Role:         "Arbeitsagent",
		Instructions: "Bearbeite nur zugewiesene Aufgaben innerhalb freigegebener Fachfähigkeiten.",
		Kind:         domainagent.CodexCLI,
		Capabilities: []string{},
	}
}
