# Ziele

Quelle: `spec/features/ziele/ziele.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Ziele erklären den Zweck der Arbeit

  Szenario: Teilziel mit Projekt verbinden
    Angenommen "Nordstern" hat das Ziel "Wissensportal veröffentlichen"
    Und das Projekt "Website" besteht
    Wenn ich das Teilziel "Startseite freigeben" anlege und mit "Website" verbinde
    Dann sehe ich "Startseite freigeben" unter dem Organisationsziel
    Und das Projekt zeigt das verknüpfte Ziel

  Szenario: Zielkette einer Aufgabe anzeigen
    Angenommen die Aufgabe "Texte prüfen" gehört zum Ziel "Startseite freigeben"
    Wenn ich "Texte prüfen" öffne
    Dann sehe ich den Weg bis zum Organisationsziel "Wissensportal veröffentlichen"

  Szenario: Kreisförmige Zielverknüpfung ablehnen
    Angenommen das Ziel "Startseite freigeben" ist ein Teilziel von "Wissensportal veröffentlichen"
    Wenn ich "Wissensportal veröffentlichen" dem Ziel "Startseite freigeben" unterordnen will
    Dann wird die Änderung mit einer verständlichen Begründung abgelehnt
```
