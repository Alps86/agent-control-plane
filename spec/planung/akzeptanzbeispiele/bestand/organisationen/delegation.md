# Delegation

Quelle: `spec/features/organisationen/delegation.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Arbeit nachvollziehbar delegieren

  Szenario: Aufgabe delegieren und Rückmeldung erhalten
    Angenommen "Kai" darf Aufgaben an "Mira" delegieren
    Und "Kai" bearbeitet "Startseite freigeben"
    Wenn "Kai" die Teilaufgabe "Texte prüfen" an "Mira" delegiert
    Dann sehe ich "Kai" als Auftraggeber und "Mira" als Empfängerin
    Wenn "Mira" die Teilaufgabe mit einem Ergebnis abschließt
    Dann sieht "Kai" die Rückmeldung an der übergeordneten Aufgabe

  Szenario: Unzulässige Delegation ablehnen
    Angenommen "Kai" darf nicht an "Mira" delegieren
    Wenn "Kai" eine Aufgabe an "Mira" delegieren will
    Dann wird die Delegation abgelehnt
    Und die Aufgabe bleibt bei "Kai"

  Szenario: Delegation über Organisationsgrenzen ablehnen
    Angenommen "Kai" gehört zu "Nordstern" und "Lena" zu "Südstern"
    Wenn "Kai" eine Aufgabe von "Nordstern" an "Lena" delegieren will
    Dann wird die Delegation abgelehnt
    Und "Südstern" sieht die Aufgabe nicht
```
