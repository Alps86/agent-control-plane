package modellverbindung

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexauth"
	"agentcontrolplane/app/internal/adapter/web/modellanbieter"
	appflow "agentcontrolplane/app/internal/app/modellverbindung"
	"github.com/cucumber/godog"
)

func NewSuite(t *testing.T) *Suite { return &Suite{t: t} }

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	previousLogWriter := log.Writer()
	log.SetOutput(&s.logs)
	s.t.Cleanup(func() { log.SetOutput(previousLogWriter) })
	if err := s.setupProvider(); err != nil {
		s.t.Fatal(err)
	}

	s.registerSteps(sc)
}

func (s *Suite) setupProvider() error {
	s.issuer = &FakeIssuer{pollOutcome: "pending", issueLifetime: 3600}
	s.issuer.server = httptest.NewServer(s.issuer)
	s.t.Cleanup(s.issuer.server.Close)
	if err := s.openFreshStore(); err != nil {
		return err
	}
	provider, err := codexauth.NewDirect(codexauth.Config{Issuer: s.issuer.server.URL, ClientID: codexauth.CodexCLIClientID, HTTPClient: s.issuer.server.Client()})
	if err != nil {
		return err
	}
	s.service = appflow.New(provider, s.store)
	s.handler = modellanbieter.New(s.service).Handler()
	s.response = nil
	return nil
}

func (s *Suite) registerSteps(sc *godog.ScenarioContext) {
	s.registerGivenSteps(sc)
	s.registerWhenSteps(sc)
	s.registerThenSteps(sc)
	s.registerStoreSteps(sc)
	s.registerTokenSteps(sc)
}

func (s *Suite) registerGivenSteps(sc *godog.ScenarioContext) {
	sc.Step("^die Codex-Abo-Verbindung ist nicht eingerichtet$", s.idle)
	sc.Step("^ich habe eine Gerätecode-Anmeldung begonnen$", s.started)
	sc.Step("^die Codex-Abo-Verbindung war zuvor gültig$", s.connected)
	sc.Step("^der Anbieter hat die Gerätecode-Anmeldung deaktiviert$", s.disabled)
}

func (s *Suite) registerWhenSteps(sc *godog.ScenarioContext) {
	sc.Step("^ich die Gerätecode-Anmeldung starte$", s.start)
	sc.Step("^ich die Gerätecode-Anmeldung erneut starte$", s.start)
	sc.Step("^ich den Verbindungsstatus prüfe$", s.status)
	sc.Step("^ich den Verbindungsversuch abbreche$", s.cancel)
	sc.Step("^ich die Anmeldung beim Anbieter erfolgreich abschließe$", s.providerSuccess)
	sc.Step("^der Anbieter die Anmeldung unmittelbar vor meinem Abbruch bestätigt$", s.providerSuccess)
	sc.Step("^der Gerätecode beim Anbieter abgelaufen ist$", s.providerExpired)
	sc.Step("^der Anbieter die Anmeldung verweigert$", s.providerDenied)
	sc.Step("^der Anbieter die Sitzung endgültig widerruft$", s.providerRevoked)
	sc.Step("^der Anbieter eine erneute Anmeldung ausdrücklich verlangt$", s.providerRevoked)
}

func (s *Suite) registerThenSteps(sc *godog.ScenarioContext) {
	sc.Step("^erhalte ich die Anmeldeseite des Anbieters und einen Gerätecode$", s.challenge)
	sc.Step("^wird der Versuch als laufend angezeigt$", s.pending)
	sc.Step("^wird die Codex-Abo-Verbindung als verbunden angezeigt$", s.isConnected)
	sc.Step("^wird die Anmeldung als abgebrochen angezeigt$", s.cancelled)
	sc.Step("^wird der Gerätecode als abgelaufen angezeigt$", s.expired)
	sc.Step("^sehe ich einen Hinweis zur fehlenden Anbieterfreigabe$", s.unavailable)
	sc.Step("^wird die Anmeldung als verweigert angezeigt$", s.denied)
	sc.Step("^sehe ich, dass eine erneute Anmeldung erforderlich ist$", s.reauthentication)
	sc.Step("^die Verbindung ist noch nicht einsatzbereit$", s.notReady)
	sc.Step("^die Antwort enthält keine Anmeldegeheimnisse$", s.noSecrets)
	sc.Step("^der bestätigte Vorgang wird beim Anbieter nicht abgebrochen$", s.notCancelled)
}

