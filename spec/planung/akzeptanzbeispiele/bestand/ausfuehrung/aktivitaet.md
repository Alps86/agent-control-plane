# Aktivitaet

Quelle: `spec/features/ausfuehrung/aktivitaet.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Verlauf und Arbeitsprodukte bleiben prüfbar

  Szenario: Änderung im Aktivitätsverlauf finden
    Angenommen "Mira" wurde eine Aufgabe zugewiesen
    Wenn ich den Aktivitätsverlauf von "Nordstern" öffne
    Dann sehe ich Zuweisung, Zeitpunkt und betroffene Aufgabe

  Szenario: Laufdetails und Arbeitsprodukt prüfen
    Angenommen ein Lauf von "Mira" hat ein Arbeitsprodukt erzeugt
    Wenn ich den Lauf öffne
    Dann sehe ich seinen Auslöser, Status und sein Ergebnis
    Und ich kann das Arbeitsprodukt von dort öffnen

  Szenario: Sitzung sicher fortsetzen
    Angenommen ein Adapter unterstützt die Fortsetzung einer Sitzung
    Und ein früherer Lauf von "Mira" hat eine fortsetzbare Sitzung
    Wenn ich die Aufgabe erneut ausführen lasse
    Dann sehe ich im Lauf, ob die Sitzung fortgesetzt oder neu begonnen wurde
```
