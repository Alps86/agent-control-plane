package kommentar

import (
	"errors"
	"regexp"

	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	portkommentar "agentcontrolplane/app/internal/port/kommentar"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

const ActionTaskComment = "task.comment"

var ErrAccessDenied = errors.New("access_denied")
var ErrActionDenied = errors.New("action_denied")
var ErrNotFound = portkommentar.ErrNotFound
var ErrContentRequired = domainkommentar.ErrContentRequired
var ErrInvalidReference = domainkommentar.ErrInvalidReference
var internalLinkPattern = regexp.MustCompile(`^/[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*$`)

// CreateInput enthält keine Autoren- oder Zeitangaben des Aufrufers.
type CreateInput struct {
	Content   string                     `json:"content"`
	Reference *domainkommentar.Reference `json:"reference,omitempty"`
}

// Service verwendet ausschließlich injizierte serverseitige Identitäten.
type Service struct {
	store            portkommentar.Store
	operatorIdentity portorganisation.Identity
	agentIdentity    portorganisation.Identity
	artifactResolver portkommentar.ArtifactResolver
}
