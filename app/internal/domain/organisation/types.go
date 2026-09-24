package organisation

import "errors"

// ErrNameRequired bezeichnet einen nach Trimmen leeren Organisationsnamen.
var ErrNameRequired = errors.New("Name ist erforderlich")

// WorkflowPolicy begrenzt erlaubte Arbeits- und Delegationsübergänge.
type WorkflowPolicy struct {
	WorkTransitions       []string `json:"work_transitions"`
	DelegationTransitions []string `json:"delegation_transitions"`
}

// Organization ist der dauerhaft gespeicherte Organisationskontext.
type Organization struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	WorkflowPolicy WorkflowPolicy `json:"workflow_policy"`
}
