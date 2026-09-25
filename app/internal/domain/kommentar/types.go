package kommentar

import "errors"

var ErrContentRequired = errors.New("Kommentarinhalt ist erforderlich")
var ErrInvalidReference = errors.New("Kommentarreferenz ist ungültig")

const (
	SourceOperator    = "operator"
	SourceAgent       = "agent"
	ReferenceTask     = "task"
	ReferenceArtifact = "artifact"
)

// Reference ist ein typisierter Link ohne Datei- oder Downloadversprechen.
type Reference struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	DisplayName    string `json:"display_name,omitempty"`
	Link           string `json:"link,omitempty"`
}

// Comment behält seine serverseitig bestimmte Herkunft und Erstellzeit.
type Comment struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	TaskID         string     `json:"task_id"`
	Content        string     `json:"content"`
	SourceKind     string     `json:"source_kind"`
	SourceID       string     `json:"source_id"`
	SourceName     string     `json:"source_name"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
	Reference      *Reference `json:"reference,omitempty"`
}
