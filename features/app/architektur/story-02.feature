# language: de
@story-02 @arc-02
Funktionalität: Lokaler Datenbestand startet zuverlässig und speichert atomar
  Als Betreiber
  möchte ich denselben lokalen Datenbestand wiederverwenden und Schreibfehler ohne Teiländerung überstehen,
  damit Neustarts und fehlgeschlagene Aktionen keine widersprüchlichen Daten hinterlassen.

  Szenario: Ein zweiter Serverstart erhält die Migrationsversion
    Angenommen ein neuer beschreibbarer lokaler Datenpfad ist für den Server konfiguriert
    Wenn ich den Server starte
    Dann ist der Server über seine öffentliche HTTP-Grenze erreichbar
    Und die Migrationsversion des lokalen Datenbestands ist ermittelt
    Wenn ich den Server beende und mit demselben Datenpfad erneut starte
    Dann ist der Server über seine öffentliche HTTP-Grenze erreichbar
    Und die Migrationsversion des lokalen Datenbestands ist unverändert

  Szenario: Unzugänglicher Datenpfad verhindert den Serverstart
    Angenommen ein für den Server unzugänglicher lokaler Datenpfad ist konfiguriert
    Wenn ich den Server starte
    Dann beendet er den Start mit einem verständlichen Datenbankfehler
    Und der Server ist über seine öffentliche HTTP-Grenze nicht erreichbar

  Szenario: Eine fehlgeschlagene Migration hinterlässt keine Teiländerung
    Angenommen ein lokaler Datenbestand mit öffentlich lesbarer Schemaversion ist vorhanden
    Und eine neue versionierte Migration mit mehreren Änderungen ist konfiguriert
    Wenn ich diese Migration über die öffentliche Anwendungsfassade ausführe und ein späterer Schritt bewusst fehlschlägt
    Dann meldet die Anwendungsfassade den Migrationsfehler
    Und die öffentlich gelesene Schemaversion ist unverändert
    Wenn ich die korrigierte Migration über dieselbe Anwendungsfassade erneut ausführe
    Dann wird die korrigierte Migration einschließlich ihres unveränderten ersten Schritts ohne Teiländerung aus dem Fehlversuch erfolgreich abgeschlossen
    Und die neue Schemaversion ist öffentlich lesbar
    Und alle Änderungen dieser Migration sind gemeinsam wirksam
