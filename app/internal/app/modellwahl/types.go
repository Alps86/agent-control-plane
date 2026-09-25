package modellwahl

import (
	"errors"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	domain "agentcontrolplane/app/internal/domain/modellwahl"
	portagent "agentcontrolplane/app/internal/port/agent"
	portmodellwahl "agentcontrolplane/app/internal/port/modellwahl"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var ErrAccessDenied = errors.New("access_denied")
var ErrNotFound = portagent.ErrNotFound
var ErrCatalog = errors.New("Modellroute ist nicht im registrierten Katalog")
var ErrOpenRouterCatalogConnection = errors.New("OpenRouter-Katalog: explizite Verbindungsreferenz openrouter-central fehlt oder ist ungültig")
var ErrExecutionKind = errors.New("Modellwahl erfordert einen Eino-Agenten")
var ErrGrant = errors.New("Modellverbindung nicht freigegeben")

// View is the public, credential-free selection and catalogue response.
type View struct {
	Agent         domainagent.Agent `json:"agent"`
	ExecutionKind string            `json:"execution_kind"`
	Selection     domain.Selection  `json:"selection"`
	Providers     []domain.Provider `json:"providers"`
}

// Service owns the model choice while reusing the existing agent boundary.
type Service struct {
	agents   portagent.Store
	identity portorganisation.Identity
	store    portmodellwahl.Store
	access   portmodellwahl.OpenRouterAccess
	catalog  domain.Catalog
}

// CatalogLoader reads and normalizes non-secret configuration.
type CatalogLoader struct{}
