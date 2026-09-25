# language: de
@story-35 @proj-03 @browser
Funktionalität: Privaten Ausführungsort eines Projekts im Browser verwalten
  Als lokale Betreiberin
  möchte ich den erlaubten Projektarbeitsordner und seine Startwirkung sehen,
  damit ich eine fehlende oder ungültige Bindung korrigieren kann.

  Szenario: Ausführungsort vom Projektdetail aus aktivieren
    Angenommen ich öffne die Anwendung mit dem Projekt "Website" in "Nordstern" im Browser
    Wenn ich im Projektdetail den Ausführungsort öffne
    Dann sehe ich den privaten appverwalteten Projektarbeitsordner ohne frei editierbare Pfadeingabe
    Und der Ausführungsort ist zunächst deaktiviert
    Wenn ich den privaten Ausführungsort aktiviere und speichere
    Dann sehe ich den aktivierten privaten Ort im Projektdetail
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich den aktivierten privaten Ort erneut

  Szenario: Fehlender Ort erklärt die Startblockade
    Angenommen ich öffne die Anwendung mit dem Projekt "Website" in "Nordstern" im Browser
    Wenn ich im Projektdetail den Ausführungsort öffne
    Dann sehe ich den fehlenden Ausführungsort und die Korrekturaktion zum privaten Projektarbeitsordner

  Szenario: Fremdes Projekt zeigt keinen privaten Ausführungsort
    Angenommen ich öffne die Anwendung mit Projekten in "Nordstern" und "Südstern" im Browser
    Wenn ich über "Nordstern" den Ausführungsort eines Projekts aus "Südstern" öffne
    Dann sehe ich keine Projektdaten und keinen privaten Pfad
