package lauf

import (
	"context"

	domainlauf "agentcontrolplane/app/internal/domain/lauf"
)

// Store hält Reservierung und Zustandswechsel über Neustarts hinweg atomar.
type Store interface {
	Reserve(ctx context.Context, organizationID, taskID string) (domainlauf.Run, bool, error)
	Lookup(ctx context.Context, organizationID, taskID, runID string) (domainlauf.Run, bool, error)
	List(ctx context.Context, organizationID, taskID string) ([]domainlauf.Run, error)
	FailPreparation(ctx context.Context, organizationID, taskID, runID, reason string) (domainlauf.Run, error)
}
