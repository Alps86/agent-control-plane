package prozess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ein lokaler Server mit kontrollierten Modellanbietern läuft$`, s.start)
	sc.Step(`^die zentrale OpenRouter-Verbindung wurde mit einem synthetischen Schlüssel eingerichtet und geprüft$`, s.connectRouter)
	sc.Step(`^ich "Settings > Modellanbieter" im echten Browser öffne$`, s.openBrowser)
	sc.Step(`^sehe ich Codex-Abo und OpenRouter mit ihren getrennten Status und nur vorhandenen Verbindungsreferenzen$`, s.bothCards)
	sc.Step(`^ich sehe für Codex-Abo nur die Gerätecode-Anmeldung als Authentifizierungsart$`, s.codexAuth)
	sc.Step(`^ich sehe für OpenRouter nur den API-Schlüssel als Authentifizierungsart$`, s.routerAuth)
	sc.Step(`^weder der sichtbare Seiteninhalt noch der DOM enthalten Token, Schlüssel oder Refresh-Token$`, s.noSecrets)
	sc.Step(`^die gemeinsame Settings-Ansicht zeigt eine geprüfte OpenRouter-Verbindung$`, s.connectedView)
	sc.Step(`^ich OpenRouter über die vorhandene Settings-Aktion trenne$`, s.browserDisconnect)
	sc.Step(`^zeigt die gemeinsame Ansicht OpenRouter als getrennt und nicht einsatzbereit$`, s.disconnectedView)
	sc.Step(`^sie zeigt Codex-Abo weiterhin mit seinem tatsächlichen Status$`, s.codexActual)
	sc.Step(`^ich den Server mit denselben Daten- und Credential-Pfaden neu starte$`, s.restart)
	sc.Step(`^bleibt OpenRouter als getrennt und nicht einsatzbereit sichtbar$`, s.disconnectedView)
}

func (s *Suite) request(method, path, body string) error {
	request, err := http.NewRequest(method, "http://"+s.address+path, strings.NewReader(body))
	if err != nil {
		return err
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet {
		request.Header.Set("Origin", "http://"+s.address)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	s.code = response.StatusCode
	s.response, err = io.ReadAll(io.LimitReader(response.Body, 1<<20))
	return err
}

func (s *Suite) connectRouter() error {
	body, _ := json.Marshal(map[string]string{"key": key})
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter", string(body)); err != nil {
		return err
	}
	if s.code != http.StatusOK {
		return fmt.Errorf("Speichern HTTP %d: %s", s.code, s.response)
	}
	if err := s.request(http.MethodPost, "/api/settings/modellanbieter/openrouter/pruefen", ""); err != nil {
		return err
	}
	if s.code != http.StatusOK || !bytes.Contains(s.response, []byte("einsatzbereit")) {
		return fmt.Errorf("Prüfen HTTP %d: %s", s.code, s.response)
	}
	return nil
}

func (s *Suite) connectedView() error {
	if err := s.connectRouter(); err != nil {
		return err
	}
	return s.openBrowser()
}

func (s *Suite) bothCards() error {
	if !s.page.OK || s.page.Codex != "Getrennt" || s.page.OpenRouter != "Verbunden" || s.page.OpenRouterRef != "openrouter-central" || s.page.CodexRef != "" {
		return fmt.Errorf("Browserstatus: %+v", s.page)
	}
	return nil
}

func (s *Suite) codexAuth() error {
	if s.page.CodexAuth != "Gerätecode" {
		return fmt.Errorf("Codex-Auth %q", s.page.CodexAuth)
	}
	return nil
}

func (s *Suite) routerAuth() error {
	if s.page.OpenRouterAuth != "API-Schlüssel" {
		return fmt.Errorf("OpenRouter-Auth %q", s.page.OpenRouterAuth)
	}
	return nil
}

func (s *Suite) noSecrets() error {
	for _, secret := range []string{key, "story79-synthetic-access-token", "story79-synthetic-refresh-token"} {
		if strings.Contains(s.page.HTML, secret) {
			return fmt.Errorf("Browser-DOM enthält synthetisches Secret")
		}
	}
	return nil
}

func (s *Suite) disconnectedView() error {
	if err := s.openBrowser(); err != nil {
		return err
	}
	if s.page.OpenRouter != "Getrennt" || s.page.OpenRouterRef != "" {
		return fmt.Errorf("OpenRouter bleibt bereit: %+v", s.page)
	}
	return nil
}

func (s *Suite) codexActual() error {
	if s.page.Codex != "Getrennt" {
		return fmt.Errorf("Codex-Status wechselte: %q", s.page.Codex)
	}
	return nil
}
