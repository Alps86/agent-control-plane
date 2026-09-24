# Aufgaben

Quelle: `spec/features/aufgaben/aufgaben.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Aufgaben anlegen und verfolgen

  @poc
  Szenario: Aufgabe einem Agenten zuweisen
    Angenommen das Projekt "Website" und der Agent "Mira" bestehen in "Nordstern"
    Wenn ich "Startseite prüfen" mit Beschreibung und Priorität im Projekt anlege
    Und die Aufgabe "Mira" zuweise
    Dann sehe ich "Mira" als zuständigen Agenten
    Und die Aufgabe steht zunächst auf "Offen"

  Szenario: Teilaufgabe und Blockade sichtbar machen
    Angenommen die Aufgabe "Startseite freigeben" besteht
    Wenn ich darunter "Texte prüfen" anlege
    Und "Texte prüfen" als Blockade für "Startseite freigeben" markiere
    Dann zeigt "Startseite freigeben" die offene Blockade
    Und "Texte prüfen" zeigt seine übergeordnete Aufgabe

  Szenario: Kommentar und Arbeitsprodukt festhalten
    Angenommen die Aufgabe "Texte prüfen" besteht
    Wenn ich den Kommentar "Fachprüfung folgt" und ein Arbeitsprodukt hinzufüge
    Dann sehe ich den Kommentar mit Zeitangabe im Aufgabenverlauf
    Und ich kann das Arbeitsprodukt von der Aufgabe aus öffnen

  @poc
  Szenario: Status nach Neustart behalten
    Angenommen die Aufgabe "Texte prüfen" steht auf "In Arbeit"
    Wenn ich die Anwendung neu starte und "Texte prüfen" öffne
    Dann steht die Aufgabe weiter auf "In Arbeit"
```
