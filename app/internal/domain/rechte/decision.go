package rechte

// NewActor erzeugt eine serverseitig ermittelte Identität.
func NewActor(id string, kind ActorKind) Actor {
	return Actor{ID: id, Kind: kind}
}

// NewResource erzeugt eine serverseitige Beispielressource.
func NewResource(id, name, organizationID string) Resource {
	return Resource{ID: id, Name: name, OrganizationID: organizationID}
}

// Valid verwirft leere und unbekannte Akteursidentitäten.
func (a Actor) Valid() bool {
	return a.ID != "" && (a.Kind == Operator || a.Kind == Agent)
}

// CanRead entscheidet ausschließlich über das Lesen einer Beispielressource.
func (a Actor) CanRead(organizationID string, resource Resource, assigned bool) bool {
	if !a.Valid() || organizationID == "" || resource.ID == "" || resource.OrganizationID != organizationID {
		return false
	}

	if a.Kind == Operator {
		return true
	}

	return assigned
}
