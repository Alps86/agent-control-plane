package projekt

import (
	"context"

	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	portprojekt "agentcontrolplane/app/internal/port/projekt"
	"github.com/google/uuid"
)

// NewService bindet Projektspeicher und Organisationszugriff an den Anwendungsfall.
func NewService(store portprojekt.ProjectStore, organizations portprojekt.OrganizationReader) *Service {
	return &Service{store: store, organizations: organizations}
}

// Create legt ein Projekt mit geprüftem Zielbezug an.
func (s *Service) Create(ctx context.Context, organizationID, name, description, goalID string) (domainprojekt.Project, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return domainprojekt.Project{}, err
	}

	project, err := domainprojekt.NewProject(uuid.NewString(), organizationID, name, description, goalID)
	if err != nil {
		return domainprojekt.Project{}, err
	}

	if err := s.store.CreateProject(ctx, project); err != nil {
		return domainprojekt.Project{}, err
	}

	return project, nil
}

// List zeigt nur Projekte einer zugänglichen Organisation.
func (s *Service) List(ctx context.Context, organizationID string) ([]domainprojekt.Project, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return nil, err
	}

	return s.store.ListProjects(ctx, organizationID)
}

// Find liest ein Projekt innerhalb seiner zugänglichen Organisation.
func (s *Service) Find(ctx context.Context, organizationID, projectID string) (domainprojekt.Project, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return domainprojekt.Project{}, err
	}

	return s.store.FindProject(ctx, organizationID, projectID)
}

func (s *Service) checkOrganization(ctx context.Context, organizationID string) error {
	if s == nil || s.store == nil || s.organizations == nil {
		return ErrAccessDenied
	}

	_, err := s.organizations.Get(ctx, organizationID)
	return err
}
