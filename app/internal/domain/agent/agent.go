package agent

import "strings"

// NewAgent setzt ein geprüfte Einzelvorlage in eine Agentenkonfiguration um.
func NewAgent(id, organizationID, name string, template Template) (Agent, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Agent{}, ErrNameRequired
	}

	return Agent{
		ID: id, OrganizationID: organizationID, Name: name,
		Role: template.Role, Instructions: template.Instructions,
		ExecutionKind: template.Kind, TemplateID: template.ID,
		Capabilities: append([]string{}, template.Capabilities...),
		Status:       StatusActive,
	}, nil
}

// Valid erlaubt ausschließlich die drei gespeicherten Agentenzustände.
func (s Status) Valid() bool {
	return s == StatusActive || s == StatusPaused || s == StatusEnded
}

// AllowsCapability lässt nur exakt benannte kuratierte Fachfähigkeiten zu.
func (a Agent) AllowsCapability(name string) bool {
	for _, capability := range a.Capabilities {
		if name == capability {
			return true
		}
	}

	return false
}
