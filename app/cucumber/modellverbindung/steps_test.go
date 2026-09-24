package modellverbindung

import (
	"context"
	"encoding/json"
	"errors"
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
	s.issuer = &FakeIssuer{pollOutcome: "pending", issueLifetime: 3600, interval: "0"}
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
	handler, err := modellanbieter.New(s.service, "127.0.0.1:8080")
	if err != nil {
		return err
	}

	s.handler = handler.Handler()
	s.response = nil
	return nil
}

func (s *Suite) registerSteps(sc *godog.ScenarioContext) {
	s.registerGivenSteps(sc)
	s.registerWhenSteps(sc)
	s.registerThenSteps(sc)
	s.registerStoreSteps(sc)
	s.registerTokenSteps(sc)
	s.registerRestartSteps(sc)
	s.registerTimingSteps(sc)
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
	sc.Step("^ich die Gerätecode-Anmeldung mit fremder Browser-Herkunft starte$", s.foreignOriginStart)
	sc.Step("^ich die Gerätecode-Anmeldung mit fremdem Host starte$", s.foreignHostStart)
	sc.Step("^ich den Verbindungsversuch mit fremder Browser-Herkunft abbreche$", s.foreignOriginCancel)
	sc.Step("^ich den Verbindungsversuch mit fremdem Host abbreche$", s.foreignHostCancel)
	sc.Step("^ich die Gerätecode-Anmeldung ohne Browser-Herkunft starte$", s.missingOriginStart)
	sc.Step("^ich den Verbindungsversuch ohne Browser-Herkunft abbreche$", s.missingOriginCancel)
	sc.Step("^ich den Verbindungsstatus mit fremdem Host prüfe$", s.foreignHostStatus)
	sc.Step("^ich die Gerätecode-Anmeldung von einem entfernten Peer mit lokalen Headern starte$", s.foreignPeerStart)
	sc.Step("^ich den Verbindungsversuch von einem entfernten Peer mit lokalen Headern abbreche$", s.foreignPeerCancel)
	sc.Step("^ich den Verbindungsstatus von einem entfernten Peer mit lokalen Headern prüfe$", s.foreignPeerStatus)
	sc.Step("^ich den Settings-Handler mit der Bind-Adresse \"([^\"]*)\" initialisiere$", s.configureBind)
	sc.Step("^ich die Gerätecode-Anmeldung über localhost starte$", s.localhostStart)
	sc.Step("^ich die Gerätecode-Anmeldung über 127.0.0.2 starte$", s.alternateLoopbackStart)
	sc.Step("^ich den Verbindungsstatus über 127.0.0.2 prüfe$", s.alternateLoopbackStatus)
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
	sc.Step("^wird die Settings-Anfrage untersagt$", s.forbidden)
	sc.Step("^beim Anbieter wurde kein Gerätecode angefordert$", s.noProviderStart)
	sc.Step("^der Verbindungsversuch bleibt laufend$", s.stillPending)
	sc.Step("^die Antwort enthält keinen Gerätecode$", s.noCode)
	sc.Step("^wird die unsichere Bind-Adresse abgelehnt$", s.bindRejected)
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
	if err := s.recordRequest(method, path, "127.0.0.1:8080", "http://127.0.0.1:8080"); err != nil {
		return err
	}

	if s.response.code != http.StatusOK {
		return fmt.Errorf("HTTP %s %s: Status %d", method, path, s.response.code)
	}

	return nil
}

func (s *Suite) recordRequest(method, path, host, origin string) error {
	return s.recordPeerRequest(method, path, host, origin, "127.0.0.1:54321")
}

func (s *Suite) recordPeerRequest(method, path, host, origin, peer string) error {
	recorder := httptest.NewRecorder()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request := httptest.NewRequestWithContext(ctx, method, path, nil)
	request.Host = host
	request.RemoteAddr = peer
	if origin != "" {
		request.Header.Set("Origin", origin)
	}

	s.handler.ServeHTTP(recorder, request)
	result := &httptestResponse{code: recorder.Code, body: recorder.Body.String()}
	s.response = result
	return json.Unmarshal(recorder.Body.Bytes(), &result.data)
}

func (s *Suite) foreignOriginStart() error {
	return s.recordRequest(http.MethodPost, "/settings/modelle/codex/device/start", "127.0.0.1:8080", "https://foreign.example")
}

