# Delegation und berichte

Quelle: `spec/features/zusammenarbeit/delegation-und-berichte.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@ZA-04
Funktionalität: Delegation und Rückmeldung im Gespräch verfolgen
  Als Betreiberin möchte ich Übergaben und Ergebnisse im Arbeitsgespräch nachvollziehen.

  Szenario: Erlaubte Delegation aus dem Teamchat verfolgen
    Angenommen "Kai" bearbeitet "Startseite freigeben" im Team "Web"
    Und "Kai" darf "Texte prüfen" an "Mira" delegieren
    Wenn "Kai" die Teilaufgabe "Texte prüfen" an "Mira" delegiert
    Dann sehe ich Auftraggeber, Empfängerin und Teilaufgabe im Teamchat
    Und die Teilaufgabe verweist auf "Startseite freigeben"

  Szenario: Fortschritt und Abschluss an den Auftraggeber melden
    Angenommen "Kai" hat "Texte prüfen" an "Mira" delegiert
    Wenn "Mira" den Fortschritt meldet und die Teilaufgabe mit einem Ergebnis abschließt
    Dann sieht "Kai" Fortschritt und Ergebnis im Gespräch zur Aufgabe
    Und ich kann von jeder Rückmeldung die betroffene Teilaufgabe öffnen

  Szenario: Unzulässige Delegation im Chat verweigern
    Angenommen "Kai" darf nicht an "Lena" delegieren
    Wenn "Kai" im Chat die Aufgabe "Texte prüfen" an "Lena" übergeben will
    Dann sehe ich einen verständlichen Grund für die Verweigerung
    Und "Lena" erhält weder die Aufgabe noch ihren Gesprächskontext

  Szenario: Teambericht verweist auf belegte Arbeit und offenen Review
    Angenommen "Web" hat die Aufgabe "Startseite freigeben" bearbeitet
    Und "Texte prüfen" wurde mit dem Arbeitsprodukt "Prüfbericht" abgeschlossen
    Und die fachliche Prüfung von "Texte prüfen" ist noch offen
    Wenn ich einen Bericht über die Arbeit von "Web" anfordere
    Dann sehe ich den Bearbeitungsstand mit Verweisen auf Aufgabe und "Prüfbericht"
    Und der Bericht kennzeichnet die fachliche Prüfung als offen
    Und er bezeichnet "Texte prüfen" nicht als freigegeben
```
