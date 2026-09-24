# language: de
@story-19 @perm-01 @eino @browser
Funktionalität: Betreiberin vergibt den Datenbereich eines Eino-Agenten im Browser
  Als lokale Betreiberin
  möchte ich die Projektfreigaben eines Agenten sehen und ändern,
  damit seine Datenrechte ausdrücklich und nachvollziehbar bleiben.

  Hintergrund:
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin

  Szenario: Lesefreigabe im Browser erteilen und wieder entziehen
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" besteht
    Und "Mira" ist ein Eino-Agent von "Nordstern" ohne Datenbereich
    Wenn ich "Mira" in einem echten Browser öffne
    Dann sehe ich, dass keine Projektfreigabe besteht
    Wenn ich "Website" zum Lesen freigebe und speichere
    Dann zeigt die Seite "Website" mit Lesen und ohne Schreiben
    Wenn ich die Lesefreigabe wieder entziehe und speichere
    Dann zeigt die Seite erneut keinen Datenbereich für "Mira"
    Und der nächste Lesezugriff von "Mira" auf "Website" wird ohne Projektdaten verweigert

  Szenario: Ein fremdes Projekt erscheint nicht und ist per Direktaufruf nicht freigebbar
    Angenommen "Mira" ist ein Eino-Agent von "Nordstern"
    Und "Fremd" ist ein Projekt der anderen Organisation "Suedstern"
    Wenn ich "Mira" Datenbereich in einem echten Browser öffne
    Dann sehe ich "Fremd" nicht als auswählbares Projekt
    Wenn ich die Kennung von "Fremd" direkt an den Speichern-Endpunkt sende
    Dann wird die Änderung ohne fremde Projektdaten abgewiesen
    Und "Mira" erhält keine Freigabe für "Fremd"

  # Der Codex-CLI-Datenbereich ist im App-Feature dieser Story geprüft.
  # Diese Browserfälle prüfen nur den Eino-Bedienpfad. Die tatsächliche
  # Codex-CLI-Ausführung mit durchgesetztem Toolprofil gehört zu Story-21.
