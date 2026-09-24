package openrouterverbindung

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) registerRaceSteps(sc *godog.ScenarioContext) {
	sc.Step(`^der kontrollierte Testanbieter hält die nächste Statusprüfung zurück$`, s.blockNextProbe)
	sc.Step(`^ich die Statusprüfung über die öffentliche Settings-Grenze nebenläufig starte$`, s.startAsyncCheck)
	sc.Step(`^der Testanbieter den bisherigen Schlüssel empfängt und die Antwort zurückhält$`, s.awaitBlockedProbe)
	sc.Step(`^ich das Ersetzen des Schlüssels über die öffentliche Settings-Grenze nebenläufig starte$`, s.startAsyncReplacement)
	sc.Step(`^ich das Trennen der Verbindung über die öffentliche Settings-Grenze nebenläufig starte$`, s.startAsyncDisconnect)
	sc.Step(`^der Schlüsseltausch bleibt an der öffentlichen HTTP-Grenze bis zur Anbieterantwort offen$`, s.mutationPending)
	sc.Step(`^die Trennung bleibt an der öffentlichen HTTP-Grenze bis zur Anbieterantwort offen$`, s.mutationPending)
	sc.Step(`^ich die zurückgehaltene Anbieterantwort freigebe$`, s.releaseProbe)
	sc.Step(`^schließen Statusprüfung und Schlüsseltausch ohne Fehler ab$`, s.awaitBoth)
	sc.Step(`^schließen Statusprüfung und Trennung ohne Fehler ab$`, s.awaitBoth)
	sc.Step(`^die öffentliche Settings-Antwort zeigt dieselbe Verbindung mit Status "nicht geprüft"$`, s.replacementFinalStatus)
	sc.Step(`^ich den Verbindungsstatus erneut ausdrücklich prüfe$`, s.checkConnection)
	sc.Step(`^erreicht ausschließlich der neue Schlüssel den Testanbieter$`, s.newKeyOnly)
	sc.Step(`^die öffentliche Settings-Antwort zeigt "nicht eingerichtet"$`, s.disconnectedFinalStatus)
	sc.Step(`^erhält der kontrollierte Testanbieter nach der Trennung keine zusätzliche Anfrage$`, s.noFurtherCall)
}

func (s *Suite) blockNextProbe() error {
	s.probeOnce = sync.Once{}
	s.releaseOnce = sync.Once{}
	s.probeEntered = make(chan string, 1)
	s.probeRelease = make(chan struct{})
	s.checkDone = make(chan AsyncResult, 1)
	s.mutationDone = make(chan AsyncResult, 1)
	s.mutationArrived = make(chan struct{}, 1)
	s.mutationHandled = make(chan struct{}, 1)
	return nil
}

func (s *Suite) startAsyncCheck() error {
	go s.asyncRequest(http.MethodPost, apiPath+"/pruefen", nil, false, s.checkDone)
	return nil
}

func (s *Suite) awaitBlockedProbe() error {
	select {
	case header := <-s.probeEntered:
		if header != "Bearer "+validKey {
			return fmt.Errorf("blockierte Prüfung erhielt falschen Schlüssel")
		}
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("Anbieterprüfung erreichte die lokale HTTP-Grenze nicht")
	}
}

func (s *Suite) startAsyncReplacement() error {
	s.previousKey, s.currentKey = s.currentKey, replacementKey
	go s.asyncRequest(http.MethodPost, apiPath, []byte(`{"key":"`+replacementKey+`"}`), true, s.mutationDone)
	return s.awaitMutationIngress()
}

func (s *Suite) startAsyncDisconnect() error {
	go s.asyncRequest(http.MethodDelete, apiPath, nil, true, s.mutationDone)
	return s.awaitMutationIngress()
}

func (s *Suite) awaitMutationIngress() error {
	select {
	case <-s.mutationArrived:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("nebenläufige Settings-Mutation erreichte den HTTP-Handler nicht")
	}
}

func (s *Suite) mutationPending() error {
	select {
	case <-s.mutationHandled:
		return fmt.Errorf("Settings-Mutation wurde vor der ausstehenden Anbieterantwort abgeschlossen")
	case <-time.After(250 * time.Millisecond):
		return nil
	}
}

func (s *Suite) releaseProbe() error {
	s.callsAtRelease = len(s.providerCalls)
	s.releaseOnce.Do(s.closeProbe)
	return nil
}

func (s *Suite) closeProbe() {
	close(s.probeRelease)
}

func (s *Suite) awaitBoth() error {
	if err := s.awaitAsync(s.checkDone); err != nil {
		return err
	}

	return s.awaitAsync(s.mutationDone)
}

func (s *Suite) awaitAsync(done <-chan AsyncResult) error {
	select {
	case result := <-done:
		return s.expectAsync(result)
	case <-time.After(5 * time.Second):
		return fmt.Errorf("nebenläufiger HTTP-Aufruf endete nicht")
	}
}

func (s *Suite) expectAsync(result AsyncResult) error {
	if result.Err != nil || result.Status != http.StatusOK {
		return fmt.Errorf("nebenläufiger HTTP-Aufruf: Status %d, Fehler %v, Antwort %s", result.Status, result.Err, result.Body)
	}

	return nil
}

func (s *Suite) asyncRequest(method, path string, body []byte, mutation bool, done chan<- AsyncResult) {
	request, err := http.NewRequest(method, s.app.URL+path, bytes.NewReader(body))
	if err != nil {
		done <- AsyncResult{Err: err}
		return
	}

	request.Header.Set("Origin", s.app.URL)
	request.Header.Set("Content-Type", "application/json")
	if mutation {
		request.Header.Set("X-Race-Probe", "1")
	}

	done <- s.performAsync(request)
}

func (s *Suite) performAsync(request *http.Request) AsyncResult {
	response, err := s.app.Client().Do(request)
	if err != nil {
		return AsyncResult{Err: err}
	}

	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	return AsyncResult{Status: response.StatusCode, Body: body, Err: err}
}

func (s *Suite) replacementFinalStatus() error {
	if err := s.getSettings(); err != nil {
		return err
	}

	if err := s.expectSameReference(); err != nil {
		return err
	}

	return s.expectStatusText("nicht geprüft")
}

func (s *Suite) disconnectedFinalStatus() error {
	if err := s.getSettings(); err != nil {
		return err
	}

	return s.expectStatusText("nicht eingerichtet")
}

func (s *Suite) newKeyOnly() error {
	if len(s.providerCalls) != 2 || s.providerCalls[0] != "Bearer "+validKey {
		return fmt.Errorf("unerwartete Anbieteraufrufe: %v", s.providerCalls)
	}

	return s.expectLatestKey()
}

func (s *Suite) noFurtherCall() error {
	if len(s.providerCalls) != s.callsAtRelease {
		return fmt.Errorf("getrennte Verbindung löste weiteren Anbieteraufruf aus")
	}

	return nil
}
