package organisationsregel

import "errors"

var ErrInvalidTransition = errors.New("invalid_transition")
var ErrApproverRequired = errors.New("approver_required")
var ErrRightsExpansion = errors.New("rights_expansion")
var ErrWorkflowDSL = errors.New("workflow_dsl_not_supported")
var ErrRuleNotFound = errors.New("rule_not_found")
var ErrInvalidOperation = errors.New("invalid_operation")

// Rule beschreibt eine einzelne, ausdrücklich freigegebene Fachkante.
type Rule struct {
	Area     string `json:"area"`
	From     string `json:"from"`
	To       string `json:"to"`
	Approver string `json:"approver"`
}

// Policy enthält die versionierten, organisationsgebundenen Fachkanten.
type Policy struct {
	Revision int64  `json:"revision"`
	Rules    []Rule `json:"rules"`
}
