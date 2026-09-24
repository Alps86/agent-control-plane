package ziel

import "strings"

// NewRootGoal validiert und erzeugt ein Stammziel ohne übergeordnetes Ziel.
func (g Goal) NewRootGoal(id, organizationID, name string) (Goal, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Goal{}, ErrNameRequired
	}

	return Goal{ID: id, OrganizationID: organizationID, Name: name}, nil
}
