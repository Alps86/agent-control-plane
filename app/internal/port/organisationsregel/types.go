package organisationsregel

import (
	"context"
	"errors"

	"agentcontrolplane/app/internal/domain/organisationsregel"
)

var ErrNotFound = errors.New("organization_not_found")
var ErrRevisionConflict = errors.New("revision_conflict")

// Store hält Regeländerungen organisationsgebunden und revisionssicher.
type Store interface {
	Read(context.Context, string, string) (organisationsregel.Policy, error)
	SaveRule(context.Context, string, string, int64, string, organisationsregel.Rule) (organisationsregel.Policy, error)
}
