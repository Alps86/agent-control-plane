package credentials

import (
	"context"
	"errors"
)

// ErrNotFound bezeichnet eine noch nicht gespeicherte Verbindung.
var ErrNotFound = errors.New("credential not found")

// Store verwahrt Anmeldegeheimnisse ausschließlich auf dem Server.
type Store interface {
	Save(context.Context, string, []byte) error
	Load(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}

// AccessResolver gibt dem ausgewählten serverseitigen Provider einen
// aktuell nutzbaren Access-Token samt Account-ID derselben Sitzung.
// Die Implementierung verantwortet Erneuerung und atomare Auflösung.
type AccessResolver interface {
	Access(context.Context) (accessToken string, accountID string, err error)
}
