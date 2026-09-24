package codexprofil

import (
	"context"
	"errors"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	domain "agentcontrolplane/app/internal/domain/codexprofil"
	"agentcontrolplane/app/internal/domain/rechte"
	portagent "agentcontrolplane/app/internal/port/agent"
	portprofil "agentcontrolplane/app/internal/port/codexprofil"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

// NewService bietet einen injizierbaren Identity-Port für die öffentliche API.
func NewService(store portprofil.Store, agents portagent.Store, identity portorganisation.Identity, organizations portorganisation.Store, workspace portprofil.Workspace) *Service {
	return &Service{store: store, agents: agents, identity: identity, organizations: organizations, workspace: workspace}
}

// Get liefert ein Profil nur nach Betreiber-, Organisations- und Agentenprüfung.
func (s *Service) Get(ctx context.Context, organizationID, agentID string) (domain.Profile, error) {
	operatorID, err := s.authorize(ctx, organizationID, agentID)
	if err != nil {
		return domain.Profile{}, err
	}

	return s.store.GetCodexProfile(ctx, organizationID, agentID, operatorID)
}

// Save persistiert die Freigabe erst nach sicherer Workspace-Materialisierung.
func (s *Service) Save(ctx context.Context, organizationID, agentID string, input domain.Input) (domain.Profile, error) {
	operatorID, err := s.authorize(ctx, organizationID, agentID)
	if err != nil {
		return domain.Profile{}, err
	}

	if err := input.Validate(); err != nil {
		return domain.Profile{}, err
	}

	if err := s.ensure(ctx, organizationID, agentID, input); err != nil {
		return domain.Profile{}, err
	}

	if err := s.store.SaveCodexProfile(ctx, organizationID, agentID, operatorID, input); err != nil {
		return domain.Profile{}, err
	}

	return s.store.GetCodexProfile(ctx, organizationID, agentID, operatorID)
}

func (s *Service) ensure(ctx context.Context, organizationID, agentID string, input domain.Input) error {
	if !input.WorkspaceEnabled {
		return nil
	}

	return s.workspace.Ensure(ctx, organizationID, agentID)
}

func (s *Service) authorize(ctx context.Context, organizationID, agentID string) (string, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return "", err
	}

	if _, err := s.organizations.Get(ctx, organizationID, operatorID); err != nil {
		return "", s.hidden(err)
	}

	agent, err := s.agents.FindAgent(ctx, organizationID, agentID, operatorID)
	if err != nil {
		return "", s.hidden(err)
	}

	if agent.ExecutionKind != domainagent.CodexCLI {
		return "", ErrExecutionKind
	}

	return operatorID, nil
}

func (s *Service) operatorID() (string, error) {
	if s == nil || s.store == nil || s.agents == nil || s.identity == nil || s.organizations == nil || s.workspace == nil {
		return "", ErrAccessDenied
	}

	actors := s.identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}

	return actors[0].ID, nil
}

func (s *Service) hidden(err error) error {
	if errors.Is(err, portorganisation.ErrNotFound) || errors.Is(err, portagent.ErrNotFound) {
		return ErrNotFound
	}

	return err
}
