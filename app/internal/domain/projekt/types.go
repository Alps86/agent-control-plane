package projekt

import "errors"

var (
	ErrNameRequired = errors.New("Name ist erforderlich")
	ErrGoalRequired = errors.New("Ziel ist erforderlich")
)

// Project ist ein Projekt mit einem Ziel in derselben Organisation.
type Project struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	GoalID         string `json:"goal_id"`
}
