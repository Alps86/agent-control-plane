# Skills und integrationen

Quelle: `spec/features/integrationen/skills-und-integrationen.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Fähigkeiten und Integrationen gezielt bereitstellen

  Szenario: Freigegebenen Skill einem Agenten zuweisen
    Angenommen in "Nordstern" gibt es den geprüften Skill "Quellen prüfen"
    Und "Mira" ist ein Agent von "Nordstern"
    Wenn ich "Quellen prüfen" für "Mira" freigebe
    Dann sehe ich den Skill in der Agentenkonfiguration von "Mira"
    Und ein Agent von "Südstern" sieht diesen Skill nicht

  Szenario: Nicht freigegebene Integration verweigern
    Angenommen "Mira" darf die Integration "Websuche" verwenden
    Wenn "Mira" über die öffentliche Ausführungsschnittstelle eine andere Integration aufruft
    Dann wird der Aufruf verweigert
    Und ich sehe die Verweigerung im Aktivitätsverlauf

  Szenario: Fehlende Integration verständlich anzeigen
    Angenommen die Integration "Websuche" ist für "Mira" freigegeben
    Wenn ihr Dienst bei einem Lauf nicht erreichbar ist
    Dann sehe ich den betroffenen Lauf und einen verständlichen Fehlergrund
```
