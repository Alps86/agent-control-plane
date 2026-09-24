package rechte

// ActorKind benennt die beiden Akteursarten der Bootstrap-Rechtegrenze.
type ActorKind string

const (
	Operator ActorKind = "operator"
	Agent    ActorKind = "agent"
)

// Actor ist eine ausschließlich serverseitig ermittelte Identität.
type Actor struct {
	ID   string
	Kind ActorKind
}

// Resource ist die serverseitige Beispielressource für ARC-03.
type Resource struct {
	ID             string
	Name           string
	OrganizationID string
}
