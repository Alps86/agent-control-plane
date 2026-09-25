package modellwahl

import (
	"context"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	"agentcontrolplane/app/internal/domain/rechte"
)

func (s *Service) agent(ctx context.Context, organizationID, agentID string) (domainagent.Agent, string, error) {
	operatorID, err := s.operatorID()
	if err != nil {
		return domainagent.Agent{}, "", err
	}

	agent, err := s.agents.FindAgent(ctx, organizationID, agentID, operatorID)
	if err != nil {
		return domainagent.Agent{}, "", err
	}

	if agent.ExecutionKind != domainagent.Eino {
		return domainagent.Agent{}, "", ErrExecutionKind
	}

	return agent, operatorID, nil
}

func (s *Service) operatorID() (string, error) {
	if s == nil || s.agents == nil || s.identity == nil || s.store == nil {
		return "", ErrAccessDenied
	}

	actors := s.identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}

	return actors[0].ID, nil
}
