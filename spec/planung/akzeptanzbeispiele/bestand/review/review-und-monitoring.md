# Review und monitoring

Quelle: `spec/features/review/review-und-monitoring.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Review, Reporting und Monitoring als eigene Zuständigkeiten

  Szenario: Ergebnis zur Prüfung geben
    Angenommen "Mira" hat die Aufgabe "Texte prüfen" abgeschlossen
    Und "Ruth" ist als Reviewerin für diese Arbeit zuständig
    Wenn "Mira" das Ergebnis zur Prüfung einreicht
    Dann sieht "Ruth" eine offene Prüfung mit Aufgabe und Ergebnis
    Und die Aufgabe gilt noch nicht als freigegeben

  Szenario: Prüfung mit Begründung zurückgeben
    Angenommen "Ruth" prüft "Texte prüfen"
    Wenn "Ruth" eine Überarbeitung mit der Begründung "Quelle fehlt" verlangt
    Dann sieht "Mira" die Entscheidung und Begründung
    Und die Aufgabe ist wieder bearbeitbar

  Szenario: Ergebnis freigeben
    Angenommen "Ruth" prüft das Ergebnis von "Texte prüfen"
    Wenn "Ruth" das Ergebnis freigibt
    Dann sieht "Mira" die Freigabe im Aufgabenverlauf
    Und die Aufgabe ist als freigegeben gekennzeichnet

  Szenario: Neue Agentin zur Freigabe vorlegen
    Angenommen "Ruth" ist für die Freigabe neuer Agenten zuständig
    Wenn ich den neuen Agenten "Lena" zur Freigabe vorlege
    Dann sieht "Ruth" eine offene Anfrage mit der Agentenkonfiguration
    Wenn "Ruth" die Anfrage annimmt
    Dann wird "Lena" als einsatzbereit angezeigt

  Szenario: Bericht und Monitoring zeigen tatsächlichen Stand
    Angenommen "Nordstern" hat offene, laufende und abgeschlossene Aufgaben
    Und "Rolf" ist Reporter und "Moni" überwacht die Arbeit
    Wenn ich die Übersichten für Reporting und Monitoring öffne
    Dann sehe ich Status, letzte Aktivität und fehlerhafte Läufe der Organisation
    Und ich kann von einer Auffälligkeit zur betroffenen Aufgabe wechseln
```
