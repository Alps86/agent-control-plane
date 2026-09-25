# language: de
@story-34 @proj-02
Funktionalität: Betreiber pflegt Projektstatus und Archiv
  Als lokale Betreiberin
  möchte ich abgeschlossene Projekte archivieren und wiederherstellen,
  damit bestehende Aufgaben nachvollziehbar bleiben und keine neue Arbeit im Archiv entsteht.

  Szenario: Archivieren entfernt ein Projekt nur aus der aktiven Liste
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Wenn ich "Website" in "Nordstern" über HTTP archiviere
    Dann meldet der Server den Status "Archiviert" für "Website"
    Und die aktive Projektliste von "Nordstern" enthält "Website" nicht
    Und das Projektarchiv von "Nordstern" enthält "Website" genau einmal
    Und die bestehende Aufgabe "Startseite prüfen" bleibt unter derselben Kennung erreichbar

  Szenario: Wiederherstellung erhält Aufgaben und erlaubt neue Arbeit
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Und "Website" ist in "Nordstern" archiviert
    Wenn ich "Website" in "Nordstern" über HTTP wiederherstelle
    Dann meldet der Server den Status "Aktiv" für "Website"
    Und die aktive Projektliste von "Nordstern" enthält "Website" genau einmal
    Und das Projektarchiv von "Nordstern" enthält "Website" nicht
    Und die bestehende Aufgabe "Startseite prüfen" bleibt unter derselben Kennung erreichbar
    Und ich kann die neue Aufgabe "Texte prüfen" im Projekt "Website" anlegen

  Szenario: Ein archiviertes Projekt nimmt über die direkte Aufgaben-URL keine neue Arbeit an
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und "Website" ist in "Nordstern" archiviert
    Wenn ich die Aufgabe "Verbotener Auftrag" direkt über die Aufgaben-URL von "Website" anlege
    Dann wird neue Arbeit mit einem konkreten Archivhinweis verweigert
    Und in "Website" entsteht keine neue Aufgabe

  Szenario: Status und Archiv überstehen einen echten Serverneustart
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Und "Website" ist in "Nordstern" archiviert
    Wenn ich den Server mit derselben SQLite-Datenbank neu starte
    Dann meldet der Server den Status "Archiviert" für "Website"
    Und die aktive Projektliste von "Nordstern" enthält "Website" nicht
    Und das Projektarchiv von "Nordstern" enthält "Website" genau einmal
    Und die bestehende Aufgabe "Startseite prüfen" bleibt unter derselben Kennung erreichbar
    Und ich kann im archivierten Projekt keine neue Aufgabe anlegen

  Szenario: Ein abgeschlossener Lauf bleibt als Verlauf nach Archivierung und Wiederherstellung erhalten
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Und ein Lauf für "Startseite prüfen" ist mit Vorbereitungsfehler beendet
    Wenn ich "Website" in "Nordstern" über HTTP archiviere
    Und ich "Website" in "Nordstern" über HTTP wiederherstelle
    Und ich den Server mit derselben SQLite-Datenbank neu starte
    Dann bleibt der abgeschlossene Lauf unter derselben Kennung im Verlauf von "Startseite prüfen" sichtbar

  Szenario: Ein öffentlicher Aktivitätsverlauf bleibt im Archiv und nach Neustart erhalten
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Und der öffentliche Aktivitätsverlauf von "Startseite prüfen" enthält den Anlageeintrag
    Wenn ich "Website" in "Nordstern" über HTTP archiviere
    Und ich den Server mit derselben SQLite-Datenbank neu starte
    Dann enthält der öffentliche Aktivitätsverlauf von "Startseite prüfen" denselben Anlageeintrag
    Wenn ich "Website" in "Nordstern" über HTTP wiederherstelle
    Dann enthält der öffentliche Aktivitätsverlauf von "Startseite prüfen" denselben Anlageeintrag

  Szenario: Archivierter Aufgabenverlauf bleibt auf die eigene Organisation begrenzt
    Angenommen die Organisationen "Nordstern" und "Südstern" bestehen
    Und das Projekt "Website" besteht in "Nordstern"
    Und der Agent "Mira" besteht in "Nordstern"
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Und der öffentliche Aktivitätsverlauf von "Startseite prüfen" enthält den Anlageeintrag
    Wenn ich "Website" in "Nordstern" über HTTP archiviere
    Dann zeigt der öffentliche Aktivitätsverlauf von "Südstern" keinen Eintrag von "Startseite prüfen"

  Szenario: Eine fremde Organisation darf weder archivieren noch wiederherstellen
    Angenommen die Organisationen "Nordstern" und "Südstern" bestehen
    Und das Projekt "Website" besteht in "Nordstern"
    Wenn ich "Website" über die Projekt-URL von "Südstern" zu archivieren versuche
    Dann erhalte ich eine datenfreie Ablehnung
    Und "Website" bleibt in "Nordstern" aktiv
    Wenn ich "Website" in "Nordstern" über HTTP archiviere
    Und ich "Website" über die Projekt-URL von "Südstern" wiederherzustellen versuche
    Dann erhalte ich eine datenfreie Ablehnung
    Und "Website" bleibt im Archiv von "Nordstern"

  Szenario: Eine ungültige Direkt-URL ändert kein Projekt
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" besteht
    Wenn ich ein unbekanntes Projekt über die Archiv-URL von "Nordstern" zu archivieren versuche
    Dann erhalte ich eine datenfreie Ablehnung
    Und "Website" bleibt in "Nordstern" aktiv

  Szenario: Ein aktiver Lauf verhindert die Archivierung
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Und ein Lauf für "Startseite prüfen" ist reserviert
    Wenn ich "Website" in "Nordstern" über HTTP archiviere
    Dann wird die Archivierung mit Hinweis auf den aktiven Lauf abgewiesen
    Und "Website" bleibt in "Nordstern" aktiv
    Und die bestehende Aufgabe "Startseite prüfen" bleibt unter derselben Kennung erreichbar

  Szenario: Ein archiviertes Projekt nimmt auch für eine bestehende Aufgabe keinen neuen Lauf an
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Und die Aufgabe "Startseite prüfen" besteht im Projekt "Website"
    Und "Website" ist in "Nordstern" archiviert
    Wenn ich über die öffentliche Lauf-Anwendungsgrenze einen Lauf für "Startseite prüfen" anfordere
    Dann wird die neue Laufreservierung verweigert
    Und der Verlauf von "Startseite prüfen" enthält keinen neuen Lauf

  Szenario: Gleichzeitiges Archivieren und Aufgabenanlage hinterlässt keinen Teilzustand
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Wenn ich "Website" archiviere und gleichzeitig die Aufgabe "Rennauftrag" über die öffentliche HTTP-Grenze anlege
    Dann ist "Website" genau einmal archiviert
    Und entweder ist "Rennauftrag" mit Aufgabe und beiden Aktivitätseinträgen vollständig angelegt oder mit Archivhinweis ganz abgewiesen
