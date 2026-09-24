package agent

import (
	"errors"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	portagent "agentcontrolplane/app/internal/port/agent"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var ErrAccessDenied = errors.New("access_denied")
var ErrNotFound = portagent.ErrNotFound
var ErrNameRequired = domainagent.ErrNameRequired
var ErrNameConflict = portagent.ErrNameConflict
var ErrTemplateNotFound = errors.New("Vorlage nicht gefunden")
var ErrExecutionKind = errors.New("Ausführungsart passt nicht zur Vorlage")
var ErrCapabilityDenied = domainagent.ErrCapabilityDenied
var ErrInvalidStatus = domainagent.ErrInvalidStatus

// CreateInput enthält nur vom Bediener wählbare Profilfelder.
type CreateInput struct {
	Name                   string                    `json:"name"`
	TemplateID             string                    `json:"template_id"`
	ExecutionKind          domainagent.ExecutionKind `json:"execution_kind"`
	AdditionalCapabilities []string                  `json:"additional_capabilities"`
}

// Profile verbindet gespeicherte Konfiguration mit der aktuellen Bereitschaft.
type Profile struct {
	domainagent.Agent
	Readiness domainagent.Readiness `json:"readiness"`
}

// Service ist die organisationsgebundene Agentenanwendungsgrenze.
type Service struct {
	store         portagent.Store
	identity      portorganisation.Identity
	organizations portorganisation.Store
	catalog       *domainagent.Catalog
	probes        map[domainagent.ExecutionKind]portagent.ReadinessProbe
}

type unconfiguredProbe struct {
	code   string
	reason string
}
