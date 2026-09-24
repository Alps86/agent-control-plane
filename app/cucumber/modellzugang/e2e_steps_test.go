package modellzugang

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"agentcontrolplane/app/internal/app/modellpruefung"
	"agentcontrolplane/app/internal/app/modellverbindung"
)

func (s *httpE2ESuite) completeDeviceLogin() error {
	if err := s.httpRequest(context.Background(), http.MethodPost, "/settings/modelle/codex/device/start", nil); err != nil {
		return err
	}
	if s.lastStatus != http.StatusOK || !strings.Contains(s.lastBody, `"state":"pending"`) {
		return fmt.Errorf("lokaler Gerätecode-Start fehlgeschlagen")
	}
	if err := s.httpRequest(context.Background(), http.MethodGet, "/settings/modelle/codex/device/status", nil); err != nil {
		return err
	}
	return s.connectedState("connected")
}

func (s *httpE2ESuite) connectedState(want string) error {
	var status modellverbindung.Status
	if s.lastStatus != http.StatusOK || json.Unmarshal([]byte(s.lastBody), &status) != nil || status.State != want {
		return fmt.Errorf("öffentlicher Verbindungsstatus ist nicht %q", want)
	}
	return nil
}

func (s *httpE2ESuite) httpRequest(ctx context.Context, method, path string, body []byte) error {
	request, err := http.NewRequestWithContext(ctx, method, s.appServer.URL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Origin", s.appServer.URL)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	s.lastStatus, s.lastBody = response.StatusCode, string(data)
	return err
}

func (s *httpE2ESuite) startProbe() error {
	if err := s.httpRequest(context.Background(), http.MethodPost, "/api/settings/modelle/codex/probe", []byte("{}")); err != nil {
		return err
	}
	s.probeStatus, s.probeBody = s.lastStatus, s.lastBody
	s.captureLocalEvidence()
	return nil
}

func (s *httpE2ESuite) sameAccess() error {
	requests := s.responses.Requests()
	starts, polls, exchanges := s.issuer.Counts()
	if starts != 1 || polls != 1 || exchanges != 1 {
		return fmt.Errorf("Gerätecode-Zugang wurde nicht genau einmal hergestellt")
	}
	if len(requests) != 2 {
		return fmt.Errorf("erwartet zwei lokale Provideranfragen, erhalten %d", len(requests))
	}
	for _, request := range requests {
		if request.bearer != "Bearer "+s.issuer.accessToken || request.account != s.issuer.accountID {
			return fmt.Errorf("gespeicherter Zugang wurde nicht an die feste Modellroute übergeben")
		}
	}
	return s.encryptedCredential()
}

func (s *httpE2ESuite) encryptedCredential() error {
	data, err := os.ReadFile(s.secretPath)
	if err != nil {
		return err
	}
	if bytes.Contains(data, []byte(s.issuer.accessToken)) || bytes.Contains(data, []byte(s.issuer.refreshToken)) {
		return fmt.Errorf("Anmeldegeheimnis im Klartext gespeichert")
	}
	return nil
}

func (s *httpE2ESuite) toolRoundtrip() error {
	requests := s.responses.Requests()
	if len(requests) != 2 || !strings.Contains(requests[1].body, "function_call_output") ||
		!strings.Contains(requests[1].body, "connected") {
		return fmt.Errorf("Eino-Toolergebnisrunde fehlt")
	}
	return nil
}

func (s *httpE2ESuite) redactedResult() error {
	if s.probeStatus != http.StatusOK || !strings.Contains(s.probeBody, "redacted-request-1") ||
		!strings.Contains(s.probeBody, "redacted-request-2") || !strings.Contains(s.probeBody, `"requests":2`) ||
		!strings.Contains(s.probeBody, `"tool_state":"connected"`) ||
		!strings.Contains(s.probeBody, `"model":"local-e2e-model"`) {
		return fmt.Errorf("redigierter HTTP-Probenachweis fehlt")
	}
	return s.noProbeSecrets()
}

func (s *httpE2ESuite) echoedMetadata() error {
	metadata := s.responses.Metadata()
	if len(metadata) != 2 {
		return fmt.Errorf("Provider-Metadaten der Toolrunde fehlen")
	}
	for _, entry := range metadata {
		if entry.requestID != s.issuer.accountID || entry.model != s.issuer.accessToken {
			return fmt.Errorf("Provider sendete nicht die beabsichtigten Testwerte")
		}
	}
	return nil
}

func (s *httpE2ESuite) redactedMetadata() error {
	if s.probeStatus != http.StatusOK || !strings.Contains(s.probeBody, "redacted-request-1") ||
		!strings.Contains(s.probeBody, "redacted-request-2") ||
		!strings.Contains(s.probeBody, `"model":"local-e2e-model"`) {
		return fmt.Errorf("redigierte öffentliche Provider-Metadaten fehlen")
	}
	return s.noProbeSecrets()
}

func (s *httpE2ESuite) noProbeSecrets() error {
	if strings.Contains(s.probeBody, s.issuer.accessToken) || strings.Contains(s.probeBody, s.issuer.accountID) ||
		strings.Contains(s.probeBody, s.issuer.refreshToken) || strings.Contains(s.probeBody, "function_call_output") ||
		strings.Contains(s.probeBody, "Call the registered") || strings.Contains(s.probeBody, "Prüfe den technischen") {
		return fmt.Errorf("HTTP-Probe enthält Zugangsdaten oder Rohinhalt")
	}
	return nil
}

func (s *httpE2ESuite) redactedLimit() error {
	if s.probeStatus != http.StatusOK || !strings.Contains(s.probeBody, "limit") {
		return fmt.Errorf("redigierter Limitfehler fehlt")
	}
	return s.noProbeSecrets()
}

func (s *httpE2ESuite) oneE2ERequest() error {
	if len(s.responses.Requests()) != 1 {
		return fmt.Errorf("unerwarteter Folgeaufruf")
	}
	return nil
}

func (s *httpE2ESuite) e2eGateOpen() error {
	status := s.nachweis.Status()
	expected := 0
	if s.responses.mode == "tool" || s.responses.mode == "echo" {
		expected = 1
	}
	if !status.Offen || status.Modellbelege != expected {
		return fmt.Errorf("App-Nachweis bewertet lokalen Teilbeleg nicht als offen")
	}
	return nil
}

func (s *httpE2ESuite) captureLocalEvidence() {
	requests := s.responses.Requests()
	if len(requests) == 0 {
		return
	}
	requestID := ""
	if strings.Contains(s.probeBody, "redacted-request-1") {
		requestID = "redacted-request-1"
	}
	toolRound := len(requests) == 2 && strings.Contains(requests[1].body, "function_call_output")
	s.nachweis.ErfasseTransport(modellpruefung.Transportbeleg{Anbieter: "local-http-e2e", Modell: "local-e2e-model",
		Anfragekennung: requestID, Antwort: strings.Contains(s.probeBody, `"state":"completed"`),
		Stream: strings.Contains(s.probeBody, "event: chunk"), Toolrunde: toolRound, Folgeaufruf: toolRound})
}

func (s *httpE2ESuite) cancelProbe() error {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- s.httpRequest(ctx, http.MethodPost, "/api/settings/modelle/codex/probe", []byte("{}"))
	}()
	select {
	case <-s.responses.firstDelta:
	case <-time.After(5 * time.Second):
		cancel()
		return fmt.Errorf("erster Providerantwortteil fehlt")
	}
	cancel()
	select {
	case <-done:
		s.probeBody = s.lastBody
		s.captureLocalEvidence()
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("HTTP-Probe reagiert nicht auf Abbruch")
	}
}

func (s *httpE2ESuite) providerCancelled() error {
	select {
	case <-s.responses.cancelled:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("lokaler Providerstream wurde nicht abgebrochen")
	}
}
