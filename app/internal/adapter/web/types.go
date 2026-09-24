package web

import (
	"context"
	"net/http"

	"agentcontrolplane/app/internal/port/system"
)

// Server stellt öffentliche HTTP-Routen bereit.
type Server struct {
	probe system.Probe
	mux   *http.ServeMux
	store SchemaVersioner
}

// SchemaVersioner liefert die tatsächlich gespeicherte Migrationsversion.
type SchemaVersioner interface {
	SchemaVersion(context.Context) (int, error)
}
