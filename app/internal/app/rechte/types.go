package rechte

import "agentcontrolplane/app/internal/port/rechte"

// Outcome hält erlaubte Ausgabe und die zwei öffentlichen Verweigerungsklassen getrennt.
type Outcome string

const (
	Allowed          Outcome = "allowed"
	AccessDenied     Outcome = "access_denied"
	ResourceNotFound Outcome = "resource_not_found"
)

// Reader erzwingt die zentrale Rechteentscheidung vor der Ressourcenausgabe.
type Reader struct {
	directory rechte.Directory
}
