package system

import "agentcontrolplane/app/internal/domain/system"

// NewProbe erzeugt die Status-Probe.
func NewProbe() *Probe {
	return &Probe{}
}

// Status liefert die Bereitschaft des gestarteten Prozesses.
func (p *Probe) Status() system.Status {
	return system.NewStatus(true)
}
