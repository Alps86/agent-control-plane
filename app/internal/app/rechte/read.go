package rechte

import (
	domainrechte "agentcontrolplane/app/internal/domain/rechte"
	portrechte "agentcontrolplane/app/internal/port/rechte"
)

// NewReader bindet die serverseitige Quelle an die Anwendungsgrenze.
func NewReader(directory portrechte.Directory) *Reader {
	return &Reader{directory: directory}
}

// Read gibt Ressourcendaten nur nach eindeutiger Zuordnung und Freigabe zurück.
func (r *Reader) Read(id string) (domainrechte.Resource, Outcome) {
	if r == nil || r.directory == nil {
		return domainrechte.Resource{}, AccessDenied
	}

	actor, organizationID, valid := r.actorAndOrganization()
	if !valid {
		return domainrechte.Resource{}, AccessDenied
	}

	return r.readFor(id, actor, organizationID)
}

func (r *Reader) readFor(id string, actor domainrechte.Actor, organizationID string) (domainrechte.Resource, Outcome) {
	resource, found := r.directory.Resource(id)
	if !found || resource.ID != id || id == "" {
		return domainrechte.Resource{}, ResourceNotFound
	}

	assigned := actor.Kind == domainrechte.Agent && r.directory.Assigned(actor.ID, id)
	if !actor.CanRead(organizationID, resource, assigned) {
		return domainrechte.Resource{}, ResourceNotFound
	}

	return resource, Allowed
}

func (r *Reader) actorAndOrganization() (domainrechte.Actor, string, bool) {
	actors := r.directory.Actors()
	if len(actors) != 1 || !actors[0].Valid() {
		return domainrechte.Actor{}, "", false
	}

	organizations := r.directory.OrganizationIDs(actors[0].ID)
	if len(organizations) != 1 || organizations[0] == "" {
		return domainrechte.Actor{}, "", false
	}

	return actors[0], organizations[0], true
}
