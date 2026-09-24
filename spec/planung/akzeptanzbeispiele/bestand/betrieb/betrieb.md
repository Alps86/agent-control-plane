# Betrieb

Quelle: `spec/features/betrieb/betrieb.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Lokaler Betrieb ohne Anmeldung und Budgetverwaltung

  Szenario: Anwendung ohne Login nutzen
    Angenommen die Anwendung wurde lokal gestartet
    Wenn ich ihre Startseite öffne
    Dann sehe ich die Organisationsübersicht ohne Anmeldeformular
    Und ich kann eine Organisation ohne Registrierung anlegen

  Szenario: Budgetfunktionen fehlen in Version 1
    Angenommen ich öffne Organisation, Projekt und Agentenkonfiguration
    Dann sehe ich dort keine Budgetfelder oder Budgetlimits

  Szenario: Arbeitsstand nach Neustart wiederfinden
    Angenommen ich habe "Nordstern", das Projekt "Website" und eine abgeschlossene Aufgabe angelegt
    Wenn ich die Anwendung neu starte
    Dann sehe ich Organisation, Projekt, Aufgabe und Ergebnis wieder

  Szenario: Betriebsfehler sichtbar machen
    Angenommen ein Agentenlauf ist wegen eines nicht erreichbaren Ausführungsadapters fehlgeschlagen
    Wenn ich die Monitoring-Ansicht öffne
    Dann sehe ich den betroffenen Agenten und den Fehlerzustand
```
