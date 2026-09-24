package agent

import "errors"

// ExecutionKind benennt den getrennten Ausführungsadapter.
type ExecutionKind string

// Status bezeichnet die fachliche Verfügbarkeit eines Agenten.
type Status string

const (
	Eino         ExecutionKind = "eino"
	CodexCLI     ExecutionKind = "codex_cli"
	StatusActive Status        = "active"
	StatusPaused Status        = "paused"
	StatusEnded  Status        = "ended"
)

var ErrNameRequired = errors.New("Agentenname ist erforderlich")
var ErrCapabilityDenied = errors.New("Fachfähigkeit ist nicht zulässig")
var ErrInvalidStatus = errors.New("Agentenstatus ist ungültig")

// Agent enthält nur fachliche Konfiguration, keine Zugangsdaten.
type Agent struct {
	ID             string        `json:"id"`
	OrganizationID string        `json:"organization_id"`
	Name           string        `json:"name"`
	Role           string        `json:"role"`
	Instructions   string        `json:"instructions"`
	ExecutionKind  ExecutionKind `json:"execution_kind"`
	TemplateID     string        `json:"template_id"`
	Capabilities   []string      `json:"capabilities"`
	Status         Status        `json:"status"`
}

// Readiness ist ein serverseitig abgeleiteter Startentscheid.
type Readiness struct {
	Ready  bool   `json:"ready"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

// Template ist eine kuratierte Einzelagentenvorlage.
type Template struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Role         string        `json:"role"`
	Instructions string        `json:"instructions"`
	Kind         ExecutionKind `json:"execution_kind"`
	Capabilities []string      `json:"capabilities"`
}

// TeamTemplate beschreibt auswählbare Agenten, ohne einen Teamimport auszulösen.
type TeamTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	TemplateIDs []string `json:"template_ids"`
}

// Catalog hält die kuratierten, ausführungsunabhängigen Profilvorlagen.
type Catalog struct {
	templates []Template
	teams     []TeamTemplate
}
