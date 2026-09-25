package modellzugang

import (
	"errors"
	"fmt"
	"strings"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"github.com/cucumber/godog"
)

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	s.registerCommon(sc)
	s.registerOffline(sc)
	s.registerOfflineExtended(sc)
	s.registerLiveSetup(sc)
	s.registerLiveAssertions(sc)
	s.registerLiveEdges(sc)
}

func (s *Suite) registerCommon(sc *godog.ScenarioContext) {
	sc.Step(`^für die technische Prüfung ist ausschließlich die Abo-Modellroute konfiguriert$`, s.aboRoute)
	sc.Step(`^die öffentliche Modellprüfgrenze protokolliert Eino-Modellaufrufe und Anbieterantworten ohne Geheimnisse$`, s.publicProbe)
}

func (s *Suite) registerOffline(sc *godog.ScenarioContext) {
	sc.Step(`^die lokale Modellgegenstelle sendet einen Stream mit der Fachaktion "([^"]*)" für "([^"]*)"$`, s.localTool)
	sc.Step(`^"([^"]*)" darf den Status von "([^"]*)" lesen$`, s.allowStatus)
	sc.Step(`^Eino den lokalen Stream und die Toolergebnisrunde verarbeitet$`, s.runLocal)
	sc.Step(`^erhält die lokale Gegenstelle zwei getrennte Modellanfragen$`, s.twoRequests)
	sc.Step(`^genau ein berechtigter Statuszugriff liefert "([^"]*)"$`, s.oneAccess)
	sc.Step(`^die lokale Antwort nennt "([^"]*)"$`, s.localAnswer)
	sc.Step(`^die echte Abo-Abnahme bleibt offen$`, s.liveGateOpen)
	sc.Step(`^die lokale Modellgegenstelle antwortet mit HTTP 429$`, s.localLimit)
	sc.Step(`^Eino den lokalen Modellaufruf startet$`, s.runLocal)
	sc.Step(`^erhält der Aufrufer einen Limitfehler des Adapters$`, s.limitError)
	sc.Step(`^die lokale Gegenstelle erhält genau eine Modellanfrage$`, s.oneRequest)
	sc.Step(`^die echte Anbieterlimit-Abnahme bleibt offen$`, s.liveGateOpen)
}

func (s *Suite) registerOfflineExtended(sc *godog.ScenarioContext) {
	sc.Step(`^die lokale Modellgegenstelle hält den SSE-Stream nach einem Antwortteil offen$`, s.localOpenStream)
	sc.Step(`^ich den Eino-Chatmodellstream nach diesem Antwortteil abbreche$`, s.runCancelLocal)
	sc.Step(`^bleibt der empfangene Antwortteil sichtbar$`, s.partialVisible)
	sc.Step(`^der HTTP-Stream wird vor einem Abschlussereignis geschlossen$`, s.closedBeforeComplete)
	sc.Step(`^keine Fachaktion wird ausgeführt$`, s.noLocalTool)
	sc.Step(`^die lokale Modellgegenstelle sendet einen SSE-Limitfehler$`, s.localSSELimit)
	sc.Step(`^ich den Eino-Chatmodellaufruf über die lokale Gegenstelle starte$`, s.runRawLocal)
	sc.Step(`^die redigierte Anbieterbeobachtung meldet den Fehler$`, s.localErrorObservation)
	sc.Step(`^die lokale Modellgegenstelle schließt eine Antwort ohne Nutzungszahlen ab$`, s.localMissingUsage)
	sc.Step(`^bleiben Nutzungszahlen in Modellantwort und Anbieterbeobachtung unbekannt$`, s.usageUnknown)
	sc.Step(`^enthält die redigierte Anbieterbeobachtung zwei verschiedene Request-IDs in Reihenfolge$`, s.offlineObservations)
	sc.Step(`^die zweite Modellanfrage enthält das Fachtoolergebnis$`, s.offlineToolOutput)
	s.registerOfflineGateAndReport(sc)
}

func (s *Suite) registerOfflineGateAndReport(sc *godog.ScenarioContext) {
	sc.Step(`^der lokale App-Nachweis erhält einen unvollständigen Modellbeleg$`, s.partialGateEvidence)
	sc.Step(`^bleibt sein technisches Gate offen$`, s.partialGateOpen)
	sc.Step(`^der lokale App-Nachweis die fehlenden Teilbelege erhält$`, s.completeGateEvidence)
	sc.Step(`^schließt sich nur das lokale technische Gate$`, s.localGateClosed)
	sc.Step(`^ein lokaler Prüfbericht enthält Modell- und Request-Metadaten sowie Testgeheimnisse$`, s.localReportSetup)
	sc.Step(`^ich ihn in eine temporäre Datei schreibe$`, s.localReportWrite)
	sc.Step(`^enthält die gespeicherte Datei Modell und Request-ID, aber weder Token noch Account-ID noch Prompt$`, s.localReportRedacted)
	sc.Step(`^die Datei hat nur Besitzerrechte$`, s.localReportMode)
	sc.Step(`^eine reguläre lokale Berichtsdatei mit altem Inhalt besteht$`, s.localExistingReport)
	sc.Step(`^ersetzt der vollständige Bericht den alten Inhalt mit Besitzerrechten$`, s.localReplacedReport)
	sc.Step(`^eine lokale Berichts-Zieldatei ist ein Symlink auf eine andere Datei$`, s.localSymlinkReport)
	sc.Step(`^ich den redigierten Bericht zu speichern versuche$`, s.localRejectedReport)
	sc.Step(`^wird der Symlink abgelehnt und sein Ziel bleibt unverändert$`, s.localSymlinkUnchanged)
	sc.Step(`^bleibt keine temporäre Berichtsdatei zurück$`, s.localNoReportTemp)
}

