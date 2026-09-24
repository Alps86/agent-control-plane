package codexprofil

import (
	"context"

	domain "agentcontrolplane/app/internal/domain/codexprofil"
)

const ActionRuntimeProbe = "runtime.probe"
const ActionMarkdownSave = "artifact.markdown.save"

// Authorize prüft die aktuellen Rechte vor jedem tatsächlichen CLI-Probe-Aufruf.
func (s *Service) Authorize(ctx context.Context, organizationID, agentID, actionID string) error {
	if actionID != ActionRuntimeProbe && actionID != ActionMarkdownSave {
		return ErrActionDenied
	}

	profile, err := s.Get(ctx, organizationID, agentID)
	if err != nil {
		return err
	}

	if err := s.actionAllowed(profile, actionID); err != nil {
		return err
	}

	return s.workspace.Verify(ctx, organizationID, agentID)
}

func (s *Service) actionAllowed(profile domain.Profile, actionID string) error {
	if !profile.WorkspaceEnabled {
		return ErrWorkspaceDisabled
	}

	if actionID == ActionMarkdownSave && !profile.WriteEnabled {
		return ErrWriteDisabled
	}

	return nil
}

// SaveProof vermittelt ausschließlich die feste Markdown-Artefaktaktion.
func (s *Service) SaveProof(ctx context.Context, organizationID, agentID, markdown string) error {
	if err := s.Authorize(ctx, organizationID, agentID, ActionMarkdownSave); err != nil {
		return err
	}

	return s.workspace.WriteProof(ctx, organizationID, agentID, markdown)
}
