package modellfreigabe

import (
	"errors"

	"agentcontrolplane/app/internal/domain/modellfreigabe"
	portagent "agentcontrolplane/app/internal/port/agent"
	portmodell "agentcontrolplane/app/internal/port/modellfreigabe"
	"agentcontrolplane/app/internal/port/modellschluessel"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var ErrAccessDenied = errors.New("access_denied")
var ErrNotFound = errors.New("Organisation oder Eino-Agent nicht gefunden")
var ErrConnection = errors.New("OpenRouter-Verbindung fehlt")
var ErrOrganizationGrantMissing = portmodell.ErrOrganizationGrantMissing

// AgentStatus ergänzt die fachliche Entscheidung nur um öffentliche Identität.
type AgentStatus struct {
	ID     string                `json:"id"`
	Name   string                `json:"name"`
	Status modellfreigabe.Status `json:"status"`
}

// Overview ist die vollständige autorisierte Settings-Ansicht einer Organisation.
type Overview struct {
	OrganizationID   string                `json:"organization_id"`
	OrganizationName string                `json:"organization_name"`
	Status           modellfreigabe.Status `json:"status"`
	Agents           []AgentStatus         `json:"agents"`
}

// Service koordiniert Freigaben, Referenzstatus und beide Entscheidungsgrenzen.
type Service struct {
	store         portmodell.Store
	organizations portorganisation.Store
	agents        portagent.Store
	identity      portorganisation.Identity
	connection    portmodell.ConnectionStatus
}

// access ist die nicht serialisierbare Berechtigung für genau einen Aufruf.
type access struct {
	service        *Service
	organizationID string
	agentID        string
	binding        modellschluessel.Binding
}
