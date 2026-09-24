package lauf

import (
	"errors"

	apprechte "agentcontrolplane/app/internal/app/rechte"
	portlauf "agentcontrolplane/app/internal/port/lauf"
)

var (
	ErrUnavailable             = errors.New("Aufgabe oder Lauf nicht verfügbar")
	ErrAdapterEvidenceRequired = errors.New("Erfolg ohne Adapterbeleg nicht zulässig")
)

// Service ist die öffentliche Anwendungsgrenze für Laufreservierungen.
type Service struct {
	store  portlauf.Store
	reader *apprechte.Reader
}
