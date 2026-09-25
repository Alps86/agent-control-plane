package berichtsweg

import (
	"context"
	"errors"

	domainberichtsweg "agentcontrolplane/app/internal/domain/berichtsweg"
)

var ErrAgentNotFound = errors.New("agent_not_found")
var ErrSelfParent = errors.New("self_parent")
var ErrCycle = errors.New("reporting_cycle")

// Store liest alle Organisationsagenten und ändert Berichtskanten atomar.
// Eine unbekannte oder fremde Elternreferenz liefert ErrAgentNotFound.
type Store interface {
	ListLines(context.Context, string, string) ([]domainberichtsweg.Line, error)
	AssignParent(context.Context, string, string, string, string) error
}
