# language: de
Funktionalität: OPS-01 – Lokalen Go-Server mit SQLite starten
  Als Betreiber
  möchte ich den Datenpfad für meinen lokalen Server festlegen,
  damit die Anwendung beim Start ihr Schema bereitstellt und Startfehler sichtbar sind.

  Szenario: Neuer lokaler Datenbestand wird beim Serverstart bereitgestellt
    Angenommen ein neuer beschreibbarer SQLite-Datenpfad ist über APP_DB_PATH konfiguriert
    Wenn ich den lokalen Go-Server starte
    Dann antwortet seine öffentliche HTTP-Statusroute erfolgreich mit einer Schemaversion
    Und am konfigurierten Pfad liegt eine SQLite-Datenbank

  Szenario: Erneuter Start verwendet denselben Datenbestand
    Angenommen ein neuer beschreibbarer SQLite-Datenpfad ist über APP_DB_PATH konfiguriert
    Wenn ich den lokalen Go-Server starte
    Dann antwortet seine öffentliche HTTP-Statusroute erfolgreich mit einer Schemaversion
    Und ich die Schemaversion über HTTP festhalte
    Und ich den Server beende und mit demselben APP_DB_PATH erneut starte
    Dann antwortet seine öffentliche HTTP-Statusroute erfolgreich mit einer Schemaversion
    Und die Schemaversion ist unverändert

  Szenario: Vorhandene unzugängliche Datenbank verhindert den Start
    Angenommen eine vorhandene unzugängliche SQLite-Datei ist über APP_DB_PATH konfiguriert
    Wenn ich den lokalen Go-Server starte
    Dann beendet sich der Server mit einem sichtbaren Datenbank-Startfehler
    Und seine öffentliche HTTP-Statusroute ist nicht erreichbar
    Und die vorhandene SQLite-Datei wurde nicht ersetzt
