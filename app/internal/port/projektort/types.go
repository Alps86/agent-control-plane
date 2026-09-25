package projektort

import (
	"context"
	"errors"

	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	domainort "agentcontrolplane/app/internal/domain/projektort"
)

var ErrMissing = errors.New("Projektort nicht gebunden")
var ErrNotFound = errors.New("Projekt nicht gefunden")
var ErrIntegrity = errors.New("Privater Projektort verletzt die Pfadintegrität")

// ProjectReader prüft die Betreiber- und Organisationsgrenze.
type ProjectReader interface {
	Find(context.Context, string, string) (domainprojekt.Project, error)
}

// Store speichert nur die logische Bindung innerhalb des Projekts.
type Store interface {
	Get(context.Context, string, string) (domainort.Binding, error)
	Save(context.Context, domainort.Binding) error
}

// Locator materialisiert und überprüft den ausschließlich serverseitigen Pfad.
type Locator interface {
	Ensure(context.Context, string, string) (string, error)
	Resolve(context.Context, string, string) (string, error)
}

// StartProbe erhält nur einen zuvor vom Locator überprüften Pfad.
type StartProbe interface {
	Probe(context.Context, string, string, string) error
}
