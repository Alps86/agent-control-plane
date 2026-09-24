# Verknuepfungen

Quelle: `spec/features/zusammenarbeit/verknuepfungen.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@ZA-03
Funktionalität: Nachrichten, Aufgaben und Arbeitsprodukte verbinden
  Als Betreiberin möchte ich aus einem Gespräch zur beauftragten Arbeit und ihrem Ergebnis wechseln.

  Szenario: Aus einer Nachricht eine Aufgabe erzeugen
    Angenommen ich bespreche im Chat mit "Mira" das Projekt "Website" in "Nordstern"
    Wenn ich aus der Nachricht "Prüfe die Startseite" eine Aufgabe für "Mira" anlege
    Dann sehe ich eine Aufgabe mit dem Auftrag im Projekt "Website"
    Und die Nachricht verweist auf diese Aufgabe
    Und die Aufgabe verweist auf die auslösende Nachricht

  Szenario: Ergebnis im Chat und an der Aufgabe finden
    Angenommen die Aufgabe "Startseite prüfen" ist mit einer Chatnachricht verknüpft
    Und "Mira" hat dazu das Arbeitsprodukt "Prüfbericht" erstellt
    Wenn ich das Ergebnis im Chat öffne
    Dann sehe ich die verknüpfte Aufgabe und den "Prüfbericht"
    Und ich kann den "Prüfbericht" auch von der Aufgabe aus öffnen

  Szenario: Eine bloße Unterhaltung erzeugt keine Aufgabe
    Angenommen ich spreche mit "Mira" über das Projekt "Website"
    Wenn ich die Nachricht "Wie ist der Stand?" sende
    Dann bleibt die Nachricht im Chatverlauf sichtbar
    Und es wird ohne Auftrag zur Aufgabenerstellung keine neue Aufgabe angelegt

  Szenario: Vorschlag, beauftragte Aktion und ausgeführtes Ergebnis unterscheiden
    Angenommen "Mira" schlägt im Chat die Aufgabe "Texte prüfen" vor
    Wenn ich den Vorschlag noch nicht beauftragt habe
    Dann sehe ich ihn als Vorschlag und nicht als bestehende Aufgabe
    Wenn ich die Anlage der Aufgabe bestätige
    Dann sehe ich den Auftrag zunächst als beauftragte Aktion
    Und nach erfolgreicher Ausführung sehe ich die angelegte Aufgabe mit Verweis auf die Nachricht

  Szenario: Fehlgeschlagene Aktion nicht als fertige Aufgabe darstellen
    Angenommen ich habe im Chat die Anlage von "Texte prüfen" beauftragt
    Wenn die Anlage fehlschlägt
    Dann sehe ich den Fehler an der beauftragten Aktion
    Und ich sehe keinen Link auf eine erfolgreich angelegte Aufgabe
```
