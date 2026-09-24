package system

import "agentcontrolplane/app/internal/domain/system"

// Probe liefert den gegenwärtigen Betriebsstatus.
type Probe interface {
	Status() system.Status
}
