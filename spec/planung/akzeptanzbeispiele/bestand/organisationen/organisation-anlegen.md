# Organisation anlegen

Quelle: `spec/features/organisationen/organisation-anlegen.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@poc
Funktionalität: Eine Organisation mit einem gemeinsamen Ziel anlegen
  Damit Arbeit einen nachvollziehbaren Zweck hat, legt die Betreiberin zuerst eine Organisation an.

  Szenario: Erste Organisation anlegen
    Angenommen ich öffne die Anwendung ohne vorhandene Organisation
    Wenn ich die Organisation "Nordstern" mit dem Ziel "Wissensportal veröffentlichen" anlege
    Dann sehe ich "Nordstern" in der Organisationsübersicht
    Und ich sehe das Ziel "Wissensportal veröffentlichen" bei "Nordstern"

  Szenario: Pflichtangaben verständlich erklären
    Angenommen ich öffne das Formular für eine neue Organisation
    Wenn ich das Formular ohne Namen absende
    Dann sehe ich einen Hinweis am Feld "Name"
    Und es wird keine neue Organisation angezeigt

  Szenario: Organisationen trennen
    Angenommen es gibt die Organisationen "Nordstern" und "Südstern"
    Und die Aufgabe "Startseite prüfen" gehört zu "Nordstern"
    Wenn ich "Südstern" auswähle
    Dann sehe ich die Aufgabe "Startseite prüfen" nicht in deren Aufgabenübersicht
```
