package modellzugang

import "github.com/cucumber/godog"

func (s *Suite) registerLiveSetup(sc *godog.ScenarioContext) {
	sc.Step(`^ein freigegebener echter ChatGPT-/Codex-Abo-Testzugang ist verbunden$`, s.liveAccount)
	sc.Step(`^der Eino-Agent "([^"]*)" darf die registrierte Fachaktion "([^"]*)" für "([^"]*)" aufrufen$`, s.livePermission)
	sc.Step(`^die Fachaktion liefert für "([^"]*)" den Status "([^"]*)" und die einmalige Prüfkennung "([^"]*)"$`, s.liveFixture)
	sc.Step(`^ich "([^"]*)" über die öffentliche Modellprüfgrenze nach Status und Prüfkennung von "([^"]*)" frage$`, s.liveQuestion)
	sc.Step(`^"([^"]*)" hat im selben Gespräch den Status von "([^"]*)" über die registrierte Fachaktion beantwortet$`, s.priorAnswer)
	sc.Step(`^ich "([^"]*)" frage "([^"]*)"$`, s.followup)
}

func (s *Suite) registerLiveAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^trifft mindestens ein Antwortteil vor dem Abschluss des Modellstreams ein$`, s.streamed)
	sc.Step(`^Eino empfängt vom Abo-Modell einen Aufruf der registrierten Fachaktion "([^"]*)" für "([^"]*)"$`, s.toolReceived)
	sc.Step(`^die Go-Anwendung prüft Aktionskennung, Parameter und Berechtigung vor dem Statuszugriff$`, s.actionChecked)
	sc.Step(`^genau ein erlaubter Statuszugriff liefert "([^"]*)" und "([^"]*)"$`, s.allowedResult)
	sc.Step(`^Eino übergibt dieses Toolergebnis in einem weiteren Abo-Modellaufruf an das Modell$`, s.toolContinued)
	sc.Step(`^die abschließende Modellantwort nennt "([^"]*)" und "([^"]*)"$`, s.finalAnswer)
	sc.Step(`^der Prüfbericht weist Eino als Ausführungsart, Codex-Abo als Modellzugang und die echten Anbieteraufrufe aus$`, s.liveReport)
	sc.Step(`^folgt ein neuer echter Abo-Modellaufruf über Eino$`, s.furtherCall)
	sc.Step(`^die Antwort nennt "([^"]*)" mit Bezug auf "([^"]*)"$`, s.followupAnswer)
	sc.Step(`^der Prüfbericht ordnet beide Fragen demselben Gespräch und getrennten Anbieteraufrufen zu$`, s.sameConversation)
}

func (s *Suite) registerLiveEdges(sc *godog.ScenarioContext) {
	sc.Step(`^"([^"]*)" empfängt über Eino einen laufenden Antwortstream vom Abo-Modell$`, s.pendingStream)
	sc.Step(`^der freigegebene Fachschritt "([^"]*)" ist noch nicht ausgeführt$`, s.pendingTool)
	sc.Step(`^ich den Modellaufruf über die öffentliche Modellprüfgrenze abbreche$`, s.cancelStream)
	sc.Step(`^zeigt der Prüfbericht den tatsächlichen Abbruchzustand und bereits empfangene Antwortteile$`, s.cancelReport)
	sc.Step(`^der ausstehende Fachschritt wird nicht ausgeführt$`, s.noToolAfterCancel)
	sc.Step(`^nach dem Abbruch startet weder ein weiterer Modellaufruf noch ein neuer Lauf ohne neuen Auftrag$`, s.noRestart)
	s.registerLimitAndNegative(sc)
}

func (s *Suite) registerLimitAndNegative(sc *godog.ScenarioContext) {
	sc.Step(`^für den freigegebenen Abo-Testzugang ist ein erreichter Anbieterlimitzustand oder ein vom Anbieter kontrolliertes Limit-Testfenster vorab belegt$`, s.limitProof)
	sc.Step(`^der Prüfbericht enthält diesen Nachweis vor dem Modellaufruf$`, s.proofRecorded)
	sc.Step(`^ich über Eino einen echten Abo-Modellaufruf starte$`, s.limitCall)
	sc.Step(`^lehnt der Anbieter den Modellaufruf wegen seines Limits ab$`, s.providerLimited)
	sc.Step(`^der Prüfbericht zeigt den Limitgrund und den abgelehnten Lauf$`, s.limitReport)
	sc.Step(`^es erfolgt kein Modellaufruf über einen separat abgerechneten API-Zugang$`, s.noAPIFallback)
	sc.Step(`^die Eino-POC-Abnahme richtet sich nach dem vollständigen Abo-Nachweis$`, s.pocGateMatches)
	sc.Step(`^für das Abo liegt nur ein erfolgreicher Codex-CLI- oder App-Server-Agentenlauf vor$`, s.codexOnly)
	sc.Step(`^ich den Nachweis für den eigenen Eino-Chatmodell-Adapter prüfe$`, s.checkGate)
	sc.Step(`^bleibt das technische Gate "([^"]*)" offen$`, s.gateStillOpen)
	sc.Step(`^der Codex-Agentenlauf wird nicht als Eino-Modellantwort oder Toolergebnisrunde gezählt$`, s.noAgentSubstitute)
}
