package ziel

import (
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
	portziel "agentcontrolplane/app/internal/port/ziel"
)

var (
	ErrNameRequired  = domainziel.ErrNameRequired
	ErrAccessDenied  = apporganisation.ErrAccessDenied
	ErrNotFound      = apporganisation.ErrNotFound
	ErrInvalidParent = portziel.ErrInvalidParent
	ErrGoalNotFound  = portziel.ErrGoalNotFound
	ErrInvalidStatus = domainziel.ErrInvalidStatus
)

// Service ist die Anwendungsgrenze für Stammziele.
type Service struct {
	store         portziel.GoalStore
	organizations portziel.OrganizationReader
}
