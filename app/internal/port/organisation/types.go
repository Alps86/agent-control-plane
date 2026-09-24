package organisation

import (
	"context"
	"errors"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
)

// ErrNameConflict bezeichnet einen bereits belegten Organisationsnamen.
var ErrNameConflict = errors.New("Name ist bereits vergeben")

// ErrNotFound hält unbekannte und nicht zugeordnete Referenzen gleich.
var ErrNotFound = errors.New("Organisation nicht gefunden")

// Identity liefert ausschließlich die serverseitig bestimmte Akteursidentität.
type Identity interface {
	Actors() []rechte.Actor
}

// Store liest und schreibt nur die freigegebenen Organisationsdaten.
type Store interface {
	Create(context.Context, string, domainorganisation.Organization) error
	List(context.Context, string) ([]domainorganisation.Organization, error)
	Get(context.Context, string, string) (domainorganisation.Organization, error)
}
