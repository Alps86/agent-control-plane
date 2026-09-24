package codexprofil

import (
	"errors"

	domain "agentcontrolplane/app/internal/domain/codexprofil"
	portagent "agentcontrolplane/app/internal/port/agent"
	portprofil "agentcontrolplane/app/internal/port/codexprofil"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

var ErrAccessDenied = errors.New("access_denied")
var ErrNotFound = portagent.ErrNotFound
var ErrExecutionKind = errors.New("Codex-CLI-Profil nicht vorhanden")
var ErrWriteWithoutWorkspace = domain.ErrWriteWithoutWorkspace
var ErrWorkspaceDisabled = errors.New("Codex-Arbeitsbereich ist deaktiviert")
var ErrWriteDisabled = errors.New("Codex-Schreiben ist deaktiviert")
var ErrActionDenied = errors.New("Codex-Aktion ist nicht freigegeben")

// Service verbindet Betreiberidentität, scoped Agentenspeicher und Arbeitsbereich.
type Service struct {
	store         portprofil.Store
	agents        portagent.Store
	identity      portorganisation.Identity
	organizations portorganisation.Store
	workspace     portprofil.Workspace
}
