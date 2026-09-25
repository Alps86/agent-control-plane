package modellfreigabe

import (
	"context"
	"errors"

	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"agentcontrolplane/app/internal/port/modellschluessel"
)

var ErrOrganizationGrantMissing = errors.New("Organisationsfreigabe fehlt")

// Store hält ausschließlich explizite Organisations- und Agentenfreigaben.
type Store interface {
	GrantOrganization(context.Context, string, string, modellschluessel.Binding) error
	RevokeOrganization(context.Context, string, string) error
	OrganizationGranted(context.Context, string, string, modellschluessel.Binding) (bool, error)
	GrantAgent(context.Context, string, string, string, modellschluessel.Binding) error
	RevokeAgent(context.Context, string, string, string) error
	AgentGranted(context.Context, string, string, string, modellschluessel.Binding) (bool, error)
}

// ConnectionStatus ist die redigierte Sicht auf den zentralen Zugang.
type ConnectionStatus interface {
	Status(context.Context) (openrouterverbindung.View, error)
	modellschluessel.BindingReader
	modellschluessel.Resolver
}

// Access löst den Schlüssel erst unmittelbar vor einem kontrollierten Transport.
type Access interface {
	Resolve(context.Context) (string, error)
}

// Caller führt einen kontrollierten Aufruf mit erneut prüfbarem Access aus.
type Caller interface {
	Call(context.Context, string, string, Access) error
}
