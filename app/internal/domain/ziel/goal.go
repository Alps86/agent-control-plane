package ziel

import "strings"

// NewRootGoal validiert und erzeugt ein Stammziel ohne übergeordnetes Ziel.
func (g Goal) NewRootGoal(id, organizationID, name string) (Goal, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Goal{}, ErrNameRequired
	}

	return Goal{ID: id, OrganizationID: organizationID, Name: name, Status: "planned"}, nil
}

// NewChildGoal validiert ein neues Teilziel ohne nachträgliche Elternänderung.
func (g Goal) NewChildGoal(id, organizationID, parentGoalID, name string) (Goal, error) {
	goal, err := g.NewRootGoal(id, organizationID, name)
	if err != nil {
		return Goal{}, err
	}

	goal.ParentGoalID = &parentGoalID
	return goal, nil
}

// WithStatus akzeptiert nur die fachlich bekannten ausdrücklichen Statuswerte.
func (g Goal) WithStatus(status string) (Goal, error) {
	if status != "planned" && status != "active" && status != "achieved" && status != "cancelled" {
		return Goal{}, ErrInvalidStatus
	}

	g.Status = status
	return g, nil
}
