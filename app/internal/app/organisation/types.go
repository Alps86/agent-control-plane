package organisation

import (
	"errors"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var (
	ErrAccessDenied = errors.New("access_denied")
	ErrNameRequired = domainorganisation.ErrNameRequired
	ErrNameConflict = portorganisation.ErrNameConflict
	ErrNotFound     = portorganisation.ErrNotFound
)

// Service ist die schmale öffentliche Organisationsanwendungsgrenze.
type Service struct {
	store    portorganisation.Store
	identity portorganisation.Identity
}

// LocalIdentity hält die feste Identität des lokalen Betreiberprozesses.
type LocalIdentity struct {
	actor rechte.Actor
}
