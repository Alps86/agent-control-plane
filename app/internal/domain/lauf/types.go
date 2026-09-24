package lauf

import "errors"

// Status bezeichnet einen belegten Zustand der Laufvorbereitung.
type Status string

const (
	Reserviert     Status = "Reserviert"
	Fehlgeschlagen Status = "Fehlgeschlagen"
)

var (
	ErrReasonRequired    = errors.New("Vorbereitungsfehler benötigt einen Grund")
	ErrInvalidTransition = errors.New("Laufzustand erlaubt diesen Übergang nicht")
)

// Run ist die persistente öffentliche Sicht auf eine Laufreservierung.
type Run struct {
	ID             string
	OrganizationID string
	TaskID         string
	Status         Status
	FailureReason  string
}
