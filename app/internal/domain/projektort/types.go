package projektort

import "errors"

const ManagedDirectory = "managed_directory"

var ErrKindInvalid = errors.New("Nur ein privater, appverwalteter Projektordner ist erlaubt")

// Binding speichert eine ausdrückliche Freigabe, niemals einen Clientpfad.
type Binding struct {
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	Kind           string `json:"kind"`
}
