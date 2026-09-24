package organisation

import "strings"

// NewOrganization validiert den Namen und setzt den restriktiven Standard.
func NewOrganization(id, name, description string) (Organization, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Organization{}, ErrNameRequired
	}

	return Organization{
		ID: id, Name: name, Description: description,
		WorkflowPolicy: WorkflowPolicy{
			WorkTransitions: []string{}, DelegationTransitions: []string{},
		},
	}, nil
}
