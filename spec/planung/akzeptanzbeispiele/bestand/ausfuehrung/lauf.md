# Lauf

Quelle: `spec/features/ausfuehrung/lauf.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@poc
Funktionalität: Agentenarbeit ausführen und nachvollziehen

  Szenariogrundriss: Eine Aufgabe mit der gewählten Ausführungsart bearbeiten
    Angenommen der einsatzbereite Agent "<Agent>" nutzt "<Ausführungsart>"
    Und die Aufgabe "Startseite prüfen" ist ihm zugewiesen
    Wenn ich die Aufgabe zur Ausführung starte
    Dann sehe ich einen neuen Lauf mit einem Status
    Und nach Abschluss sehe ich Ergebnis oder Fehlergrund bei diesem Lauf

    Beispiele:
      | Agent | Ausführungsart |
      | Mira  | Eino          |
      | Kai   | Codex CLI     |

  Szenario: Laufverlauf nach Neustart behalten
    Angenommen "Mira" hat die Aufgabe "Startseite prüfen" abgeschlossen
    Wenn ich die Anwendung neu starte und die Aufgabe öffne
    Dann sehe ich ihren abgeschlossenen Lauf mit Zeitangabe und Ergebnis

  Szenario: Fehler sichtbar machen und erneut ausführen
    Angenommen der letzte Lauf von "Startseite prüfen" ist fehlgeschlagen
    Wenn ich die Aufgabe öffne
    Dann sehe ich einen verständlichen Fehlergrund
    Wenn ich einen neuen Lauf starte
    Dann bleibt der fehlgeschlagene Lauf im Verlauf sichtbar
```
