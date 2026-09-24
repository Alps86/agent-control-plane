package organisation

import "agentcontrolplane/app/internal/domain/rechte"

// NewLocalIdentity bindet den lokalen Server an genau einen Betreiber.
func NewLocalIdentity() *LocalIdentity {
	return &LocalIdentity{actor: rechte.NewActor("local-operator", rechte.Operator)}
}

// Actors liefert eine frische Kopie der serverseitigen Identität.
func (i *LocalIdentity) Actors() []rechte.Actor {
	if i == nil {
		return nil
	}

	return []rechte.Actor{i.actor}
}
