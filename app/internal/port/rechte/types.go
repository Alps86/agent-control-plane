package rechte

import "agentcontrolplane/app/internal/domain/rechte"

// Directory liefert ausschließlich serverseitig ermittelte Rechte- und Ressourcendaten.
type Directory interface {
	Actors() []rechte.Actor
	OrganizationIDs(actorID string) []string
	Resource(id string) (rechte.Resource, bool)
	Assigned(agentID, resourceID string) bool
}
