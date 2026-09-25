package ziel

import "errors"

// ErrNameRequired bezeichnet einen nach Trimmen leeren Zielnamen.
var ErrNameRequired = errors.New("Name ist erforderlich")

var ErrInvalidStatus = errors.New("Zielstatus ist ungültig")

// Goal ist ein dauerhaftes Ziel innerhalb einer Organisation.
type Goal struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Name           string  `json:"name"`
	ParentGoalID   *string `json:"parent_goal_id"`
	Status         string  `json:"status"`
}

// ProjectLink bezeichnet die gespeicherte Zuordnung eines Projekts zu einem Ziel.
type ProjectLink struct {
	ProjectID   string
	ProjectName string
	GoalID      string
}

// ProjectPath zeigt den vollständigen Zielpfad eines Projekts.
type ProjectPath struct {
	ProjectID   string   `json:"project_id"`
	ProjectName string   `json:"project_name"`
	GoalIDs     []string `json:"goal_ids"`
	GoalNames   []string `json:"goal_names"`
}
