package datenbereich

import (
	"context"
	"errors"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	domaindatenbereich "agentcontrolplane/app/internal/domain/datenbereich"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	"agentcontrolplane/app/internal/domain/rechte"
	portdatenbereich "agentcontrolplane/app/internal/port/datenbereich"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

// NewService übernimmt ausschließlich serverseitig injizierte Identitäten.
func NewService(store portdatenbereich.Store, operatorIdentity, agentIdentity portorganisation.Identity) *Service {
	return &Service{store: store, operatorIdentity: operatorIdentity, agentIdentity: agentIdentity}
}

// Replace ersetzt alle Projektfreigaben des Agenten in einer Transaktion.
func (s *Service) Replace(ctx context.Context, organizationID, agentID string, inputs []ScopeInput) ([]domaindatenbereich.Scope, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return nil, err
	}

	scopes, err := s.prepare(organizationID, agentID, inputs)
	if err != nil {
		return nil, err
	}

	if err := s.store.ReplaceScopes(ctx, organizationID, agentID, operatorID, scopes); err != nil {
		return nil, err
	}

	return s.store.ListScopes(ctx, organizationID, agentID, operatorID)
}

// List liest nur Freigaben des zugeordneten Agenten.
func (s *Service) List(ctx context.Context, organizationID, agentID string) ([]domaindatenbereich.Scope, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return nil, err
	}

	return s.store.ListScopes(ctx, organizationID, agentID, operatorID)
}

// Projects zeigt dem Betreiber nur wählbare Projekte dieser Organisation.
func (s *Service) Projects(ctx context.Context, organizationID, agentID string) ([]domainprojekt.Project, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return nil, err
	}

	return s.store.ListProjectsForOperator(ctx, organizationID, agentID, operatorID)
}

// Agent liefert nur das zugeordnete Profil für die Betreiberansicht.
func (s *Service) Agent(ctx context.Context, organizationID, agentID string) (domainagent.Agent, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return domainagent.Agent{}, err
	}

	return s.store.FindAgentForOperator(ctx, organizationID, agentID, operatorID)
}

// ListDenials macht persistierte Verweigerungen ohne fremde Projektdaten sichtbar.
func (s *Service) ListDenials(ctx context.Context, organizationID, agentID string) ([]domaindatenbereich.Denial, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return nil, err
	}

	return s.store.ListDenials(ctx, organizationID, agentID, operatorID)
}

// ReadProject liefert echte Projektdaten nur im aktuellen Agentenumfang.
func (s *Service) ReadProject(ctx context.Context, organizationID, projectID string) (domainprojekt.Project, error) {
	agentID, err := s.agentID()
	if err != nil {
		return domainprojekt.Project{}, err
	}

	project, err := s.store.ReadProjectForAgent(ctx, organizationID, agentID, projectID)
	return s.projectResult(ctx, organizationID, agentID, projectID, domaindatenbereich.ReadProject, project, err)
}

// UpdateProjectDescription ist die einzige Story-19-Schreiboperation des Agenten.
func (s *Service) UpdateProjectDescription(ctx context.Context, organizationID, projectID, description string) (domainprojekt.Project, error) {
	agentID, err := s.agentID()
	if err != nil {
		return domainprojekt.Project{}, err
	}

	project, err := s.store.UpdateProjectDescriptionForAgent(ctx, organizationID, agentID, projectID, description)
	return s.projectResult(ctx, organizationID, agentID, projectID, domaindatenbereich.WriteProjectDetail, project, err)
}

func (s *Service) projectResult(ctx context.Context, organizationID, agentID, projectID string, action domaindatenbereich.Action, project domainprojekt.Project, err error) (domainprojekt.Project, error) {
	if !errors.Is(err, portdatenbereich.ErrNotFound) {
		return project, err
	}

	if auditErr := s.store.AuditDenied(ctx, organizationID, agentID, projectID, action); auditErr != nil {
		return domainprojekt.Project{}, auditErr
	}

	return domainprojekt.Project{}, ErrNotFound
}

func (s *Service) operatorID() (string, error) {
	if s == nil || s.store == nil || s.operatorIdentity == nil {
		return "", ErrAccessDenied
	}

	actors := s.operatorIdentity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}

	return actors[0].ID, nil
}

func (s *Service) agentID() (string, error) {
	if s == nil || s.store == nil || s.agentIdentity == nil {
		return "", ErrAccessDenied
	}

	actors := s.agentIdentity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Agent {
		return "", ErrAccessDenied
	}

	return actors[0].ID, nil
}

func (s *Service) prepare(organizationID, agentID string, inputs []ScopeInput) ([]domaindatenbereich.Scope, error) {
	seen := map[string]bool{}
	scopes := make([]domaindatenbereich.Scope, 0, len(inputs))
	for _, input := range inputs {
		if input.ProjectID == "" || seen[input.ProjectID] {
			return nil, ErrInvalidScope
		}

		seen[input.ProjectID] = true
		scopes = append(scopes, domaindatenbereich.Scope{AgentID: agentID, OrganizationID: organizationID,
			ProjectID: input.ProjectID, CanRead: input.CanRead, CanWrite: input.CanWrite})
	}

	return scopes, nil
}
