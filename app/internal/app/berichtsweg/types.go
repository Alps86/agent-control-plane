package berichtsweg

import (
	"errors"

	domainberichtsweg "agentcontrolplane/app/internal/domain/berichtsweg"
	portagent "agentcontrolplane/app/internal/port/agent"
	portberichtsweg "agentcontrolplane/app/internal/port/berichtsweg"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var ErrAccessDenied = errors.New("access_denied")
var ErrNotFound = portorganisation.ErrNotFound
var ErrAgentNotFound = portberichtsweg.ErrAgentNotFound
var ErrSelfParent = portberichtsweg.ErrSelfParent
var ErrCycle = portberichtsweg.ErrCycle

type Chart = domainberichtsweg.Chart
type Line = domainberichtsweg.Line

// Service ist die betreibergebundene öffentliche Anwendungsgrenze.
type Service struct {
	store         portberichtsweg.Store
	agents        portagent.Store
	organizations portorganisation.Store
	identity      portorganisation.Identity
}
