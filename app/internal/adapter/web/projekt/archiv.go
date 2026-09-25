package projekt

import appprojektarchiv "agentcontrolplane/app/internal/app/projektarchiv"

// SetArchive bindet den gespeicherten Projektstatus an die Detailansicht.
func (h *Handler) SetArchive(archive *appprojektarchiv.Service) {
	h.archive = archive
}
