package modellzugang

import (
	"fmt"
	"os"
	"strings"

	"agentcontrolplane/app/internal/app/modellpruefung"
)

func (s *Suite) liveAccount() error {
	return s.prepareLive()
}

func (s *Suite) livePermission(agent, aktion, projekt string) error {
	if agent != "Mira" || aktion != modellpruefung.StatusAktion {
		return fmt.Errorf("unbekannter Agent oder Fachaktion")
	}

	s.pruefung.Erlaube(agent, projekt)
	return nil
}

func (s *Suite) liveFixture(projekt, wert, marker string) error {
	s.fixture.status = modellpruefung.Status{Projekt: projekt, Wert: wert, Pruefkennung: marker}
	return nil
}

func (s *Suite) liveQuestion(agent, projekt string) error {
	if agent != "Mira" {
		return fmt.Errorf("unerwarteter Agent")
	}

	return s.runLive("Lies mit der registrierten Fachaktion den Status und die Prüfkennung von " + projekt + ".")
}

func (s *Suite) streamed() error {
	if s.modelErr != nil || s.chunks == 0 {
		return fmt.Errorf("kein erfolgreicher Antwortstream: %v", s.modelErr)
	}

	return nil
}

func (s *Suite) toolReceived(aktion, projekt string) error {
	if aktion != modellpruefung.StatusAktion || s.fixture.Accesses() != 1 {
		return fmt.Errorf("registrierter Fachaufruf fehlt")
	}

	for _, aufruf := range s.pruefung.Aufrufe() {
		if aufruf.Projekt == projekt && aufruf.Aktion == aktion {
			return nil
		}
	}

	return fmt.Errorf("Fachaktion für %q fehlt", projekt)
}

func (s *Suite) actionChecked() error {
	accesses := s.pruefung.Aufrufe()
	if len(accesses) != 1 || !accesses[0].Berechtigt || !accesses[0].Ausgefuehrt {
		return fmt.Errorf("fachliche Prüfung oder Ausführung fehlt")
	}

	return nil
}

func (s *Suite) allowedResult(wert, marker string) error {
	if s.fixture.Accesses() != 1 || s.fixture.status.Wert != wert || s.fixture.status.Pruefkennung != marker {
		return fmt.Errorf("abweichendes Fachresultat")
	}

	return nil
}

func (s *Suite) toolContinued() error {
	if len(s.trace.Requests()) < 2 {
		return fmt.Errorf("keine Abo-Modellfortsetzung")
	}

	return nil
}

func (s *Suite) finalAnswer(wert, marker string) error {
	if !strings.Contains(s.answer, wert) || !strings.Contains(s.answer, marker) {
		return fmt.Errorf("Modellantwort enthält Fachresultat nicht")
	}

	return nil
}

func (s *Suite) liveReport() error {
	if !s.live || len(s.trace.Requests()) < 2 || len(s.observations.Entries()) < 2 {
		return fmt.Errorf("Eino-Abo-Nachweis fehlt")
	}

	if err := s.checkLiveTrace(); err != nil {
		return err
	}

	s.transportBeleg.Anbieter = "codex-abo"
	s.transportBeleg.Modell = s.observedModel()
	s.transportBeleg.Anfragekennung = s.responseID()
	s.transportBeleg.Antwort, s.transportBeleg.Stream, s.transportBeleg.Toolrunde = true, true, true
	return s.syncEvidence()
}

func (s *Suite) checkLiveTrace() error {
	requests, responses := s.trace.Requests(), s.observations.Entries()
	if !requests[1].hasToolOutput || !requests[1].hasMarker {
		return fmt.Errorf("Toolergebnis im zweiten Request fehlt")
	}

	for index := 0; index < 2; index++ {
		if !requests[index].accountMatched || !requests[index].hasBearer || requests[index].requestID == "" ||
			responses[index].RequestID != requests[index].requestID ||
			responses[index].Model != os.Getenv("MS01_LIVE_MODEL") || responses[index].Status != "completed" {
			return fmt.Errorf("Verbindung oder Providerbeleg unvollständig")
		}
	}

	if requests[0].requestID == requests[1].requestID {
		return fmt.Errorf("Anbieter-Request-IDs sind nicht getrennt")
	}

	return s.noAPIFallback()
}

func (s *Suite) priorAnswer(agent, projekt string) error {
	if err := s.livePermission(agent, modellpruefung.StatusAktion, projekt); err != nil {
		return err
	}

	if err := s.runLive("Lies Status und Prüfkennung von " + projekt + " mit der Fachaktion."); err != nil {
		return err
	}

	s.priorRequests = len(s.trace.Requests())
	return s.finalAnswer("in Arbeit", "MS01-WEBSITE-4711")
}

func (s *Suite) followup(agent, frage string) error {
	if agent != "Mira" {
		return fmt.Errorf("unerwarteter Agent")
	}

	return s.runLive(frage)
}

func (s *Suite) furtherCall() error {
	if len(s.trace.Requests()) <= s.priorRequests {
		return fmt.Errorf("Anschlussfrage ohne neuen Anbieteraufruf")
	}

	return nil
}

func (s *Suite) followupAnswer(marker, projekt string) error {
	if !strings.Contains(s.answer, marker) || !strings.Contains(s.answer, projekt) {
		return fmt.Errorf("Anschlussantwort ohne Gesprächsbezug")
	}

	return nil
}

func (s *Suite) sameConversation() error {
	if len(s.messages) < 4 || len(s.trace.Requests()) <= s.priorRequests {
		return fmt.Errorf("Gespräch oder getrennter Anbieteraufruf fehlt")
	}

	s.transportBeleg.Folgeaufruf = true
	return s.syncEvidence()
}
