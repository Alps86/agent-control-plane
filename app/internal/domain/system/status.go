package system

// NewStatus erzeugt einen Betriebsstatus.
func NewStatus(ready bool) Status {
	return Status{Ready: ready}
}
