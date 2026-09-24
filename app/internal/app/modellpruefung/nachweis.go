package modellpruefung

// NewNachweis erstellt den noch offenen technischen Modellnachweis.
func NewNachweis() *Nachweis {
	return &Nachweis{}
}

// ErfasseAgentenlauf zählt einen Codex-Agentenlauf ohne Eino-Modellbeleg.
func (n *Nachweis) ErfasseAgentenlauf() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.agentenlaeufe++
}

// ErfasseTransport bewertet nur vollständig beobachtete Modellrunden.
func (n *Nachweis) ErfasseTransport(beleg Transportbeleg) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if beleg.Anbieter == "" || beleg.Modell == "" || beleg.Anfragekennung == "" || !beleg.Antwort {
		return
	}

	n.modellbelege++
	n.vollstaendig = n.vollstaendig ||
		(beleg.Stream && beleg.Toolrunde && beleg.Folgeaufruf && beleg.Abbruch && beleg.Limitfehler)
}

// Status zeigt an, ob die Eino-POC-Abnahme technisch belegt ist.
func (n *Nachweis) Status() Nachweisstatus {
	n.mu.Lock()
	defer n.mu.Unlock()
	return Nachweisstatus{Offen: !n.vollstaendig, Agentenlaeufe: n.agentenlaeufe, Modellbelege: n.modellbelege}
}
