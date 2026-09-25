package modellzugang

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestHTTPFeatures(t *testing.T) {
	suite := &httpE2ESuite{t: t}
	runner := godog.TestSuite{ScenarioInitializer: suite.InitializeScenario,
		Options: &godog.Options{Format: "pretty", Paths: []string{"../../../features/app/modelle-sprache/story-08.feature"},
			Tags: "@lokales_http_e2e", TestingT: t}}
	if runner.Run() != 0 {
		t.Fatal("Story-08-HTTP-E2E-Szenarien fehlgeschlagen")
	}
}

func (s *httpE2ESuite) InitializeScenario(sc *godog.ScenarioContext) {
	s.registerHTTPSetup(sc)
	s.registerHTTPAssertions(sc)
}

func (s *httpE2ESuite) registerHTTPSetup(sc *godog.ScenarioContext) {
	sc.Step(`^für die technische Prüfung ist ausschließlich die Abo-Modellroute konfiguriert$`, s.e2eRoute)
	sc.Step(`^die öffentliche Modellprüfgrenze protokolliert Eino-Modellaufrufe und Anbieterantworten ohne Geheimnisse$`, s.e2eProbe)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine lokale Responses-Gegenstelle sind eingerichtet$`, s.setupNormal)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine lokale Responses-Gegenstelle mit Limitfehler sind eingerichtet$`, s.setupLimit)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine offene Responses-Gegenstelle sind eingerichtet$`, s.setupCancel)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine Responses-Gegenstelle mit Zugangswerten in Provider-Metadaten sind eingerichtet$`, s.setupEcho)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine Responses-Gegenstelle mit abweichendem Modell sind eingerichtet$`, s.setupMismatch)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine Responses-Gegenstelle mit fehlendem und leerem Modell sind eingerichtet$`, s.setupMissing)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine Responses-Gegenstelle mit falscher Schlussantwort sind eingerichtet$`, s.setupWrongAnswer)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine zeitlich kontrollierte Responses-Gegenstelle sind eingerichtet$`, s.setupTimed)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine Responses-Gegenstelle ohne Abschlussstatus sind eingerichtet$`, s.setupNoStatus)
	sc.Step(`^ein lokaler Gerätecode-Anbieter und eine Responses-Gegenstelle mit ungültigem Abschlussstatus sind eingerichtet$`, s.setupBadStatus)
	sc.Step(`^ich die Gerätecode-Anmeldung und ihren Abschluss über HTTP aufrufe$`, s.completeDeviceLogin)
	sc.Step(`^die Gerätecode-Anmeldung ist über HTTP abgeschlossen$`, s.completeDeviceLogin)
	sc.Step(`^meldet der öffentliche Verbindungsstatus "([^"]*)"$`, s.connectedState)
	sc.Step(`^ich die öffentliche Eino-Probe über HTTP starte$`, s.startProbe)
	sc.Step(`^ich die laufende öffentliche Eino-Probe nach dem ersten Antwortteil abbreche$`, s.cancelProbe)
	sc.Step(`^ich den öffentlichen redigierten Provider-Fortschritt vor dem Providerabschluss empfange und diesen dann freigebe$`, s.timedProbe)
}

func (s *httpE2ESuite) registerHTTPAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^erreichen derselbe gespeicherte Zugang und das Konto die feste Responses-Route$`, s.sameAccess)
	sc.Step(`^die Gegenstelle erhält eine Toolrunde mit redigiertem Verbindungszustand$`, s.toolRoundtrip)
	sc.Step(`^die HTTP-Probe meldet zwei Anbieteraufrufe und ein redigiertes Ergebnis$`, s.redactedResult)
	sc.Step(`^hat die Gegenstelle Konto und Token als Provider-Metadaten gesendet$`, s.echoedMetadata)
	sc.Step(`^die öffentliche SSE nennt nur redigierte Request-IDs und das konfigurierte Modell$`, s.redactedMetadata)
	sc.Step(`^beide Providerbeobachtungen bestätigen das konfigurierte Modell$`, s.providerModelMatches)
	sc.Step(`^(?:beide Providerbeobachtungen melden|melden beide Providerbeobachtungen) eine Modellabweichung$`, s.providerModelDiffers)
	sc.Step(`^der Abschluss bestätigt Kennung und Antwortteil vor dem Providerabschluss$`, s.completeAnswerEvidence)
	sc.Step(`^bestätigt der Abschluss Kennung und Antwortteil vor dem Providerabschluss$`, s.completeAnswerEvidence)
	sc.Step(`^meldet der Abschluss eine fehlende Prüfkennung trotz Toolrunde$`, s.wrongAnswerEvidence)
	sc.Step(`^meldet der Abschluss trotz Antwortteil keinen gültigen Providerabschluss$`, s.missingStatusEvidence)
	sc.Step(`^meldet der Fehler keinen gültigen Providerabschluss$`, s.badStatusEvidence)
	sc.Step(`^meldet sie den Anbieterlimitfehler ohne Zugangsdaten oder Rohprompt$`, s.redactedLimit)
	sc.Step(`^die Gegenstelle erhält nur einen Modellaufruf$`, s.oneE2ERequest)
	sc.Step(`^endet der lokale Providerstream vor einem Abschlussereignis$`, s.providerCancelled)
	sc.Step(`^die Gegenstelle erhält keinen Folgeaufruf$`, s.oneE2ERequest)
	sc.Step(`^das echte Abo-Gate bleibt offen$`, s.e2eGateOpen)
}

func (s *httpE2ESuite) setupNormal() error      { return s.setup("tool") }
func (s *httpE2ESuite) setupLimit() error       { return s.setup("limit") }
func (s *httpE2ESuite) setupCancel() error      { return s.setup("cancel") }
func (s *httpE2ESuite) setupEcho() error        { return s.setup("echo") }
func (s *httpE2ESuite) setupMismatch() error    { return s.setup("mismatch") }
func (s *httpE2ESuite) setupMissing() error     { return s.setup("missing") }
func (s *httpE2ESuite) setupWrongAnswer() error { return s.setup("wrong_answer") }
func (s *httpE2ESuite) setupTimed() error       { return s.setup("timed") }
func (s *httpE2ESuite) setupNoStatus() error    { return s.setup("no_status") }
func (s *httpE2ESuite) setupBadStatus() error   { return s.setup("bad_status") }
func (s *httpE2ESuite) e2eRoute() error {
	s.routeOnly = true
	return nil
}
func (s *httpE2ESuite) e2eProbe() error {
	s.probeRequired = true
	return nil
}
