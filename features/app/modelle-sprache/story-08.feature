# language: de
# Story-08 / MS-01: technisches Gate mit echtem, freigegebenem ChatGPT-/Codex-Abo-Testzugang.
# Die positiven Nachweise verwenden weder einen Modellstub noch einen separat abgerechneten API-Key.
# Der Prüfbericht hält Transport, verwendetes Konto, Modell, Anbieter-Request-IDs und
# Kontingent-/Limitbeobachtung ohne Zugangsdaten fest. Ohne Testzugang bleibt das Gate offen.
# Für den Limitfall ist zusätzlich ein vorab belegter, vom Anbieter erzeugter Limitzustand
# oder ein vom Anbieter kontrolliertes Testfenster nötig. Ohne diesen Nachweis bleibt
# die Limitabnahme offen; ein Stub oder künstlicher Fehler zählt nicht als Anbieterlimit.
@Story-08 @MS-01 @externes_gate @echtes_codex_abo
Funktionalität: Codex-Abo als eigener Eino-Chatmodell-Adapter
  Als Betreiber
  will ich einen echten Abo-Modelltransport in Eino nachweisen,
  damit ein Eino-Agent Antworten und freigegebene Fachaktionen ohne API-Fallback ausführen kann.

  Grundlage:
    Angenommen ein freigegebener echter ChatGPT-/Codex-Abo-Testzugang ist verbunden
    Und für die technische Prüfung ist ausschließlich die Abo-Modellroute konfiguriert
    Und die öffentliche Modellprüfgrenze protokolliert Eino-Modellaufrufe und Anbieterantworten ohne Geheimnisse

  Szenario: Gestreamte Antwort mit registrierter Fachaktion und Modellfortsetzung
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

  Szenario: Anschlussfrage löst einen weiteren Eino-Modellaufruf im selben Gespräch aus
    Angenommen "Mira" hat im selben Gespräch den Status von "Website" über die registrierte Fachaktion beantwortet
    Wenn ich "Mira" frage "Welche Prüfkennung gehörte zu diesem Status?"
    Dann folgt ein neuer echter Abo-Modellaufruf über Eino
    Und die Antwort nennt "MS01-WEBSITE-4711" mit Bezug auf "Website"
    Und der Prüfbericht ordnet beide Fragen demselben Gespräch und getrennten Anbieteraufrufen zu

  Szenario: Laufenden Modellstream vor einem ausstehenden Fachschritt abbrechen
    Angenommen "Mira" empfängt über Eino einen laufenden Antwortstream vom Abo-Modell
    Und ein angeforderter Fachschritt "Projektstatus lesen" ist noch nicht ausgeführt
    Wenn ich den Modellaufruf über die öffentliche Modellprüfgrenze abbreche
    Dann zeigt der Prüfbericht den tatsächlichen Abbruchzustand und bereits empfangene Antwortteile
    Und der ausstehende Fachschritt wird nicht ausgeführt
    Und nach dem Abbruch startet weder ein weiterer Modellaufruf noch ein neuer Lauf ohne neuen Auftrag

  Szenario: Echtes Anbieterlimit beendet den betroffenen Lauf ohne kostenpflichtigen Ersatzaufruf
    Angenommen für den freigegebenen Abo-Testzugang ist ein erreichter Anbieterlimitzustand oder ein vom Anbieter kontrolliertes Limit-Testfenster vorab belegt
    Und der Prüfbericht enthält diesen Nachweis vor dem Modellaufruf
    Wenn ich über Eino einen echten Abo-Modellaufruf starte
    Dann lehnt der Anbieter den Modellaufruf wegen seines Limits ab
    Und der Prüfbericht zeigt den Limitgrund und den abgelehnten Lauf
    Und es erfolgt kein Modellaufruf über einen separat abgerechneten API-Zugang
    Und die Eino-POC-Abnahme bleibt bis zu einem erfolgreichen Abo-Nachweis offen

  Szenario: Ein Codex-Agentenlauf ersetzt den Eino-Chatmodellnachweis nicht
    Angenommen für das Abo liegt nur ein erfolgreicher Codex-CLI- oder App-Server-Agentenlauf vor
    Wenn ich den Nachweis für den eigenen Eino-Chatmodell-Adapter prüfe
    Dann bleibt das technische Gate "MS-01" offen
    Und der Codex-Agentenlauf wird nicht als Eino-Modellantwort oder Toolergebnisrunde gezählt
