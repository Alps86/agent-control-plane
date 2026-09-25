package berichtsweg

import domainagent "agentcontrolplane/app/internal/domain/agent"

// Line beschreibt einen Agenten und genau seinen unmittelbaren Vorgesetzten.
// Eine leere ParentID bezeichnet eine Wurzel im Organigramm.
type Line struct {
	AgentID       string                    `json:"agent_id"`
	Name          string                    `json:"name"`
	ParentID      string                    `json:"parent_id"`
	ExecutionKind domainagent.ExecutionKind `json:"execution_kind"`
}

// Chart enthält alle Agenten einer zugänglichen Organisation.
type Chart struct {
	OrganizationID string `json:"organization_id"`
	Lines          []Line `json:"lines"`
}
