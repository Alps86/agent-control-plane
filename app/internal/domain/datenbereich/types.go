package datenbereich

// Scope bindet zwei unabhängige Datenrechte an ein Projekt eines Agenten.
type Scope struct {
	AgentID        string `json:"agent_id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	CanRead        bool   `json:"can_read"`
	CanWrite       bool   `json:"can_write"`
}

// Denial enthält nur für die Organisation sichtbare Auditdaten.
type Denial struct {
	ID             int64  `json:"id"`
	AgentID        string `json:"agent_id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id,omitempty"`
	Action         Action `json:"action"`
	OccurredAt     string `json:"occurred_at"`
}

// Action benennt die konkrete, auditierte Datenoperation.
type Action string

const (
	ReadProject        Action = "read_project"
	WriteProjectDetail Action = "write_project_description"
)
