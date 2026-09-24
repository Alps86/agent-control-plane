# Wiederaufnahme

Quelle: `spec/features/zusammenarbeit/wiederaufnahme.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@ZA-06
Funktionalität: Nach Wiederholung oder Verbindungsabbruch ohne doppelte Wirkung fortfahren
  Als Betreiberin möchte ich nach einer Störung den wirklichen Arbeitsstand sehen.

  Szenario: Dieselbe gesendete Nachricht nach Zeitüberschreitung erneut übermitteln
    Angenommen ich habe im Chat die Nachricht "Lege die Aufgabe Texte prüfen an" gesendet
    Und ich habe vor der Empfangsbestätigung eine Zeitüberschreitung erhalten
    Wenn die Anwendung dieselbe Sendung erneut übermittelt
    Dann sehe ich die Nachricht nur einmal im Gespräch
    Und die zugehörige Aufgabe wird höchstens einmal angelegt

  Szenario: Nach Verbindungsabbruch den bestätigten Stand laden
    Angenommen "Mira" bearbeitet meinen Auftrag im Chat
    Wenn die Verbindung während der Bearbeitung abbricht und ich mich erneut verbinde
    Dann sehe ich den bestätigten Nachrichten- und Aufgabenstand
    Und eine bereits bestätigte Aktion wird durch die Wiederverbindung nicht erneut ausgeführt

  Szenario: Absichtlicher neuer Auftrag bleibt möglich
    Angenommen die Aufgabe "Texte prüfen" wurde durch eine Chatnachricht angelegt
    Wenn ich ausdrücklich eine neue Aufgabe "Bilder prüfen" beauftrage
    Dann sehe ich die neue Aufgabe zusätzlich zu "Texte prüfen"
    Und beide Aufträge bleiben ihren jeweiligen Nachrichten zugeordnet

  Szenario: Unklaren Ausgang einer externen Aktion sichtbar lassen
    Angenommen "Mira" hat eine freigegebene externe Aktion für meinen Chat-Auftrag gestartet
    Und die Verbindung zur externen Stelle bricht vor ihrer Bestätigung ab
    Wenn ich mich erneut mit dem Chat verbinde
    Dann sehe ich den Ausgang der Aktion als ungeklärt
    Und die Anwendung behauptet weder Erfolg noch führt sie die Aktion ungefragt erneut aus

  Szenario: Rechteentzug zwischen Annahme und Ausführung beachten
    Angenommen ich habe im Chat eine Aktion für "Mira" beauftragt
    Und "Mira" hatte beim Annehmen die dafür nötige Freigabe
    Wenn die Freigabe vor dem tatsächlichen Aufruf entzogen wird
    Dann wird der Aufruf verweigert
    Und ich sehe den Rechteentzug als Grund am Auftrag

  Szenario: Bewusst einen neuen Versuch für eine fehlgeschlagene Aktion starten
    Angenommen die beauftragte Aktion "Texte prüfen anlegen" ist fehlgeschlagen
    Wenn ich ausdrücklich einen neuen Versuch starte
    Dann sehe ich den neuen Versuch getrennt vom fehlgeschlagenen Versuch
    Und eine inzwischen erfolgreich angelegte Aufgabe wird nicht nochmals angelegt
```
