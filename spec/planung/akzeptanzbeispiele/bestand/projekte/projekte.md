# Projekte

Quelle: `spec/features/projekte/projekte.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Projekte verbinden Ziele mit konkreter Arbeit

  @poc
  Szenario: Projekt in einer Organisation anlegen
    Angenommen die Organisation "Nordstern" mit dem Ziel "Wissensportal veröffentlichen" besteht
    Wenn ich in "Nordstern" das Projekt "Website" mit Beschreibung "Öffentlicher Auftritt" anlege
    Dann sehe ich "Website" in der Projektübersicht von "Nordstern"
    Und ich kann das Projekt mit dem Ziel "Wissensportal veröffentlichen" verknüpfen

  @poc
  Szenario: Projektzugehörigkeit einer Aufgabe
    Angenommen das Projekt "Website" gehört zu "Nordstern"
    Wenn ich im Projekt die Aufgabe "Startseite prüfen" anlege
    Dann sehe ich die Aufgabe in der Aufgabenliste von "Website"
    Und ich sehe ihre Zugehörigkeit zu "Nordstern"

  Szenario: Projekt archivieren und wiederherstellen
    Angenommen das Projekt "Website" ist aktiv
    Wenn ich "Website" archiviere
    Dann erscheint es nicht mehr in der Liste aktiver Projekte
    Wenn ich das archivierte Projekt wiederherstelle
    Dann erscheint "Website" wieder in der Liste aktiver Projekte
```
