package modellverbindung

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (s *Suite) registerTimingSteps(sc *godog.ScenarioContext) {
	sc.Step(`^der Anbieter liefert ein überhöhtes Gerätecode-Abfrageintervall$`, s.hugeInterval)
	sc.Step(`^ich den Verbindungsstatus zweimal unmittelbar prüfe$`, s.statusTwice)
	sc.Step(`^bleibt die Anmeldung laufend ohne zweite Anbieterabfrage$`, s.pendingOnePoll)
	sc.Step(`^der Anbieter beim Abbruch vorübergehend nicht erreichbar ist$`, s.providerOutage)
	sc.Step(`^bleibt die Anmeldung abgebrochen ohne weitere Anbieterabfrage$`, s.cancelledNoExtraPoll)
}

func (s *Suite) hugeInterval() error {
	s.issuer.interval = "9223372036854775807"
	return nil
}

func (s *Suite) statusTwice() error {
	if err := s.status(); err != nil {
		return err
	}
	return s.status()
}

func (s *Suite) pendingOnePoll() error {
	if err := s.pending(); err != nil {
		return err
	}
	if s.issuer.polls != 1 {
		return fmt.Errorf("provider was polled %d times before interval", s.issuer.polls)
	}
	return nil
}

func (s *Suite) providerOutage() error {
	s.issuer.pollOutcome = "outage"
	return nil
}

func (s *Suite) cancelledNoExtraPoll() error {
	if err := s.cancelled(); err != nil {
		return err
	}
	if s.issuer.polls != 1 {
		return fmt.Errorf("provider was polled %d times despite cancellation", s.issuer.polls)
	}
	return nil
}
