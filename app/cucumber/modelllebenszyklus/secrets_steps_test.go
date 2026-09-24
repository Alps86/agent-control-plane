package modelllebenszyklus

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) registerSecrets(sc *godog.ScenarioContext) {
	sc.Step(`^die öffentliche Antwort enthält kein Anmeldegeheimnis$`, s.noSecrets)
	sc.Step(`^eine Codex-Abo-Sitzung und eine OpenRouter-Verbindung mit eindeutigen Testgeheimnissen bestehen$`, s.bothConnections)
	sc.Step(`^ich beide öffentlichen Settings-Ansichten und JSON-Antworten abrufe$`, s.readPublicViews)
	sc.Step(`^ich den OpenRouter-Schlüssel ersetze und die Codex-Sitzung erneuere$`, s.rotateBoth)
	sc.Step(`^enthalten keine dieser Antworten Tokens oder Schlüssel$`, s.noSecrets)
	sc.Step(`^die Verbindungen zeigen nur Referenzen und nichtgeheime Statusdaten$`, s.safeMetadata)
}

func (s *Suite) bothConnections() error {
	if err := s.validCodex(); err != nil {
		return err
	}
	return s.routerAction(http.MethodPost, routerPath, oldKey)
}

func (s *Suite) readPublicViews() error {
	if err := s.restartProcess(); err != nil {
		return err
	}
	for _, path := range []string{"/settings/modelle/codex", "/settings/modelle/codex/device/status", "/settings/modellanbieter/openrouter", routerPath} {
		if err := s.processGet(path); err != nil {
			return err
		}
		s.publicBodies = append(s.publicBodies, append([]byte(nil), s.response...))
	}
	return nil
}

func (s *Suite) rotateBoth() error {
	s.stopProcess()
	if err := s.routerAction(http.MethodPost, routerPath, newKey); err != nil {
		return err
	}
	if err := s.seedCodex(time.Minute); err != nil {
		return err
	}
	if err := s.resolveCodex(); err != nil {
		return err
	}
	if s.access[0].err != nil {
		return s.access[0].err
	}
	return s.readPublicViews()
}

func (s *Suite) noSecrets() error {
	for _, body := range append(s.publicBodies, s.response) {
		for _, secret := range []string{priorToken, priorRefresh, nextRefresh, oldKey, newKey, accountID, s.access[0].token} {
			if secret != "" && strings.Contains(string(body), secret) {
				return fmt.Errorf("öffentliches Geheimnis gefunden")
			}
		}
	}
	return nil
}

func (s *Suite) safeMetadata() error {
	if err := s.processGet(routerPath); err != nil {
		return err
	}
	if !strings.Contains(string(s.response), "openrouter-central") {
		return fmt.Errorf("Verbindungsreferenz fehlt")
	}
	if err := s.processCodexView(); err != nil {
		return err
	}
	if !strings.Contains(string(s.response), `"connected"`) {
		return fmt.Errorf("Codex-Status fehlt")
	}
	return s.noSecrets()
}
