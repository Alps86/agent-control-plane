package ziel

import "errors"

// ErrNameRequired bezeichnet einen nach Trimmen leeren Zielnamen.
var ErrNameRequired = errors.New("Name ist erforderlich")

// Goal ist ein dauerhaftes Ziel innerhalb einer Organisation.
type Goal struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Name           string  `json:"name"`
	ParentGoalID   *string `json:"parent_goal_id"`
}