func (s *Suite) idle() error     { return s.status() }
func (s *Suite) started() error  { return s.start() }
func (s *Suite) disabled() error { s.issuer.startDisabled = true; return nil }
func (s *Suite) start() error {
	return s.request(http.MethodPost, "/settings/modelle/codex/device/start")
}
func (s *Suite) status() error {
	return s.request(http.MethodGet, "/settings/modelle/codex/device/status")
}
func (s *Suite) cancel() error {
	return s.request(http.MethodPost, "/settings/modelle/codex/device/cancel")
}

func (s *Suite) connected() error {
	s.issuer.issueLifetime = 240
	if err := s.start(); err != nil {
		return err
	}
	if err := s.providerSuccess(); err != nil {
		return err
	}
	return s.status()
}

func (s *Suite) request(method, path string) error {
	recorder := httptest.NewRecorder()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	s.handler.ServeHTTP(recorder, httptest.NewRequestWithContext(ctx, method, path, nil))
	result := &httptestResponse{code: recorder.Code, body: recorder.Body.String()}
	s.response = result
	if result.code != http.StatusOK {
		return fmt.Errorf("HTTP %s %s: Status %d", method, path, result.code)
	}
	return json.Unmarshal(recorder.Body.Bytes(), &result.data)
}

func (s *Suite) providerSuccess() error { s.issuer.pollOutcome = "connected"; return nil }
func (s *Suite) providerExpired() error { s.issuer.pollOutcome = "expired"; return nil }
func (s *Suite) providerDenied() error  { s.issuer.pollOutcome = "denied"; return nil }
func (s *Suite) providerRevoked() error { s.issuer.refreshOutcome = "invalid_grant"; return nil }

func (s *Suite) challenge() error {
	if err := s.expectState("pending", ""); err != nil {
		return err
	}
	if s.response.data.VerificationURL == "" || s.response.data.UserCode != "ABCD-EFGH" {
		return fmt.Errorf("Anmeldeseite oder Gerätecode fehlt")
	}
	return nil
}

func (s *Suite) pending() error     { return s.expectState("pending", "") }
func (s *Suite) isConnected() error { return s.expectState("connected", "") }
func (s *Suite) cancelled() error   { return s.expectState("cancelled", "") }
func (s *Suite) expired() error     { return s.expectState("expired", "code_expired") }
func (s *Suite) unavailable() error { return s.expectState("unavailable", "device_code_disabled") }
func (s *Suite) denied() error      { return s.expectState("denied", "authorization_denied") }
func (s *Suite) reauthentication() error {
	return s.expectState("reauthentication_required", "session_invalid")
}
func (s *Suite) notCancelled() error { return s.expectState("connected", "") }

func (s *Suite) notReady() error {
	if s.response == nil {
		return fmt.Errorf("HTTP-Antwort fehlt")
	}
	if s.response.data.State == "connected" {
		return fmt.Errorf("Verbindung ist vor erfolgreicher Anmeldung einsatzbereit")
	}
	return nil
}

func (s *Suite) noSecrets() error {
	body := strings.ToLower(s.response.body)
	for _, forbidden := range []string{"synthetic-refresh", "access_token", "refresh_token", "bearer ", "abcd-efgh", mockAccountID, s.issuer.issuedToken, s.issuer.rotatedToken} {
		if forbidden == "" {
			continue
		}
		if strings.Contains(body, forbidden) {
			return fmt.Errorf("HTTP-Antwort enthält Anmeldegeheimnis")
		}
	}
	return nil
}

func (s *Suite) expectState(state, reason string) error {
	if s.response == nil {
		return fmt.Errorf("HTTP-Antwort fehlt")
	}
	if s.response.data.State != state || s.response.data.Reason != reason {
		return fmt.Errorf("Verbindungsstatus: %q/%q statt %q/%q", s.response.data.State, s.response.data.Reason, state, reason)
	}
	return nil
}
