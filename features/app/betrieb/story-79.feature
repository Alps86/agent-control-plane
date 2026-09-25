# language: de
@story-79 @ops-02
Funktionalität: Geheimnisreferenzen und Modellverbindungsstatus sicher bereitstellen
  Als Betreiber
  möchte ich die tatsächlich nutzbare Verbindung und ihre Referenz erkennen,
  damit ein verlorener Zugang keinen weiteren Anbieteraufruf und keinen stillen Routenwechsel auslöst.
  Alle Zugangsdaten in diesen Szenarien sind synthetisch; Anbieteraufrufe gehen nur an einen kontrollierten Endpunkt.

  Grundlage:
    Angenommen ein lokaler Server mit temporärer SQLite-Datei und geschütztem Credential-Speicher läuft
    Und die Modellanbieter verwenden kontrollierte Testendpunkte

  Szenario: Beide Verbindungen haben eine gemeinsame redigierte Statusprojektion
    Angenommen Codex-Abo wurde über den kontrollierten Gerätecode-Ablauf verbunden
    Und die zentrale OpenRouter-Verbindung wurde mit einem synthetischen Schlüssel eingerichtet und geprüft
    Wenn ich die öffentlichen Verbindungsstatus über HTTP abrufe
    Dann sehe ich Codex-Abo und OpenRouter mit jeweils eigener Verbindungsreferenz und verständlichem Status
    Und Codex-Abo nennt Gerätecode als Authentifizierungsart
    Und OpenRouter nennt API-Schlüssel als Authentifizierungsart
    Und keine öffentliche Statusantwort enthält Token, Schlüssel oder Refresh-Token

  Szenario: Gespeicherte Referenzen bleiben nach Neustart sichtbar
    Angenommen Codex-Abo wurde über den kontrollierten Gerätecode-Ablauf verbunden
    Und die zentrale OpenRouter-Verbindung wurde mit einem synthetischen Schlüssel eingerichtet und geprüft
    Wenn ich den Server mit derselben Datenbank und demselben geschützten Credential-Speicher neu starte
    Und ich die öffentlichen Verbindungsstatus über HTTP abrufe
    Dann haben beide Verbindungen dieselben Referenzen wie vor dem Neustart
    Und ihre Status entsprechen den serverseitig vorhandenen Zugängen
    Und keine öffentliche Statusantwort enthält Token, Schlüssel oder Refresh-Token

  Szenario: Getrennte Verbindung verliert den einsatzbereiten Status
    Angenommen die zentrale OpenRouter-Verbindung ist für "Nord" und den Eino-Agenten "Mira" freigegeben
    Wenn ich OpenRouter über die öffentliche Settings-Grenze trenne
    Und ich die öffentlichen Verbindungsstatus über HTTP abrufe
    Dann ist OpenRouter als getrennt und nicht einsatzbereit sichtbar
    Und die bisherige OpenRouter-Referenz ist nicht mehr als nutzbare Verbindung ausgewiesen
    Und Miras öffentliche Modellwahl liefert keinen geheimen Ersatzwert
    Wenn ich den Server mit derselben Datenbank und demselben geschützten Credential-Speicher neu starte
    Dann bleibt OpenRouter als getrennt und nicht einsatzbereit sichtbar

  Szenario: Widerrufene Codex-Sitzung verliert den einsatzbereiten Status
    Angenommen der kontrollierte Codex-Anbieter liefert zunächst eine kurzlebige Sitzung
    Angenommen Codex-Abo wurde über den kontrollierten Gerätecode-Ablauf verbunden
    Und der kontrollierte Codex-Anbieter widerruft die erneuerbare Sitzung endgültig
    Wenn ich den Codex-Status über die öffentliche Settings-Grenze aktualisiere
    Und ich die öffentlichen Verbindungsstatus über HTTP abrufe
    Dann verlangt Codex-Abo eine erneute Anmeldung statt einen einsatzbereiten Status zu zeigen
    Und die öffentliche Fehler- und Statusantwort enthält weder Token noch Refresh-Token

  Szenario: Verlust der gewählten Verbindung sperrt den nächsten Modellaufruf vor dem Anbieter
    Angenommen "Mira" in "Nord" verwendet eine geprüfte und freigegebene OpenRouter-Verbindung
    Wenn "Mira" über die exportierte Modellzugangsgrenze den kontrollierten OpenRouter-Anbieter aufruft
    Dann erhält der kontrollierte OpenRouter-Anbieter genau einen Request
    Wenn ich OpenRouter über die öffentliche Settings-Grenze trenne
    Und "Mira" über dieselbe exportierte Modellzugangsgrenze erneut aufruft
    Dann wird der neue Aufruf mit einem verständlichen Verbindungsfehler abgelehnt
    Und der kontrollierte OpenRouter-Anbieter erhält keinen weiteren Request
    Und der Verbindungsfehler enthält weder Token noch Schlüssel

  Szenario: Später registrierter Anbieter zeigt ausschließlich seine konfigurierte Authentifizierungsart
    Angenommen ein weiterer kontrollierter Anbieter ist mit der Authentifizierungsart "API-Schlüssel" registriert
    Wenn ich die öffentlichen Verbindungsstatus über HTTP abrufe
    Dann zeigt dieser Anbieter nur "API-Schlüssel" als Authentifizierungsart
    Und er zeigt weder Gerätecode-Anmeldung noch Abo-Anmeldung als verfügbare Aktion
    Und keine öffentliche Antwort enthält den synthetischen Schlüssel
