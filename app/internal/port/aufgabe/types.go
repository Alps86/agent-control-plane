package aufgabe

import (
	"context"
	"errors"

	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
)

var ErrNotFound = errors.New("Aufgabe nicht gefunden")
var ErrInvalidReference = errors.New("Projekt oder Agent ist ungültig")

// Store speichert Aufgaben ausschließlich in ihrem Organisationsbereich.
type Store interface {
	CreateTask(context.Context, domainaufgabe.Task) error
	ListTasks(context.Context, string, string) ([]domainaufgabe.Task, error)
	FindTask(context.Context, string, string) (domainaufgabe.Task, error)
}
