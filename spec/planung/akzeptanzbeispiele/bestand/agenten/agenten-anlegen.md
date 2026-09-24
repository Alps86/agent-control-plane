# Agenten anlegen

Quelle: `spec/features/agenten/agenten-anlegen.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Agenten ohne Entwicklerwissen anlegen

  @poc
  Szenario: Eino-Agent aus einer Vorlage anlegen
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich die Vorlage "Recherche" auswähle
    Und den Agenten "Mira" mit der Ausführungsart "Eino" anlege
    Dann sehe ich "Mira" mit Rolle und Aufgabenbeschreibung in der Agentenübersicht
    Und seine Ausführungsart ist "Eino"
    Und ein Shell-Werkzeug ist nicht freigegeben

  @poc
  Szenario: Codex-CLI-Agent anlegen
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich den Agenten "Kai" mit der Ausführungsart "Codex CLI" anlege
    Dann sehe ich "Kai" mit der Ausführungsart "Codex CLI"
    Und ich kann seine Ausführungsgrenzen über die Oberfläche prüfen

  @poc
  Szenario: Unvollständige Konfiguration erklären
    Angenommen ich lege einen Agenten mit der Ausführungsart "Codex CLI" an
    Wenn die für einen Start nötige Konfiguration fehlt
    Dann sehe ich einen verständlichen Hinweis zur fehlenden Angabe
    Und der Agent wird nicht als einsatzbereit angezeigt

  Szenario: Berichtsweg festlegen
    Angenommen "Mira" und "Kai" sind Agenten von "Nordstern"
    Wenn ich "Kai" als Vorgesetzten von "Mira" auswähle
    Dann zeigt die Organisationsansicht "Mira" unter "Kai"
```
