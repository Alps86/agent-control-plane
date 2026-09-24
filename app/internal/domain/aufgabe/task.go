package aufgabe

import "strings"

// NewTask prüft die Pflichtfelder und setzt den anfänglichen Status.
func NewTask(id, organizationID, projectID, title, description, priority, assigneeID string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrTitleRequired
	}

	assigneeID = strings.TrimSpace(assigneeID)
	if assigneeID == "" {
		return Task{}, ErrAssigneeRequired
	}

	if !(Task{Priority: priority}).ValidPriority() {
		return Task{}, ErrInvalidPriority
	}

	return Task{ID: id, OrganizationID: organizationID, ProjectID: projectID,
		Title: title, Description: description, Priority: priority,
		AssigneeID: assigneeID, Status: StatusOpen}, nil
}

// ValidPriority kennt die feste Prioritätsmenge der ersten Aufgabenversion.
func (t Task) ValidPriority() bool {
	return t.Priority == PriorityLow || t.Priority == PriorityNormal ||
		t.Priority == PriorityHigh || t.Priority == PriorityUrgent
}
