package modellpruefung

import (
	"context"
	"fmt"
)

// NewVerbindungspruefung bindet den öffentlichen Verbindungsstatus ein.
func NewVerbindungspruefung(quelle VerbindungsstatusQuelle) *Verbindungspruefung {
	return &Verbindungspruefung{quelle: quelle}
}

// Lies führt nur den freigegebenen Statusabruf aus und redigiert das Ergebnis.
func (p *Verbindungspruefung) Lies(ctx context.Context, action string) (Verbindungszustand, error) {
	if action != VerbindungsstatusAktion || p == nil || p.quelle == nil {
		return Verbindungszustand{}, fmt.Errorf("unbekannte oder nicht verfügbare Fachaktion")
	}
	if err := ctx.Err(); err != nil {
		return Verbindungszustand{}, err
	}

	status, err := p.quelle.Status(ctx)
	if err != nil {
		return Verbindungszustand{}, fmt.Errorf("Verbindungsstatus nicht verfügbar")
	}

	return Verbindungszustand{Zustand: status.State}, nil
}
