package ziel

import (
	"context"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
)

// GoalStore speichert und liest Ziele innerhalb einer Organisation.
type GoalStore interface {
	CreateGoal(context.Context, domainziel.Goal) error
	ListGoals(context.Context, string) ([]domainziel.Goal, error)
}

// OrganizationReader prüft Zugehörigkeit und Betreiberidentität serverseitig.
type OrganizationReader interface {
	Get(context.Context, string) (domainorganisation.Organization, error)
}
