# Review und entscheidungen

Quelle: `spec/features/kern-ergaenzungen/review-und-entscheidungen.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Entscheidungen steuern Arbeit und bleiben nachvollziehbar

  @APP-01
  Szenario: Abgelehnte Agentenanlage bleibt inaktiv
    Angenommen "Lena" wurde mit Konfiguration und Begründung zur Freigabe vorgelegt
    Wenn ich die Anlage mit der Begründung "Auftrag unklar" ablehne
    Dann ist "Lena" nicht einsatzbereit
    Und die Ablehnung und Begründung stehen in der Anfragehistorie
    Wenn die Konfiguration überarbeitet und erneut vorgelegt wird
    Dann sehe ich eine neue offene Entscheidung mit der Vorgeschichte

  @REVIEW-02 @REVIEW-03
  Szenario: Erzwungene Prüfung verhindert direkten Abschluss
    Angenommen "Texte prüfen" verlangt Qualitätsreview durch "Ruth" und danach Freigabe durch "Kai"
    Und "Mira" führt die Aufgabe aus
    Wenn "Mira" mit einem Ergebnis-Kommentar abschließen will
    Dann steht die Aufgabe auf "In Prüfung" bei "Ruth"
    Und sie steht noch nicht auf "Erledigt"
    Wenn "Ruth" mit einem Kommentar freigibt
    Dann liegt die Freigabe bei "Kai"
    Wenn "Kai" mit einem Kommentar freigibt
    Dann steht die Aufgabe auf "Erledigt"

  @REVIEW-03
  Szenario: Überarbeitung geht an den ursprünglichen Ausführenden zurück
    Angenommen "Ruth" prüft das Ergebnis von "Mira" an "Texte prüfen"
    Wenn "Ruth" mit der Begründung "Quelle fehlt" Überarbeitung verlangt
    Dann ist "Mira" wieder zuständig
    Und der Entscheid mit Begründung steht im Aufgabenverlauf
    Wenn "Ruth" denselben Entscheid erneut absendet
    Dann wird die zweite Entscheidung als Konflikt abgewiesen

  @VIEW-01
  Szenario: Blockierte Arbeit in der Inbox finden
    Angenommen "Startseite freigeben" wartet auf einen offenen Blocker
    Wenn ich die Ansicht "Blockiert" öffne
    Dann sehe ich Aufgabe, Blockadegrund, zuständige Person und Stillstandsdauer
    Und ich kann zur betroffenen Aufgabe wechseln
    Wenn der Blocker erledigt ist
    Dann verschwindet die Aufgabe aus der Ansicht "Blockiert"

  @VIEW-02
  Szenario: Veralteten Entscheidungsvorschlag nicht ausführen
    Angenommen ein Agent schlägt eine Entscheidung zu "Startseite freigeben" vor
    Und die Aufgabe wurde seit dem Vorschlag geändert
    Wenn ich den Vorschlag in "Entscheidungen" öffne
    Dann sehe ich den geänderten Zielstand
    Und eine auf dem alten Stand beruhende Aktion ist nicht auswählbar

  @NAV-01
  Szenario: Suche respektiert Organisation und Archiv
    Angenommen "Website" gehört zu "Nordstern" und "Altprojekt" ist archiviert
    Und "Südstern" hat ein eigenes Projekt
    Wenn ich in "Nordstern" die Befehlssuche öffne
    Dann finde ich "Website" sowie Aktionen für neue Aufgaben und Agenten
    Und ich finde weder "Altprojekt" noch das Projekt von "Südstern"
```
