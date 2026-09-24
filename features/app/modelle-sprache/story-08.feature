# language: de
# Story-08 / MS-01: technisches Gate mit echtem, freigegebenem ChatGPT-/Codex-Abo-Testzugang.
# Die positiven Nachweise verwenden weder einen Modellstub noch einen separat abgerechneten API-Key.
# Der Prüfbericht hält Transport, verwendetes Konto, Modell, Anbieter-Request-IDs und
# Kontingent-/Limitbeobachtung ohne Zugangsdaten fest. Ohne Testzugang bleibt das Gate offen.
# Für den Limitfall ist zusätzlich ein vorab belegter, vom Anbieter erzeugter Limitzustand
# oder ein vom Anbieter kontrolliertes Testfenster nötig. Ohne diesen Nachweis bleibt
# die Limitabnahme offen; ein Stub oder künstlicher Fehler zählt nicht als Anbieterlimit.
@Story-08 @MS-01
Funktionalität: Codex-Abo als eigener Eino-Chatmodell-Adapter
  Als Betreiber
  will ich einen echten Abo-Modelltransport in Eino nachweisen,
  damit ein Eino-Agent Antworten und freigegebene Fachaktionen ohne API-Fallback ausführen kann.

  Grundlage:
    Angenommen für die technische Prüfung ist ausschließlich die Abo-Modellroute konfiguriert
    Und die öffentliche Modellprüfgrenze protokolliert Eino-Modellaufrufe und Anbieterantworten ohne Geheimnisse

  @externes_gate @echtes_codex_abo
  Szenario: Gestreamte Antwort mit registrierter Fachaktion und Modellfortsetzung
    Angenommen ein freigegebener echter ChatGPT-/Codex-Abo-Testzugang ist verbunden
    Angenommen der Eino-Agent "Mira" darf die registrierte Fachaktion "Projektstatus lesen" für "Website" aufrufen
    Und die Fachaktion liefert für "Website" den Status "in Arbeit" und die einmalige Prüfkennung "MS01-WEBSITE-4711"
    Wenn ich "Mira" über die öffentliche Modellprüfgrenze nach Status und Prüfkennung von "Website" frage
    Dann trifft mindestens ein Antwortteil vor dem Abschluss des Modellstreams ein
    Und Eino empfängt vom Abo-Modell einen Aufruf der registrierten Fachaktion "Projektstatus lesen" für "Website"
    Und die Go-Anwendung prüft Aktionskennung, Parameter und Berechtigung vor dem Statuszugriff
    Und genau ein erlaubter Statuszugriff liefert "in Arbeit" und "MS01-WEBSITE-4711"
    Und Eino übergibt dieses Toolergebnis in einem weiteren Abo-Modellaufruf an das Modell
    Und die abschließende Modellantwort nennt "in Arbeit" und "MS01-WEBSITE-4711"
    Und der Prüfbericht weist Eino als Ausführungsart, Codex-Abo als Modellzugang und die echten Anbieteraufrufe aus

  @externes_gate @echtes_codex_abo
  Szenario: Anschlussfrage löst einen weiteren Eino-Modellaufruf im selben Gespräch aus
    Angenommen ein freigegebener echter ChatGPT-/Codex-Abo-Testzugang ist verbunden
    Angenommen "Mira" hat im selben Gespräch den Status von "Website" über die registrierte Fachaktion beantwortet
    Wenn ich "Mira" frage "Welche Prüfkennung gehörte zu diesem Status?"
    Dann folgt ein neuer echter Abo-Modellaufruf über Eino
    Und die Antwort nennt "MS01-WEBSITE-4711" mit Bezug auf "Website"
    Und der Prüfbericht ordnet beide Fragen demselben Gespräch und getrennten Anbieteraufrufen zu

  @externes_gate @echtes_codex_abo
  Szenario: Laufenden Modellstream vor einem ausstehenden Fachschritt abbrechen
    Angenommen ein freigegebener echter ChatGPT-/Codex-Abo-Testzugang ist verbunden
    Angenommen "Mira" empfängt über Eino einen laufenden Antwortstream vom Abo-Modell
    Und der freigegebene Fachschritt "Projektstatus lesen" ist noch nicht ausgeführt
    Wenn ich den Modellaufruf über die öffentliche Modellprüfgrenze abbreche
    Dann zeigt der Prüfbericht den tatsächlichen Abbruchzustand und bereits empfangene Antwortteile
    Und der ausstehende Fachschritt wird nicht ausgeführt
    Und nach dem Abbruch startet weder ein weiterer Modellaufruf noch ein neuer Lauf ohne neuen Auftrag

  @externes_gate @echtes_codex_abo
  Szenario: Echtes Anbieterlimit beendet den betroffenen Lauf ohne kostenpflichtigen Ersatzaufruf
    Angenommen ein freigegebener echter ChatGPT-/Codex-Abo-Testzugang ist verbunden
    Angenommen für den freigegebenen Abo-Testzugang ist ein erreichter Anbieterlimitzustand oder ein vom Anbieter kontrolliertes Limit-Testfenster vorab belegt
    Und der Prüfbericht enthält diesen Nachweis vor dem Modellaufruf
    Wenn ich über Eino einen echten Abo-Modellaufruf starte
    Dann lehnt der Anbieter den Modellaufruf wegen seines Limits ab
    Und der Prüfbericht zeigt den Limitgrund und den abgelehnten Lauf
    Und es erfolgt kein Modellaufruf über einen separat abgerechneten API-Zugang
    Und die Eino-POC-Abnahme richtet sich nach dem vollständigen Abo-Nachweis

  @externes_gate @echtes_codex_abo
  Szenario: Ein Codex-Agentenlauf ersetzt den Eino-Chatmodellnachweis nicht
    Angenommen für das Abo liegt nur ein erfolgreicher Codex-CLI- oder App-Server-Agentenlauf vor
    Wenn ich den Nachweis für den eigenen Eino-Chatmodell-Adapter prüfe
    Dann bleibt das technische Gate "MS-01" offen
    Und der Codex-Agentenlauf wird nicht als Eino-Modellantwort oder Toolergebnisrunde gezählt

  @offline_schnittstelle
  Szenario: Lokale Gegenstelle prüft Toolergebnisrunde ohne Abo-Abnahme
    Angenommen die lokale Modellgegenstelle sendet einen Stream mit der Fachaktion "Projektstatus lesen" für "Website"
    Und "Mira" darf den Status von "Website" lesen
    Wenn Eino den lokalen Stream und die Toolergebnisrunde verarbeitet
    Dann erhält die lokale Gegenstelle zwei getrennte Modellanfragen
    Und genau ein berechtigter Statuszugriff liefert "MS01-WEBSITE-4711"
    Und die lokale Antwort nennt "MS01-WEBSITE-4711"
    Aber die echte Abo-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Lokaler Limitfehler löst keinen API-Ersatzaufruf aus
    Angenommen die lokale Modellgegenstelle antwortet mit HTTP 429
    Wenn Eino den lokalen Modellaufruf startet
    Dann erhält der Aufrufer einen Limitfehler des Adapters
    Und die lokale Gegenstelle erhält genau eine Modellanfrage
    Aber die echte Anbieterlimit-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Lokalen laufenden SSE-Stream nach einem Antwortteil abbrechen
    Angenommen die lokale Modellgegenstelle hält den SSE-Stream nach einem Antwortteil offen
    Wenn ich den Eino-Chatmodellstream nach diesem Antwortteil abbreche
    Dann bleibt der empfangene Antwortteil sichtbar
    Und der HTTP-Stream wird vor einem Abschlussereignis geschlossen
    Und keine Fachaktion wird ausgeführt
    Aber die echte Abo-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Anbieterlimit als SSE-Fehler übersetzen
    Angenommen die lokale Modellgegenstelle sendet einen SSE-Limitfehler
    Wenn ich den Eino-Chatmodellaufruf über die lokale Gegenstelle starte
    Dann erhält der Aufrufer einen Limitfehler des Adapters
    Und die redigierte Anbieterbeobachtung meldet den Fehler
    Aber die echte Anbieterlimit-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Fehlende Nutzungszahlen bleiben unbekannt
    Angenommen die lokale Modellgegenstelle schließt eine Antwort ohne Nutzungszahlen ab
    Wenn ich den Eino-Chatmodellaufruf über die lokale Gegenstelle starte
    Dann bleiben Nutzungszahlen in Modellantwort und Anbieterbeobachtung unbekannt
    Aber die echte Abo-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Zwei Modellrunden bewahren getrennte Request-IDs
    Angenommen die lokale Modellgegenstelle sendet einen Stream mit der Fachaktion "Projektstatus lesen" für "Website"
    Und "Mira" darf den Status von "Website" lesen
    Wenn Eino den lokalen Stream und die Toolergebnisrunde verarbeitet
    Dann enthält die redigierte Anbieterbeobachtung zwei verschiedene Request-IDs in Reihenfolge
    Und die zweite Modellanfrage enthält das Fachtoolergebnis
    Aber die echte Abo-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: App-Gate bewertet Teilbeleg und Vollbeleg getrennt
    Angenommen der lokale App-Nachweis erhält einen unvollständigen Modellbeleg
    Dann bleibt sein technisches Gate offen
    Wenn der lokale App-Nachweis die fehlenden Teilbelege erhält
    Dann schließt sich nur das lokale technische Gate
    Aber die echte Abo-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Lokaler Prüfbericht speichert redigierte Metadaten dauerhaft
    Angenommen ein lokaler Prüfbericht enthält Modell- und Request-Metadaten sowie Testgeheimnisse
    Wenn ich ihn in eine temporäre Datei schreibe
    Dann enthält die gespeicherte Datei Modell und Request-ID, aber weder Token noch Account-ID noch Prompt
    Und die Datei hat nur Besitzerrechte
    Aber die echte Abo-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Bestehenden regulären Bericht vollständig und atomar ersetzen
    Angenommen eine reguläre lokale Berichtsdatei mit altem Inhalt besteht
    Wenn ich ihn in eine temporäre Datei schreibe
    Dann ersetzt der vollständige Bericht den alten Inhalt mit Besitzerrechten
    Und bleibt keine temporäre Berichtsdatei zurück
    Aber die echte Abo-Abnahme bleibt offen

  @offline_schnittstelle
  Szenario: Symlink am Berichtsziel ablehnen
    Angenommen eine lokale Berichts-Zieldatei ist ein Symlink auf eine andere Datei
    Wenn ich den redigierten Bericht zu speichern versuche
    Dann wird der Symlink abgelehnt und sein Ziel bleibt unverändert
    Und bleibt keine temporäre Berichtsdatei zurück
    Aber die echte Abo-Abnahme bleibt offen
