# language: de
@story-34 @proj-02 @browser
Funktionalität: Projektarchiv im Browser bedienen
  Als lokale Betreiberin
  möchte ich Projekte aus der Übersicht archivieren und zurückholen,
  damit ich den aktiven Bestand und vorhandene Arbeit unterscheiden kann.

  Szenario: Archivieren und Wiederherstellen sind im Browser sichtbar
    Angenommen ich öffne die Anwendung mit "Nordstern" und dem Projekt "Website" im Browser
    Und der Agent "Mira" besteht in "Nordstern"
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Wenn ich die Projektübersicht von "Nordstern" öffne
    Und ich "Website" über die sichtbare Aktion archiviere
    Dann sehe ich "Website" nicht mehr in der aktiven Projektliste
    Wenn ich das Projektarchiv von "Nordstern" öffne
    Dann sehe ich "Website" mit dem Status "Archiviert"
    Und ich kann die bestehende Aufgabe "Startseite prüfen" weiterhin öffnen
    Wenn ich "Website" über die sichtbare Aktion wiederherstelle
    Dann sehe ich "Website" erneut in der aktiven Projektliste

  Szenario: Direktes Formular für neue Arbeit an archiviertem Projekt bleibt gesperrt
    Angenommen ich öffne die Anwendung mit "Nordstern" und dem Projekt "Website" im Browser
    Und der Agent "Mira" besteht in "Nordstern"
    Und "Website" ist in "Nordstern" archiviert
    Wenn ich die direkte Browseradresse für eine neue Aufgabe in "Website" öffne
    Dann sehe ich einen verständlichen Archivhinweis statt eines speicherbaren Aufgabenformulars
    Und im Projekt "Website" entsteht keine neue Aufgabe
