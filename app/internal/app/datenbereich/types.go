package datenbereich

import (
	"errors"

	portdatenbereich "agentcontrolplane/app/internal/port/datenbereich"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var ErrAccessDenied = errors.New("access_denied")
var ErrNotFound = portdatenbereich.ErrNotFound
var ErrInvalidScope = errors.New("invalid_scope")

// ScopeInput enthält nur vom Betreiber wählbare Projektfreigaben.
type ScopeInput struct {
	ProjectID string `json:"project_id"`
	CanRead   bool   `json:"can_read"`
	CanWrite  bool   `json:"can_write"`
}

// Service bindet Betreibervergabe und Agentendaten an getrennte Identitäten.
type Service struct {
	store            portdatenbereich.Store
	operatorIdentity portorganisation.Identity
	agentIdentity    portorganisation.Identity
}
