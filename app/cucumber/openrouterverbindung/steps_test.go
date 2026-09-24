package openrouterverbindung

import "github.com/cucumber/godog"

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.cleanup)
	s.registerSetupSteps(sc)
	s.registerActionSteps(sc)
	s.registerResultSteps(sc)
	s.registerSecuritySteps(sc)
	s.registerUISteps(sc)
	s.registerRaceSteps(sc)
	s.registerRemoteSteps(sc)
}

func (s *Suite) registerSetupSteps(sc *godog.ScenarioContext) {
	sc.Step(`^der lokale Server verwendet einen neuen geschützten Credentials-Speicher$`, s.newStore)
	sc.Step(`^ein kontrollierter OpenRouter-Testanbieter steht bereit$`, s.readyMockProvider)
	sc.Step(`^noch keine OpenRouter-Verbindung eingerichtet ist$`, s.noConnection)
	sc.Step(`^eine zentrale OpenRouter-Verbindung mit (?:gültigem Schlüssel|eindeutigem Testschlüssel) eingerichtet ist$`, s.validConnection)
}

func (s *Suite) registerActionSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich als Betreiber einen gültigen OpenRouter-Schlüssel über die öffentliche Settings-Grenze speichere$`, s.saveValid)
	sc.Step(`^ich als Betreiber einen anderen gültigen OpenRouter-Schlüssel speichere$`, s.saveReplacement)
	sc.Step(`^ich als Betreiber einen leeren OpenRouter-Schlüssel über die öffentliche Settings-Grenze speichere$`, s.saveEmpty)
	sc.Step(`^ich als Betreiber einen ungültigen OpenRouter-Schlüssel speichere$`, s.saveInvalid)
	sc.Step(`^ich den Verbindungsstatus ausdrücklich prüfe$`, s.checkConnection)
	sc.Step(`^ich als Betreiber die OpenRouter-Verbindung über die öffentliche Settings-Grenze trenne$`, s.disconnect)
	sc.Step(`^ich den Verbindungsstatus erfolgreich geprüft habe$`, s.checkedReady)
	sc.Step(`^ich den lokalen Server mit demselben geschützten Credentials-Speicher neu starte$`, s.restartApplication)
	sc.Step(`^ich die öffentliche Settings-Ansicht und ihre JSON-Antwort abrufe$`, s.fetchBothViews)
	sc.Step(`^ich als Betreiber eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel einrichte$`, s.saveValid)
}

func (s *Suite) registerResultSteps(sc *godog.ScenarioContext) {
	sc.Step(`^zeigt die Settings-Antwort genau eine zentrale OpenRouter-Verbindungsreferenz$`, s.rememberReference)
	sc.Step(`^der gespeicherte Schlüssel ist in jeder öffentlichen Antwort maskiert$`, s.expectNoSecret)
	sc.Step(`^meldet die Verbindung mit dem kontrollierten Testanbieter "einsatzbereit"$`, s.readyStatus)
	sc.Step(`^die Prüfung verwendet nur den eingerichteten OpenRouter-Anbieter$`, s.expectLatestKey)
	sc.Step(`^zeigt die Settings-Antwort weiterhin genau dieselbe Verbindungsreferenz$`, s.expectSameReference)
	sc.Step(`^erreicht nur der neue Schlüssel den kontrollierten Testanbieter$`, s.expectLatestKey)
	sc.Step(`^der alte Schlüssel wird nicht erneut verwendet$`, s.expectNoOldKey)
	sc.Step(`^wird die Eingabe mit dem Hinweis "Schlüssel fehlt" abgelehnt$`, s.missingRejected)
	sc.Step(`^die OpenRouter-Verbindung ist nicht einsatzbereit$`, s.notReady)
	sc.Step(`^der kontrollierte Testanbieter erhält keine Anfrage$`, s.expectNoProviderCall)
	sc.Step(`^meldet die Verbindung "nicht einsatzbereit" mit einem verständlichen Authentifizierungsgrund$`, s.invalidStatus)
	sc.Step(`^es erfolgt keine Anfrage an einen anderen kostenpflichtigen Anbieter$`, s.onlyConfiguredProvider)
	sc.Step(`^meldet die Verbindung "nicht eingerichtet"$`, s.notConnectedStatus)
	s.registerMoreResults(sc)
}

func (s *Suite) registerMoreResults(sc *godog.ScenarioContext) {
	sc.Step(`^zeigt die Settings-Antwort "Verbindung getrennt"$`, s.disconnectedNotice)
	sc.Step(`^erhält der kontrollierte Testanbieter keine weitere Anfrage$`, s.expectNoProviderCall)
	sc.Step(`^zeigt die Settings-Antwort dieselbe zentrale OpenRouter-Verbindungsreferenz$`, s.sameAfterRestart)
	sc.Step(`^die Settings-Antwort enthält keinen Klartextschlüssel$`, s.expectNoSecret)
	sc.Step(`^enthalten HTML und JSON den eindeutigen Testschlüssel nicht$`, s.bothNoSecret)
	sc.Step(`^HTML und JSON enthalten keine teilweise entschlüsselten Schlüsselwerte$`, s.bothNoFragments)
	sc.Step(`^ich sehe nur die zentrale Verbindungsreferenz und nichtgeheime Statusdaten$`, s.bothPublicMetadata)
	sc.Step(`^enthält die öffentliche Settings-Antwort keine Organisations- oder Agentenfreigabe$`, s.expectNoGrants)
	sc.Step(`^sie zeigt nur die zentrale Verbindungsreferenz ohne Schlüssel$`, s.expectConnection)
	sc.Step(`^erhält der kontrollierte Testanbieter noch keine Anfrage$`, s.expectNoProviderCall)
	sc.Step(`^erst meine ausdrückliche Statusprüfung darf den Anbieter kontaktieren$`, s.onlyAfterCheck)
}

func (s *Suite) registerSecuritySteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich einen anderen Schlüssel mit "(fremdem Origin|ohne Origin)" an die öffentliche Settings-Grenze sende$`, s.rejectedSave)
	sc.Step(`^wird die Änderung mit HTTP 403 verweigert$`, s.forbidden)
	sc.Step(`^der kontrollierte Testanbieter erhält durch die verweigerte Änderung keine Anfrage$`, s.expectNoProviderCall)
	sc.Step(`^die zentrale Verbindung verwendet weiterhin den bisherigen Schlüssel$`, s.unchangedKey)
	sc.Step(`^ich eine Statusprüfung mit fremdem Host an die öffentliche Settings-Grenze sende$`, s.rejectedCheck)
	sc.Step(`^wird die Prüfung mit HTTP 403 verweigert$`, s.forbidden)
}
