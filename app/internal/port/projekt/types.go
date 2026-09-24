package projekt

import (
	"context"
	"errors"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
)

// ErrInvalidGoal bezeichnet ein unbekanntes oder organisationsfremdes Ziel.
var ErrInvalidGoal = errors.New("Ziel ist ungültig")

// ProjectStore speichert und liest Projekte nur innerhalb einer Organisation.
type ProjectStore interface {
	CreateProject(context.Context, domainprojekt.Project) error
	ListProjects(context.Context, string) ([]domainprojekt.Project, error)
	FindProject(context.Context, string, string) (domainprojekt.Project, error)
}

// OrganizationReader prüft die serverseitige Betreiberzuordnung.
type OrganizationReader interface {
	Get(context.Context, string) (domainorganisation.Organization, error)
}
