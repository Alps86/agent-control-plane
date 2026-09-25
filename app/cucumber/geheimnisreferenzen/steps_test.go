package geheimnisreferenzen

import "github.com/cucumber/godog"

func (s *Suite) registerSteps(sc *godog.ScenarioContext) {
	s.registerStatusSteps(sc)
	s.registerDisconnectSteps(sc)
	s.registerInvokeSteps(sc)
	s.registerLaterSteps(sc)
}

func (s *Suite) registerStatusSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ein lokaler Server mit temporärer SQLite-Datei und geschütztem Credential-Speicher läuft$`, s.start)
	sc.Step(`^die Modellanbieter verwenden kontrollierte Testendpunkte$`, s.controlled)
	sc.Step(`^Codex-Abo wurde über den kontrollierten Gerätecode-Ablauf verbunden$`, s.connectCodex)
	sc.Step(`^die zentrale OpenRouter-Verbindung wurde mit einem synthetischen Schlüssel eingerichtet und geprüft$`, s.connectRouter)
	sc.Step(`^ich die öffentlichen Verbindungsstatus über HTTP abrufe$`, s.getStatus)
	sc.Step(`^sehe ich Codex-Abo und OpenRouter mit jeweils eigener Verbindungsreferenz und verständlichem Status$`, s.bothConnected)
	sc.Step(`^Codex-Abo nennt Gerätecode als Authentifizierungsart$`, s.codexAuth)
	sc.Step(`^OpenRouter nennt API-Schlüssel als Authentifizierungsart$`, s.routerAuth)
	sc.Step(`^keine öffentliche Statusantwort enthält Token, Schlüssel oder Refresh-Token$`, s.noStatusSecrets)
	sc.Step(`^ich den Server mit derselben Datenbank und demselben geschützten Credential-Speicher neu starte$`, s.restart)
	sc.Step(`^haben beide Verbindungen dieselben Referenzen wie vor dem Neustart$`, s.sameReferences)
	sc.Step(`^ihre Status entsprechen den serverseitig vorhandenen Zugängen$`, s.bothConnected)
}

func (s *Suite) registerDisconnectSteps(sc *godog.ScenarioContext) {
	sc.Step(`^die zentrale OpenRouter-Verbindung ist für "Nord" und den Eino-Agenten "Mira" freigegeben$`, s.setupInvoke)
	sc.Step(`^ich OpenRouter über die öffentliche Settings-Grenze trenne$`, s.disconnectRouter)
	sc.Step(`^ist OpenRouter als getrennt und nicht einsatzbereit sichtbar$`, s.routerDisconnected)
	sc.Step(`^die bisherige OpenRouter-Referenz ist nicht mehr als nutzbare Verbindung ausgewiesen$`, s.noRouterReference)
	sc.Step(`^Miras öffentliche Modellwahl liefert keinen geheimen Ersatzwert$`, s.noFallbackValue)
	sc.Step(`^bleibt OpenRouter als getrennt und nicht einsatzbereit sichtbar$`, s.verifyDisconnectedAfterRestart)
	sc.Step(`^der kontrollierte Codex-Anbieter widerruft die erneuerbare Sitzung endgültig$`, s.revokeCodex)
	sc.Step(`^der kontrollierte Codex-Anbieter liefert zunächst eine kurzlebige Sitzung$`, s.shortCodex)
	sc.Step(`^ich den Codex-Status über die öffentliche Settings-Grenze aktualisiere$`, s.requestRevokedCodex)
	sc.Step(`^verlangt Codex-Abo eine erneute Anmeldung statt einen einsatzbereiten Status zu zeigen$`, s.codexReauthentication)
	sc.Step(`^die öffentliche Fehler- und Statusantwort enthält weder Token noch Refresh-Token$`, s.noStatusSecrets)
}

func (s *Suite) registerInvokeSteps(sc *godog.ScenarioContext) {
	sc.Step(`^"Mira" in "Nord" verwendet eine geprüfte und freigegebene OpenRouter-Verbindung$`, s.setupInvoke)
	sc.Step(`^"Mira" über die exportierte Modellzugangsgrenze den kontrollierten OpenRouter-Anbieter aufruft$`, s.invoke)
	sc.Step(`^erhält der kontrollierte OpenRouter-Anbieter genau einen Request$`, s.oneCall)
	sc.Step(`^"Mira" über dieselbe exportierte Modellzugangsgrenze erneut aufruft$`, s.invoke)
	sc.Step(`^wird der neue Aufruf mit einem verständlichen Verbindungsfehler abgelehnt$`, s.invokeDenied)
	sc.Step(`^der kontrollierte OpenRouter-Anbieter erhält keinen weiteren Request$`, s.oneCall)
	sc.Step(`^der Verbindungsfehler enthält weder Token noch Schlüssel$`, s.noInvokeSecrets)
}

func (s *Suite) registerLaterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ein weiterer kontrollierter Anbieter ist mit der Authentifizierungsart "API-Schlüssel" registriert$`, s.registerLaterProvider)
	sc.Step(`^zeigt dieser Anbieter nur "API-Schlüssel" als Authentifizierungsart$`, s.laterProviderAuth)
	sc.Step(`^er zeigt weder Gerätecode-Anmeldung noch Abo-Anmeldung als verfügbare Aktion$`, s.noOtherAuthAction)
	sc.Step(`^keine öffentliche Antwort enthält den synthetischen Schlüssel$`, s.noStatusSecrets)
}
