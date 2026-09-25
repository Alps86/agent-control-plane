package projektarchiv

import (
	"context"
	"errors"

	domainprojektarchiv "agentcontrolplane/app/internal/domain/projektarchiv"
)

var ErrNotFound = errors.New("Projekt nicht gefunden")
var ErrAlreadyArchived = errors.New("Projekt ist bereits archiviert")
var ErrNotArchived = errors.New("Projekt ist nicht archiviert")
var ErrActiveRun = errors.New("Projekt hat einen aktiven Lauf")

// Store ändert den Archivzustand und liest die Archivansicht organisationsgebunden.
type Store interface {
	ArchiveProject(context.Context, string, string) error
	RestoreProject(context.Context, string, string) error
	ListArchivedProjects(context.Context, string) ([]domainprojektarchiv.ArchivedProject, error)
	ProjectStatus(context.Context, string, string) (domainprojektarchiv.Status, error)
}
