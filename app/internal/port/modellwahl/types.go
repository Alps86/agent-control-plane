package modellwahl

import (
	"context"
	"errors"

	domain "agentcontrolplane/app/internal/domain/modellwahl"
)

var ErrNotFound = errors.New("model selection not found")
var ErrOrganizationGrant = errors.New("Organisationsfreigabe fehlt")
var ErrAgentGrant = errors.New("Agentenfreigabe fehlt")
var ErrConnectionUnavailable = errors.New("Verbindung nicht einsatzbereit")

// Store persists one selected route per agent without credentials.
type Store interface {
	Get(context.Context, string, string, string) (domain.Selection, error)
	Replace(context.Context, string, string, string, domain.Selection) error
}

// OpenRouterAccess checks both grants and the current connection state.
type OpenRouterAccess interface {
	Allowed(context.Context, string, string, string) error
}
