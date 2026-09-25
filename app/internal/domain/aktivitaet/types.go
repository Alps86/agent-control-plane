package aktivitaet

import "time"

const (
	KindCreated   = "created"
	KindAssigned  = "assigned"
	SourceAPI     = "api"
	SourceBrowser = "browser"
)

// Event hält ein fachliches Aufgabenereignis mit seiner Herkunft fest.
type Event struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ProjectID      string    `json:"project_id"`
	TaskID         string    `json:"task_id"`
	Kind           string    `json:"kind"`
	Actor          string    `json:"actor"`
	Source         string    `json:"source"`
	OccurredAt     time.Time `json:"occurred_at"`
	ObjectTitle    string    `json:"object_title"`
	DeepLink       string    `json:"deep_link"`
	AssigneeID     string    `json:"assignee_id,omitempty"`
	AssigneeName   string    `json:"assignee_name,omitempty"`
}

// Filter grenzt Ereignisse innerhalb genau einer Organisation ein.
// Leeres Limit liefert alle Treffer; alle Zeitgrenzen sind einschließlich.
type Filter struct {
	AgentID  string
	Action   string
	From     string
	To       string
	ObjectID string
	Limit    int
	Offset   int
}