func (s *Suite) foreignHostStart() error {
	return s.recordRequest(http.MethodPost, "/settings/modelle/codex/device/start", "foreign.example:8080", "http://foreign.example:8080")
}

func (s *Suite) foreignOriginCancel() error {
	return s.recordRequest(http.MethodPost, "/settings/modelle/codex/device/cancel", "127.0.0.1:8080", "https://foreign.example")
}

func (s *Suite) foreignHostCancel() error {
	return s.recordRequest(http.MethodPost, "/settings/modelle/codex/device/cancel", "foreign.example:8080", "http://foreign.example:8080")
}

func (s *Suite) missingOriginStart() error {
	return s.recordRequest(http.MethodPost, "/settings/modelle/codex/device/start", "127.0.0.1:8080", "")
}

func (s *Suite) missingOriginCancel() error {
	return s.recordRequest(http.MethodPost, "/settings/modelle/codex/device/cancel", "127.0.0.1:8080", "")
}

func (s *Suite) foreignHostStatus() error {
	return s.recordRequest(http.MethodGet, "/settings/modelle/codex/device/status", "foreign.example:8080", "")
}

func (s *Suite) foreignPeerStart() error {
	return s.recordPeerRequest(http.MethodPost, "/settings/modelle/codex/device/start", "127.0.0.1:8080", "http://127.0.0.1:8080", "198.51.100.77:54321")
}

func (s *Suite) foreignPeerCancel() error {
	return s.recordPeerRequest(http.MethodPost, "/settings/modelle/codex/device/cancel", "127.0.0.1:8080", "http://127.0.0.1:8080", "198.51.100.77:54321")
}

func (s *Suite) foreignPeerStatus() error {
	return s.recordPeerRequest(http.MethodGet, "/settings/modelle/codex/device/status", "127.0.0.1:8080", "", "198.51.100.77:54321")
}

func (s *Suite) configureBind(address string) error {
	handler, err := modellanbieter.New(s.service, address)
	s.bindErr = err
	s.handler = nil
	if err == nil {
		s.handler = handler.Handler()
	}

	return nil
}

func (s *Suite) localhostStart() error {
	if err := s.recordRequest(http.MethodPost, "/settings/modelle/codex/device/start", "localhost:8080", "http://localhost:8080"); err != nil {
		return err
	}

	if s.response.code != http.StatusOK {
		return fmt.Errorf("lokaler Browser erhielt HTTP %d", s.response.code)
	}

	return nil
}

func (s *Suite) alternateLoopbackStart() error {
	return s.alternateLoopbackRequest(http.MethodPost, "/settings/modelle/codex/device/start")
}

func (s *Suite) alternateLoopbackStatus() error {
	return s.alternateLoopbackRequest(http.MethodGet, "/settings/modelle/codex/device/status")
}

func (s *Suite) alternateLoopbackRequest(method, path string) error {
	if err := s.recordRequest(method, path, "127.0.0.2:8080", "http://127.0.0.2:8080"); err != nil {
		return err
	}

	if s.response.code != http.StatusOK {
		return fmt.Errorf("zweite Loopback-Adresse erhielt HTTP %d", s.response.code)
	}

	return nil
}

func (s *Suite) bindRejected() error {
	if !errors.Is(s.bindErr, modellanbieter.ErrUntrustedBind) || s.handler != nil {
		return fmt.Errorf("unsichere Bind-Adresse wurde nicht vor dem Mounten abgelehnt")
	}

	return nil
}

func (s *Suite) forbidden() error {
	if s.response.code != http.StatusForbidden || s.response.data.Reason != "request_not_allowed" {
		return fmt.Errorf("Settings-Anfrage wurde nicht untersagt: HTTP %d", s.response.code)
	}

	return nil
}

func (s *Suite) noProviderStart() error {
	s.issuer.mu.Lock()
	defer s.issuer.mu.Unlock()
	if s.issuer.starts != 0 {
		return fmt.Errorf("Anbieter erhielt trotz Abweisung eine Startanfrage")
	}

	return nil
}

func (s *Suite) stillPending() error {
	if err := s.status(); err != nil {
		return err
	}

	return s.pending()
}

func (s *Suite) noCode() error {
	if strings.Contains(s.response.body, "ABCD-EFGH") || s.response.data.UserCode != "" || s.response.data.VerificationURL != "" {
		return fmt.Errorf("Gerätecode trotz fremdem Host ausgeliefert")
	}

	return nil
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
