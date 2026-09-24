package modellzugang

import (
	"fmt"

	"agentcontrolplane/app/internal/app/modellpruefung"
)

func (s *Suite) partialGateEvidence() error {
	s.local = true
	s.nachweis = modellpruefung.NewNachweis()
	s.transportBeleg = modellpruefung.Transportbeleg{Anbieter: "lokaler Testanbieter", Modell: "testmodell",
		Anfragekennung: "offline-request-1", Antwort: true, Stream: true, Toolrunde: true}
	s.nachweis.ErfasseTransport(s.transportBeleg)
	return nil
}

func (s *Suite) partialGateOpen() error {
	if !s.nachweis.Status().Offen {
		return fmt.Errorf("unvollständiger Beleg schloss App-Gate")
	}

	return nil
}

func (s *Suite) completeGateEvidence() error {
	s.transportBeleg.Folgeaufruf, s.transportBeleg.Abbruch, s.transportBeleg.Limitfehler = true, true, true
	s.nachweis.ErfasseTransport(s.transportBeleg)
	return nil
}

func (s *Suite) localGateClosed() error {
	if s.nachweis.Status().Offen {
		return fmt.Errorf("vollständiger lokaler Beleg hielt App-Gate offen")
	}

	return nil
}