func (s *Suite) aboRoute() error {
	s.provider = "codexabo"
	return nil
}

func (s *Suite) publicProbe() error {
	if s.provider != "codexabo" {
		return fmt.Errorf("keine Abo-Modellroute")
	}

	return nil
}

func (s *Suite) localTool(aktion, projekt string) error {
	if aktion != "Projektstatus lesen" || projekt != "Website" {
		return fmt.Errorf("unerwartete Fachaktion oder Projekt")
	}

	return s.prepareLocal("tool")
}

func (s *Suite) allowStatus(agent, projekt string) error {
	s.pruefung.Erlaube(agent, projekt)
	return nil
}

func (s *Suite) twoRequests() error {
	if len(s.requestBodies()) != 2 || !s.hasToolResult() {
		return fmt.Errorf("Toolergebnisrunde fehlt: %d Modellanfragen", len(s.requestBodies()))
	}

	return nil
}

func (s *Suite) oneAccess(marker string) error {
	if s.fixture.Accesses() != 1 || s.fixture.status.Pruefkennung != marker {
		return fmt.Errorf("erwarteter Statuszugriff fehlt")
	}

	accesses := s.pruefung.Aufrufe()
	if len(accesses) != 1 || !accesses[0].Berechtigt || !accesses[0].Ausgefuehrt {
		return fmt.Errorf("Fachaktion wurde nicht berechtigt ausgeführt")
	}

	return nil
}

func (s *Suite) localAnswer(marker string) error {
	if !strings.Contains(s.answer, marker) {
		return fmt.Errorf("lokale Antwort enthält Prüfkennung nicht")
	}

	return nil
}

func (s *Suite) liveGateOpen() error {
	if !s.local {
		return fmt.Errorf("keine lokale Gegenstelle")
	}

	return nil
}

func (s *Suite) localLimit() error {
	return s.prepareLocal("limit")
}

func (s *Suite) limitError() error {
	var providerErr *codexabo.ProviderError
	if !errors.As(s.modelErr, &providerErr) || providerErr.Kind != "limit" {
		return fmt.Errorf("erwarteter lokaler Limitfehler fehlt: %v", s.modelErr)
	}

	return nil
}

func (s *Suite) oneRequest() error {
	if len(s.requestBodies()) != 1 {
		return fmt.Errorf("erwartet eine Modellanfrage, erhalten %d", len(s.requestBodies()))
	}

	return nil
}

func (s *Suite) localOpenStream() error {
	return s.prepareLocal("stream_cancel")
}

func (s *Suite) partialVisible() error {
	if s.answer != "Teilantwort" || s.chunks != 1 {
		return fmt.Errorf("erster Antwortteil fehlt")
	}

	return nil
}

func (s *Suite) closedBeforeComplete() error {
	if !s.cancelled || !s.cancelBody.Closed() {
		return fmt.Errorf("SSE-Stream wurde nicht vor Abschluss geschlossen")
	}

	for _, entry := range s.observations.Entries() {
		if entry.Status == "completed" {
			return fmt.Errorf("Abschlussereignis trotz Abbruch beobachtet")
		}
	}

	return nil
}

func (s *Suite) noLocalTool() error {
	if s.fixture.Accesses() != 0 {
		return fmt.Errorf("Fachaktion trotz Abbruch ausgeführt")
	}

	return nil
}

func (s *Suite) localSSELimit() error {
	return s.prepareLocal("sse_limit")
}

func (s *Suite) localErrorObservation() error {
	entries := s.observations.Entries()
	if len(entries) != 1 || entries[0].RequestID != "offline-request-1" ||
		!strings.Contains(entries[0].Status, "failed") {
		return fmt.Errorf("redigierter SSE-Fehler fehlt")
	}

	return nil
}

func (s *Suite) localMissingUsage() error {
	return s.prepareLocal("usage")
}

func (s *Suite) usageUnknown() error {
	entries := s.observations.Entries()
	if !s.unknownUsage || len(entries) != 1 || entries[0].Usage != nil {
		return fmt.Errorf("fehlende Nutzung wurde nicht als unbekannt erhalten")
	}

	return nil
}
