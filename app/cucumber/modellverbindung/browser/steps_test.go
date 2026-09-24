//go:build browser

package browser

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"
	"testing"

	"agentcontrolplane/ui/bridge"
	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite { return &Suite{t: t} }

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	var err error
	s.bridge, err = bridge.New()
	if err != nil {
		s.t.Fatal(err)
	}
	s.registerSteps(sc)
}

func (s *Suite) registerSteps(sc *godog.ScenarioContext) {
	s.registerGiven(sc)
	s.registerWhen(sc)
	s.registerThen(sc)
}

func (s *Suite) registerGiven(sc *godog.ScenarioContext) {
	sc.Step("^ich öffne die Modelleinstellungen ohne Codex-Abo-Verbindung$", s.idle)
	sc.Step("^eine Gerätecode-Anmeldung läuft in den Modelleinstellungen$", s.pending)
	sc.Step("^der Anbieter hat die Gerätecode-Anmeldung deaktiviert$", s.unavailable)
	sc.Step("^meine Codex-Abo-Verbindung war verbunden$", s.connected)
}

func (s *Suite) registerWhen(sc *godog.ScenarioContext) {
	sc.Step("^ich \"Codex-Abo verbinden\" wähle$", s.start)
	sc.Step("^ich in den Modelleinstellungen \"Codex-Abo verbinden\" wähle$", s.start)
	sc.Step("^ich den Verbindungsstatus aktualisiere$", s.refresh)
	sc.Step("^der Anbieter die Anmeldung bestätigt und ich den Status aktualisiere$", s.confirm)
	sc.Step("^ich \"Anmeldung abbrechen\" wähle$", s.cancel)
	sc.Step("^der Anbieter die Anmeldung mit \"([^\"]*)\" beendet und ich den Status aktualisiere$", s.failure)
	sc.Step("^der Anbieter die Sitzung endgültig für ungültig erklärt und ich den Status aktualisiere$", s.reauthentication)
	sc.Step("^ich \"Erneut verbinden\" wähle$", s.start)
	sc.Step("^im Browser eine verspätete Anmeldung-läuft-Antwort nach \"([^\"]*)\" eintrifft$", s.lateBrowserStatus)
}

func (s *Suite) registerThen(sc *godog.ScenarioContext) {
	sc.Step("^sehe ich den Anbieterlink und den Gerätecode$", s.challenge)
	sc.Step("^sehe ich einen neuen Anbieterlink und Gerätecode$", s.challenge)
	sc.Step("^ich sehe den Status \"([^\"]*)\"$", s.statusText)
	sc.Step("^sehe ich weiterhin \"([^\"]*)\"$", s.statusText)
	sc.Step("^sehe ich \"([^\"]*)\"$", s.statusText)
	sc.Step("^Anbieterlink und Gerätecode werden nicht mehr angezeigt$", s.noChallenge)
	sc.Step("^ich sehe keinen Gerätecode$", s.noChallenge)
	sc.Step("^die Verbindung wird nicht als einsatzbereit angezeigt$", s.notReady)
	sc.Step("^ich sehe einen Hinweis zur fehlenden Anbieterfreigabe$", s.disabledNotice)
	sc.Step("^sehe ich einen Hinweis zur fehlenden Anbieterfreigabe$", s.disabledNotice)
	sc.Step("^die Seite zeigt keine Zugangstokens oder Konto-ID$", s.noSecrets)
	sc.Step("^bleibt im Browser der Status \"([^\"]*)\" ohne Gerätecode sichtbar$", s.browserStatusStays)
}

func (s *Suite) idle() error        { return s.show("idle") }
func (s *Suite) pending() error     { s.issuerStarts = 1; return s.show("pending") }
func (s *Suite) unavailable() error { return s.show("unavailable") }
func (s *Suite) connected() error   { s.connectionLive = true; return s.show("connected") }
func (s *Suite) refresh() error     { return s.show(s.state) }
func (s *Suite) confirm() error     { s.connectionLive = true; return s.show("connected") }
func (s *Suite) cancel() error      { s.connectionLive = false; return s.show("cancelled") }
func (s *Suite) reauthentication() error {
	s.connectionLive = false
	return s.show("reauthentication-required")
}

func (s *Suite) start() error {
	if s.state == "unavailable" {
		return s.show("unavailable")
	}
	s.issuerStarts++
	return s.show("pending")
}

func (s *Suite) failure(outcome string) error {
	if outcome == "abgelaufen" {
		return s.show("expired")
	}
	if outcome == "verweigert" {
		return s.show("denied")
	}
	return fmt.Errorf("unbekanntes Anbieterergebnis")
}

func (s *Suite) show(state string) error {
	data, err := s.bridge.LoadFixture("codex-verbindung-" + state)
	if err != nil {
		return err
	}
	var output bytes.Buffer
	if err := s.bridge.Render(&output, "modelle/codex/page", data); err != nil {
		return err
	}
	s.state = state
	s.html = output.String()
	return nil
}

func (s *Suite) challenge() error {
	if !s.visible("codex-verification") || !s.visible("codex-verification-link") || !s.visible("codex-user-code") {
		return fmt.Errorf("Anbieterlink oder Gerätecode nicht sichtbar")
	}
	return nil
}

func (s *Suite) noChallenge() error {
	if s.visible("codex-verification") {
		return fmt.Errorf("Gerätecode nach Abschluss sichtbar")
	}
	return nil
}

func (s *Suite) statusText(expected string) error {
	pattern := regexp.MustCompile("(?s)<div id=\"codex-status\"[^>]*>(.*?)</div>")
	match := pattern.FindStringSubmatch(s.html)
	if !s.visible("codex-status") || len(match) != 2 || !strings.Contains(html.UnescapeString(match[1]), expected) {
		return fmt.Errorf("Verbindungsstatus nicht sichtbar")
	}
	return nil
}

func (s *Suite) notReady() error {
	if s.state == "connected" {
		return fmt.Errorf("Verbindung fälschlich einsatzbereit")
	}
	return nil
}

func (s *Suite) disabledNotice() error {
	if !s.visible("codex-notice") {
		return fmt.Errorf("Hinweis zur Anbieterfreigabe fehlt")
	}
	return nil
}

func (s *Suite) noSecrets() error {
	for _, term := range []string{"access_token", "refresh_token", "account-test", "synthetic-refresh-private"} {
		if strings.Contains(strings.ToLower(s.html), term) {
			return fmt.Errorf("Anmeldegeheimnis oder Konto-ID im DOM")
		}
	}
	return nil
}

func (s *Suite) visible(id string) bool {
	pattern := regexp.MustCompile("<[^>]+id=\"" + regexp.QuoteMeta(id) + "\"[^>]*>")
	tag := pattern.FindString(s.html)
	return tag != "" && !strings.Contains(tag, " hidden") && !strings.Contains(tag, "aria-hidden=\"true\"")
}
