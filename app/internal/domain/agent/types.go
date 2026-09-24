package agent

import "errors"

// ExecutionKind benennt den getrennten Ausführungsadapter.
type ExecutionKind string

const (
	Eino     ExecutionKind = "eino"
	CodexCLI ExecutionKind = "codex_cli"
)

var ErrNameRequired = errors.New("Agentenname ist erforderlich")
var ErrCapabilityDenied = errors.New("Fachfähigkeit ist nicht zulässig")

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
