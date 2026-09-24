# Struktur und arbeit

Quelle: `spec/features/kern-ergaenzungen/struktur-und-arbeit.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Organisationsstruktur und Arbeit bleiben auch in Grenzfällen konsistent

  @ORG-01 @poc
  Szenario: Organisation vor dem ersten Ziel eigenständig anlegen
    Angenommen es gibt noch keine Organisation
    Wenn ich "Nordstern" mit Name und Beschreibung anlege
    Dann sehe ich die gespeicherte Organisation
    Wenn ich die Anwendung neu starte
    Dann sehe ich "Nordstern" erneut

  @GOAL-01 @poc
  Szenario: Erstes Ziel nach der Organisation anlegen
    Angenommen "Nordstern" ist als Organisation gespeichert
    Wenn ich das Stammziel "Portal veröffentlichen" anlege
    Dann sehe ich das Ziel in der Zielübersicht von "Nordstern"
    Wenn ich die Anwendung neu starte
    Dann sehe ich das Ziel erneut

  @PROJ-01 @poc
  Szenario: Projekt erhält beim Anlegen seinen Zielbezug
    Angenommen "Nordstern" hat das Ziel "Portal veröffentlichen"
    Wenn ich das Projekt "Website" mit diesem Ziel anlege
    Dann zeigt das Projektdetail "Portal veröffentlichen" als verknüpftes Ziel
    Wenn ich die Aufgabe "Startseite prüfen" im Projekt "Website" anlege
    Dann zeigt die Aufgabe ihren Zielpfad über "Website"

  @GOAL-03
  Szenario: Fremdes Ziel und Kreis im Zielbaum ablehnen
    Angenommen "Nordstern" und "Südstern" haben eigene Ziele
    Wenn ich ein Ziel von "Südstern" als Teilziel von "Nordstern" eintrage
    Dann wird die Änderung ohne Datenübernahme abgelehnt
    Wenn ich ein Elternziel seinem eigenen Nachfahren unterordne
    Dann bleibt der bisherige Zielbaum unverändert

  @AGT-04
  Szenario: Berichtsweg kann keinen Zyklus bilden
    Angenommen "Mira" berichtet an "Kai" in "Nordstern"
    Wenn ich "Mira" als Vorgesetzte von "Kai" eintrage
    Dann sehe ich die verletzte Berichtslinie als Fehlergrund
    Und bestehende Aufgaben von "Mira" und "Kai" behalten ihre Zuweisung

  @TASK-03
  Szenario: Blocker verhindert Aufgabenaufnahme ohne Elternbeziehung zu ändern
    Angenommen "Texte prüfen" blockiert "Startseite freigeben"
    Und "Startseite freigeben" ist "Mira" zugewiesen
    Wenn "Mira" die blockierte Aufgabe aufnehmen will
    Dann beginnt für sie kein Lauf zu "Startseite freigeben"
    Und die Blockade und die übergeordnete Aufgabe bleiben getrennt sichtbar
    Wenn "Texte prüfen" abgeschlossen wird
    Dann kann "Startseite freigeben" wieder aufgenommen werden

  @TASK-04
  Szenario: Konkurrierende Aufnahme ergibt genau einen Besitzer
    Angenommen die offene Aufgabe "Startseite freigeben" ist ausführbar
    Wenn zwei berechtigte Agenten sie gleichzeitig aufnehmen wollen
    Dann erhält genau einer die aktive Aufnahme
    Und der andere sieht einen Konflikt mit dem aktuellen Besitzer
    Und es entsteht kein zweiter aktiver Lauf für diese Aufnahme

  @DEL-02
  Szenario: Delegationszyklus und inaktiver Empfänger erzeugen keine Teilaufgabe
    Angenommen "Kai" bearbeitet eine offene Aufgabe mit einer Teilaufgabe von "Mira"
    Wenn "Mira" die Teilaufgabe zurück an "Kai" delegieren will
    Dann wird der offene Vorfahrenzyklus erklärt
    Und es entsteht keine weitere Teilaufgabe
    Wenn "Mira" an einen pausierten Agenten delegieren will
    Dann wird dessen fehlende Einsatzbereitschaft erklärt

  @TASK-06
  Szenariogrundriss: Arbeitsmodus bestimmt das erwartete Ergebnis
    Angenommen ich lege eine Aufgabe im Modus "<Modus>" für "Mira" an
    Wenn "Mira" sie ausführt
    Dann sehe ich "<Ergebnisort>" als Ergebnis
    Und der gewählte Modus bleibt am Aufgabendetail sichtbar

    Beispiele:
      | Modus | Ergebnisort                 |
      | Ask   | eine Antwort im Aufgabenthread |
      | Plan  | einen prüfbaren Plan        |
      | Agent | ein Arbeitsprodukt          |

  @TASK-07
  Szenario: Strukturierte Rückfrage nur einmal entscheiden
    Angenommen "Mira" hat an ihrer Aufgabe eine Mehrfachauswahl mit drei Optionen gestellt
    Wenn ich zwei Optionen bestätige
    Dann sieht "Mira" die gewählten Optionen im Aufgabenthread
    Wenn ich dieselbe Rückfrage erneut beantworten will
    Dann wird die zweite Antwort abgewiesen
```
