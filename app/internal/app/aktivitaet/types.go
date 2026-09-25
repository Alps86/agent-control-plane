package aktivitaet

import (
	"errors"

	apporganisation "agentcontrolplane/app/internal/app/organisation"
	portaktivitaet "agentcontrolplane/app/internal/port/aktivitaet"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var ErrAccessDenied = apporganisation.ErrAccessDenied
var ErrNotFound = apporganisation.ErrNotFound
var ErrInvalidSource = errors.New("ungültige Ereignisquelle")

// Service liest Ereignisse und zeichnet den ersten Aufgabenübergang auf.
type Service struct {
	store         portaktivitaet.Store
	organizations *apporganisation.Service
	identity      portorganisation.Identity
}
