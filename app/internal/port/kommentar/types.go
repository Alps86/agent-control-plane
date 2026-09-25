package kommentar

import (
	"context"
	"errors"

	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	"agentcontrolplane/app/internal/domain/rechte"
)

var ErrNotFound = errors.New("not_found")

// ArtifactMetadata enthält nur einen überprüften, app-internen Verweis.
type ArtifactMetadata struct {
	ID             string
	OrganizationID string
	TaskID         string
	DisplayName    string
	Link           string
}

// ArtifactResolver prüft vorhandene Artefaktmetadaten ohne Inhalt oder Dateiabruf.
type ArtifactResolver interface {
	ResolveArtifact(context.Context, string, string, string) (ArtifactMetadata, error)
}

// Store erzwingt Task-, Organisations- und Agentenbereich bei jeder Operation.
type Store interface {
	CreateComment(context.Context, rechte.Actor, domainkommentar.Comment) (domainkommentar.Comment, error)
	ListComments(context.Context, rechte.Actor, string, string) ([]domainkommentar.Comment, error)
	UpdateComment(context.Context, rechte.Actor, string, string, string, string) (domainkommentar.Comment, error)
	AuditDeniedComment(context.Context, rechte.Actor, string, string, string) error
}
