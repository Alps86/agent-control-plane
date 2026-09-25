package projektarchiv

import (
	"context"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainprojektarchiv "agentcontrolplane/app/internal/domain/projektarchiv"
	portprojekt "agentcontrolplane/app/internal/port/projekt"
	portprojektarchiv "agentcontrolplane/app/internal/port/projektarchiv"
)

// NewService bindet den Archivspeicher und die bestehende Organisationsgrenze.
func NewService(store portprojektarchiv.Store, organizations portprojekt.OrganizationReader) *Service {
	return &Service{store: store, organizations: organizations}
}

// Archive archiviert ein Projekt nur innerhalb seiner Organisation.
func (s *Service) Archive(ctx context.Context, organizationID, projectID string) error {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return err
	}

	return s.store.ArchiveProject(ctx, organizationID, projectID)
}

// Restore stellt ein Projekt mit seiner gesamten Historie wieder her.
func (s *Service) Restore(ctx context.Context, organizationID, projectID string) error {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return err
	}

	return s.store.RestoreProject(ctx, organizationID, projectID)
}

// ListArchived liest ausschließlich das Archiv der gewählten Organisation.
func (s *Service) ListArchived(ctx context.Context, organizationID string) ([]domainprojektarchiv.ArchivedProject, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return nil, err
	}

	return s.store.ListArchivedProjects(ctx, organizationID)
}

// Status liefert den gespeicherten Status auch für direkte Projekt-URLs.
func (s *Service) Status(ctx context.Context, organizationID, projectID string) (domainprojektarchiv.Status, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return "", err
	}

	return s.store.ProjectStatus(ctx, organizationID, projectID)
}

func (s *Service) checkOrganization(ctx context.Context, organizationID string) error {
	if s == nil || s.store == nil || s.organizations == nil {
		return apporganisation.ErrAccessDenied
	}

	_, err := s.organizations.Get(ctx, organizationID)
	return err
}
