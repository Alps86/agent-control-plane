package lauf

import "strings"

// FailPreparation beschreibt ausschließlich einen Fehler vor Adapterstart.
func (r Run) FailPreparation(reason string) (Run, error) {
	if r.Status != Reserviert {
		return Run{}, ErrInvalidTransition
	}

	if strings.TrimSpace(reason) == "" {
		return Run{}, ErrReasonRequired
	}

	r.Status = Fehlgeschlagen
	r.FailureReason = reason
	return r, nil
}
