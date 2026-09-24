package codexprofil

import (
	"context"

	domain "agentcontrolplane/app/internal/domain/codexprofil"
)

// Store bindet jedes Profil an eine bereits überprüfte Organisation und Betreiberin.
type Store interface {
	GetCodexProfile(context.Context, string, string, string) (domain.Profile, error)
	SaveCodexProfile(context.Context, string, string, string, domain.Input) error
}

// Workspace materialisiert ausschließlich den serverseitig abgeleiteten privaten Ort.
type Workspace interface {
	Ensure(context.Context, string, string) error
	Verify(context.Context, string, string) error
	WriteProof(context.Context, string, string, string) error
}

// ActionRequest ist das einzige Wire-Format zwischen isolierter MCP-Aktion und Host-Broker.
type ActionRequest struct {
	ActionID string `json:"action_id"`
	Markdown string `json:"markdown"`
}

// ActionResponse meldet nur den Aktionsentscheid, keine Hostpfade oder Identitäten.
type ActionResponse struct {
	Allowed bool   `json:"allowed"`
	Code    string `json:"code"`
}

// Session hält nur den privaten Broker-Socket für einen Runtime-Prüfaufruf.
type Session interface {
	SocketPath() string
	Close() error
}

// BrokerFactory bindet einen frischen Socket an den vertrauenswürdigen Aufrufkontext.
type BrokerFactory interface {
	Start(context.Context, string, string) (Session, error)
}
