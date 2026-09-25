package modellpruefung

import (
	"context"
	"fmt"
	"strings"
)

// NewPruefung baut die öffentliche Fachgrenze mit einer Statusquelle.
func NewPruefung(quelle StatusQuelle) *Pruefung {
	return &Pruefung{quelle: quelle, rechte: make(map[string]map[string]bool)}
}

// Erlaube gibt einem Agenten den Statuszugriff auf ein bestimmtes Projekt.
func (p *Pruefung) Erlaube(agent, projekt string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.rechte[agent] == nil {
		p.rechte[agent] = make(map[string]bool)
	}

	p.rechte[agent][projekt] = true
}

// Rufe führt nur die registrierte, berechtigte Fachaktion aus.
func (p *Pruefung) Rufe(ctx context.Context, agent, aktion, projekt string) (Status, error) {
	if err := p.pruefe(agent, aktion, projekt); err != nil {
		return Status{}, err
	}

	status, err := p.quelle.Status(ctx, projekt)
	p.vermerke(Aufruf{Agent: agent, Aktion: aktion, Projekt: projekt, Berechtigt: true, Ausgefuehrt: err == nil})
	return status, err
}

func (p *Pruefung) pruefe(agent, aktion, projekt string) error {
	if aktion != StatusAktion || strings.TrimSpace(projekt) == "" || strings.TrimSpace(agent) == "" {
		p.vermerke(Aufruf{Agent: agent, Aktion: aktion, Projekt: projekt, Grund: "ungültige Fachaktion oder Parameter"})
		return fmt.Errorf("ungültige Fachaktion oder Parameter")
	}

	p.mu.Lock()
	erlaubt := p.rechte[agent][projekt]
	p.mu.Unlock()
	if !erlaubt {
		p.vermerke(Aufruf{Agent: agent, Aktion: aktion, Projekt: projekt, Grund: "kein Projektrecht"})
		return fmt.Errorf("kein Projektrecht für %q", projekt)
	}

	return nil
}

// Aufrufe liefert einen unabhängigen Schnappschuss der Prüfentscheidungen.
func (p *Pruefung) Aufrufe() []Aufruf {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]Aufruf(nil), p.aufrufe...)
}

func (p *Pruefung) vermerke(aufruf Aufruf) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.aufrufe = append(p.aufrufe, aufruf)
}
