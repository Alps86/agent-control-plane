package modelllebenszyklus

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

const routerPath = "/api/settings/modellanbieter/openrouter"
const checkPath = routerPath + "/pruefen"

func (s *Suite) registerOpenRouter(sc *godog.ScenarioContext) {
	sc.Step(`^die zentrale OpenRouter-Verbindung enthält einen später widerrufenen Testschlüssel$`, s.oldRouterKey)
	sc.Step(`^ich ihren Status über die öffentliche Settings-Grenze prüfe$`, s.checkRouter)
	sc.Step(`^wird der Schlüssel als nicht einsatzbereit mit Ersatzhinweis angezeigt$`, s.invalidRouter)
	sc.Step(`^ich über die öffentliche Settings-Grenze einen neuen Testschlüssel speichere$`, s.replaceRouter)
	sc.Step(`^ich den Status erneut prüfe$`, s.checkRouter)
	sc.Step(`^bleibt die Verbindungsreferenz gleich und der neue Schlüssel ist einsatzbereit$`, s.readyRouter)
	sc.Step(`^der kontrollierte OpenRouter-Anbieter erhält nur den neuen Schlüssel bei der zweiten Prüfung$`, s.onlyNewKey)
}

func (s *Suite) oldRouterKey() error {
	if err := s.setup(); err != nil {
		return err
	}
	if err := s.routerAction(http.MethodPost, routerPath, oldKey); err != nil {
		return err
	}
	view, err := s.routerView()
	if err != nil {
		return err
	}
	s.priorReference = view.Reference
	return nil
}

func (s *Suite) checkRouter() error {
	return s.routerAction(http.MethodPost, checkPath, "")
}

func (s *Suite) invalidRouter() error {
	view, err := s.routerView()
	if err != nil {
		return err
	}
	if view.Status != "nicht einsatzbereit" || !strings.Contains(view.StatusDetail, "Ersetzen") {
		return fmt.Errorf("Widerruf nicht als Ersatzbedarf angezeigt: %+v", view)
	}
	s.firstProbeCount = len(s.providerState.snapshot())
	return s.noSecrets()
}

func (s *Suite) replaceRouter() error {
	return s.routerAction(http.MethodPost, routerPath, newKey)
}

func (s *Suite) readyRouter() error {
	view, err := s.routerView()
	if err != nil {
		return err
	}
	if view.Reference != s.priorReference || view.Status != "einsatzbereit" {
		return fmt.Errorf("rotierte Verbindung nicht einsatzbereit: %+v", view)
	}
	return s.noSecrets()
}

func (s *Suite) onlyNewKey() error {
	calls := s.providerState.snapshot()
	if len(calls) != s.firstProbeCount+1 || calls[s.firstProbeCount] != "Bearer "+newKey {
		return fmt.Errorf("zweite Anbieterprüfung verwendete nicht ausschließlich neuen Schlüssel: %v", calls)
	}
	return nil
}
