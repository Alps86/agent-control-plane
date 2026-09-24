package codexprofil

import "errors"

var ErrWriteWithoutWorkspace = errors.New("write_enabled requires workspace_enabled")

// Profile enthält nur betreiberseitig freigegebene Schalter und logische Zugehörigkeit.
type Profile struct {
	OrganizationID   string `json:"organization_id"`
	AgentID          string `json:"agent_id"`
	WorkspaceEnabled bool   `json:"workspace_enabled"`
	WriteEnabled     bool   `json:"write_enabled"`
}

// Input enthält die beiden einzigen vom Client wählbaren Werte.
type Input struct {
	WorkspaceEnabled bool `json:"workspace_enabled"`
	WriteEnabled     bool `json:"write_enabled"`
}
