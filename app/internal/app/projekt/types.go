package projekt

import (
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	portprojekt "agentcontrolplane/app/internal/port/projekt"
)

var (
	ErrNameRequired = domainprojekt.ErrNameRequired
	ErrGoalRequired = domainprojekt.ErrGoalRequired
	ErrInvalidGoal  = portprojekt.ErrInvalidGoal
	ErrAccessDenied = apporganisation.ErrAccessDenied
	ErrNotFound     = apporganisation.ErrNotFound
)

// Service bildet die Anwendungsgrenze für Projekte.
type Service struct {
	store         portprojekt.ProjectStore
	organizations portprojekt.OrganizationReader
}
