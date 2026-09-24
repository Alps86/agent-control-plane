# Abo nachweis

Quelle: `spec/features/modelle-sprache/abo-nachweis.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Codex-Abo für Eino nachweisen und per Gerätecode wieder verbinden

  @MS-01
  Szenario: Abo-Modell unter Eino mit Stream und Toolergebnisrunde prüfen
    Angenommen "Mira" ist ein Eino-Agent mit gültig verbundenem Codex-Abo
    Und "Mira" darf den Status des Projekts "Website" lesen
    Wenn ich "Mira" nach dem Status von "Website" frage
    Dann sehe ich Antwortteile während des Laufs eintreffen
    Und ich sehe den freigegebenen Statuszugriff und sein Ergebnis im Lauf
    Und "Mira" beantwortet die Frage unter Verwendung dieses Ergebnisses
    Wenn ich im selben Gespräch eine Anschlussfrage zu diesem Status sende
    Dann sehe ich eine dazu passende Antwort aus einem weiteren Eino-Modellaufruf
    Und der Lauf weist Eino als Ausführungsart und Codex-Abo als Modellzugang aus

  @MS-01
  Szenario: Ein Codex-Agentenlauf belegt den Eino-Modellzugang nicht
    Angenommen für das Codex-Abo ist nur ein erfolgreicher Codex-CLI- oder App-Server-Agentenlauf bekannt
    Wenn ich den Modellzugang für einen Eino-Agenten prüfe
    Dann wird der Eino-Abo-Modellzugang als nicht nachgewiesen angezeigt
    Und die Eino-POC-Abnahme bleibt offen

  @MS-01
  Szenario: Gestreamten Eino-Abo-Modellaufruf abbrechen
    Angenommen "Mira" beantwortet meine Frage unter Eino über das Codex-Abo als laufenden Stream
    Wenn ich den Modellaufruf abbreche
    Dann sehe ich den tatsächlichen Abbruchzustand und die bis dahin eingetroffenen Antwortteile
    Und kein noch ausstehender Eino-Toolschritt wird ausgeführt
    Und ein neuer Lauf beginnt nur durch einen neuen Auftrag

  @MS-01
  Szenario: Abo-Limit ohne stillen API-Fallback behandeln
    Angenommen "Mira" verwendet das Codex-Abo als Eino-Modellzugang
    Wenn der Anbieter einen Modellaufruf wegen eines erreichten Limits ablehnt
    Dann sehe ich den Limitgrund beim betroffenen Lauf
    Und es erfolgt kein Modellaufruf über einen separat abgerechneten API-Zugang
    Und der betroffene Lauf bleibt als wegen des Limits abgelehnt erkennbar

  @MS-02
  Szenario: Laufenden Gerätecode-Verbindungsversuch verfolgen
    Angenommen ich habe noch keine verbundene Codex-Abo-Verbindung
    Wenn ich in den Modelleinstellungen die Gerätecode-Verbindung starte
    Dann sehe ich die Anmeldeseite des Anbieters und den Gerätecode
    Und ich sehe den Verbindungsversuch als laufend, bis Anmeldung, Abbruch oder Ablauf feststeht
    Und vor erfolgreicher Anmeldung wird die Verbindung nicht als einsatzbereit angezeigt

  @MS-02
  Szenario: Nach endgültigem Sitzungsfehler erneut anmelden
    Angenommen meine bisherige Codex-Abo-Sitzung wurde vom Anbieter endgültig widerrufen
    Wenn ich die betroffene Verbindung in den Modelleinstellungen öffne
    Dann sehe ich "Erneute Anmeldung erforderlich"
    Wenn ich die erneute Gerätecode-Anmeldung starte und beim Anbieter erfolgreich abschließe
    Dann sehe ich die Verbindung wieder als verbunden
    Und ein neuer Lauf kann sie bei weiterhin gültiger Anbieterfreigabe verwenden
```
