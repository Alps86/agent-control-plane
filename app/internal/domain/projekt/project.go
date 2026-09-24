package projekt

import "strings"

// NewProject validiert die Pflichtfelder eines neuen Projekts.
func NewProject(id, organizationID, name, description, goalID string) (Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, ErrNameRequired
	}

	goalID = strings.TrimSpace(goalID)
	if goalID == "" {
		return Project{}, ErrGoalRequired
	}

	return Project{ID: id, OrganizationID: organizationID, Name: name,
		Description: description, GoalID: goalID}, nil
}
