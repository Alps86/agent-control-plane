package agent

import (
	"context"
	"errors"

	domainagent "agentcontrolplane/app/internal/domain/agent"
)

var ErrNotFound = errors.New("Agent nicht gefunden")
var ErrNameConflict = errors.New("Agentenname ist bereits vergeben")

// Store erzwingt Organisations- und Betreiberzuordnung in jeder Abfrage.
type Store interface {
	CreateAgent(context.Context, string, domainagent.Agent) error
	ListAgents(context.Context, string, string) ([]domainagent.Agent, error)
	FindAgent(context.Context, string, string, string) (domainagent.Agent, error)
}

// StatusStore ergänzt gezielte Statusänderungen ohne den Agentenbasisvertrag zu erweitern.
type StatusStore interface {
	UpdateAgentStatus(context.Context, string, string, string, domainagent.Status) error
}

// ReadinessProbe leitet Adapterbereitschaft ohne Seiteneffekt ab.
type ReadinessProbe interface {
	Status(context.Context, domainagent.Agent) domainagent.Readiness
}
