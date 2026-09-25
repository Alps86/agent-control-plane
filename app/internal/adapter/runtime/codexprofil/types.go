package codexprofil

import (
	"errors"
	"net"

	appprofil "agentcontrolplane/app/internal/app/codexprofil"
)

var ErrIntegrity = errors.New("Codex-Arbeitsbereich verletzt die Integrität")

// Locator hält nur den konfigurierten Datenbankpfad, nie einen Clientpfad.
type Locator struct {
	databasePath string
}

// Broker ist eine je Prüfaufruf begrenzte Host-Grenze für die eine Go-Aktion.
type Broker struct {
	service        *appprofil.Service
	organizationID string
	agentID        string
	listener       net.Listener
	directory      string
	socketPath     string
}

// Factory erstellt pro Aufruf einen gebundenen Host-Broker.
type Factory struct {
	service *appprofil.Service
}
