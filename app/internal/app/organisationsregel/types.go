package organisationsregel

import (
	"errors"

	"agentcontrolplane/app/internal/domain/organisationsregel"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	portregel "agentcontrolplane/app/internal/port/organisationsregel"
)

var ErrAccessDenied = errors.New("access_denied")
var ErrNotFound = portregel.ErrNotFound
var ErrRevisionConflict = portregel.ErrRevisionConflict
var ErrInvalidTransition = organisationsregel.ErrInvalidTransition
var ErrApproverRequired = organisationsregel.ErrApproverRequired
var ErrRightsExpansion = organisationsregel.ErrRightsExpansion
var ErrWorkflowDSL = organisationsregel.ErrWorkflowDSL
var ErrRuleNotFound = organisationsregel.ErrRuleNotFound
var ErrInvalidOperation = organisationsregel.ErrInvalidOperation

type Rule = organisationsregel.Rule

// Change ist die vollständige, revisionsgebundene Eingabe für eine Kante.
type Change struct {
	Operation string `json:"operation"`
	Area      string `json:"area"`
	From      string `json:"from"`
	To        string `json:"to"`
	Approver  string `json:"approver"`
	Revision  int64  `json:"revision"`
}

// Snapshot zeigt die aktuelle Regelmenge einer Organisation.
type Snapshot struct {
	OrganizationID string   `json:"organization_id"`
	Revision       int64    `json:"revision"`
	Areas          []string `json:"areas"`
	Rules          []Rule   `json:"rules"`
}

// Preview erklärt die Wirkung, bevor eine Kante gespeichert wird.
type Preview struct {
	Operation     string   `json:"operation"`
	Area          string   `json:"area"`
	AffectedFlows []string `json:"affected_flows"`
	Rule          Rule     `json:"rule"`
	Revision      int64    `json:"revision"`
}

// Service ist die öffentliche App-Grenze für Arbeitsregeln.
type Service struct {
	organizations portorganisation.Store
	identity      portorganisation.Identity
	store         portregel.Store
}
