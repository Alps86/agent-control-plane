package ziel

import (
	"context"
	"errors"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainziel "agentcontrolplane/app/internal/domain/ziel"
)

var (
	ErrInvalidParent = errors.New("Übergeordnetes Ziel ist ungültig")
	ErrGoalNotFound  = errors.New("Ziel nicht gefunden")
)

// GoalStore speichert und liest Ziele innerhalb einer Organisation.
type GoalStore interface {
	CreateGoal(context.Context, domainziel.Goal) error
	CreateChildGoal(context.Context, domainziel.Goal) error
	ListGoals(context.Context, string) ([]domainziel.Goal, error)
	SetGoalStatus(context.Context, string, string, string) (domainziel.Goal, error)
	ListGoalProjectLinks(context.Context, string) ([]domainziel.ProjectLink, error)
}

// OrganizationReader prüft Zugehörigkeit und Betreiberidentität serverseitig.
type OrganizationReader interface {
	Get(context.Context, string) (domainorganisation.Organization, error)
}
