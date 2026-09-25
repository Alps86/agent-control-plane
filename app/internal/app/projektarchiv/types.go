package projektarchiv

import (
	portprojekt "agentcontrolplane/app/internal/port/projekt"
	portprojektarchiv "agentcontrolplane/app/internal/port/projektarchiv"
)

var ErrNotFound = portprojektarchiv.ErrNotFound
var ErrAlreadyArchived = portprojektarchiv.ErrAlreadyArchived
var ErrNotArchived = portprojektarchiv.ErrNotArchived
var ErrActiveRun = portprojektarchiv.ErrActiveRun

// Service setzt Organisationszugriff vor jede Archivoperation.
type Service struct {
	store         portprojektarchiv.Store
	organizations portprojekt.OrganizationReader
}
