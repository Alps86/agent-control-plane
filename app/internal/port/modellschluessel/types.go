package modellschluessel

import (
	"context"
	"errors"
)

var ErrMissing = errors.New("model connection missing")
var ErrNotReady = errors.New("model connection not ready")
var ErrRevoked = errors.New("model connection revoked")

// Binding identifiziert genau eine Einrichtung einer Modellverbindung.
// Eine Rotation behält die Generation; eine neue Einrichtung erhält eine neue.
type Binding struct {
	Reference  string
	Generation string
}

// Resolver liefert einen Schlüssel ausschließlich an vertrauenswürdige Go-Provider.
// Der Provider ruft Resolve unmittelbar vor jedem neuen Modellaufruf auf.
type Resolver interface {
	Resolve(context.Context, Binding) (string, error)
}

// BindingReader liefert die aktuell eingerichtete Verbindungsinstanz.
type BindingReader interface {
	Binding(context.Context) (Binding, error)
}
