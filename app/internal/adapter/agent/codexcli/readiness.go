package codexcli

import (
	"context"
	"strings"

	domainagent "agentcontrolplane/app/internal/domain/agent"
)

// NewProbe bindet eine gespeicherte CLI-Referenz ohne deren Benutzung ein.
func NewProbe(cliReference string) *Probe {
	return &Probe{cliReference: cliReference}
}

// Status verweigert den Start, bis Story 21 die gesamte Fähigkeitsgrenze belegt.
func (p *Probe) Status(_ context.Context, _ domainagent.Agent) domainagent.Readiness {
	if strings.TrimSpace(p.cliReference) == "" {
		return domainagent.Readiness{
			Ready:  false,
			Code:   "codex_cli_reference_missing",
			Reason: "Die Codex-CLI-Konfiguration fehlt.",
		}
	}

	return domainagent.Readiness{
		Ready:  false,
		Code:   "codex_cli_enforcement_unverified",
		Reason: "Die Sperre direkter Shell-, Terminal-, Prozess-, Interpreter-, HTTP- und Dateisystemzugriffe ist für dieses Codex-CLI-Profil nicht nachgewiesen.",
	}
}
