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
		Capabilities: append([]string(nil), template.Capabilities...),
	}, nil
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
