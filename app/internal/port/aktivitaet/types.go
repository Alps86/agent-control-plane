package aktivitaet

import (
	"context"
	"errors"

	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
)

var ErrInvalidEvent = errors.New("Aktivitätsereignis ist ungültig")
var ErrInvalidReference = errors.New("Aufgabenbezug ist ungültig")

// Store persistiert und liest organisationsgebundene Fachereignisse.
type Store interface {
	Record(context.Context, domainaktivitaet.Event) error
	ListEvents(context.Context, string) ([]domainaktivitaet.Event, error)
}

// Recorder wird aus dem atomaren Aufgabenanwendungsfall aufgerufen.
type Recorder interface {
	RecordTaskCreation(context.Context, domainaufgabe.Task, string, string) error
}

// Transactor führt Aufgabe und Ereignisse mit demselben Kontext atomar aus.
type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}
