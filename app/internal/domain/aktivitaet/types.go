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
