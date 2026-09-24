package aufgabe

import "errors"

var ErrTitleRequired = errors.New("Titel ist erforderlich")
var ErrAssigneeRequired = errors.New("Zuständiger Agent ist erforderlich")
var ErrInvalidPriority = errors.New("Priorität ist ungültig")

const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
	StatusOpen     = "open"
)

// Task ist eine Aufgabe mit genau einem zuständigen Agenten.
type Task struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Priority       string `json:"priority"`
	AssigneeID     string `json:"assignee_id"`
	Status         string `json:"status"`
}
