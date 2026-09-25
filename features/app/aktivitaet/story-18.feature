# language: de
@story-18 @act-01
Funktionalität: Aktivitätsverlauf einer Organisation
  Als lokale Betreiberin
  möchte ich die Anlage und erste Zuweisung einer Aufgabe nachvollziehen,
  damit ich Akteur, Quelle, Zeitpunkt und betroffenes Objekt wiederfinde.

  Szenario: Aufgabenanlage erzeugt zwei dauerhaft verlinkte Ereignisse
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Wenn ich "Startseite prüfen" im Projekt "Website" dem Agenten "Mira" über die öffentliche API zuweise
    Dann ist die Aufgabe "Startseite prüfen" angelegt und "Mira" zugewiesen
    Und der Aktivitätsverlauf von "Nordstern" enthält genau ein Anlageereignis für "Startseite prüfen"
    Und der Aktivitätsverlauf von "Nordstern" enthält genau ein Zuweisungsereignis an "Mira" für "Startseite prüfen"
    Und beide Ereignisse nennen die serverseitige Betreiberkennung und "api" als Quelle
    Und beide Ereignisse haben einen gültigen Zeitpunkt und einen Deep Link zur gespeicherten Aufgabe
    Wenn ich den Server beende und mit derselben SQLite-Datenbank neu starte
    Dann enthält der Aktivitätsverlauf von "Nordstern" dieselben zwei Ereignisse mit denselben Kennungen und Zeitpunkten
    Und beide Deep Links öffnen die ursprüngliche Aufgabe

  Szenario: Ereignisse bleiben auf ihre Organisation begrenzt
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Und die Organisation "Südstern" besteht
    Wenn ich "Startseite prüfen" im Projekt "Website" dem Agenten "Mira" über die öffentliche API zuweise
    Dann enthält der Aktivitätsverlauf von "Südstern" keine Ereignisse aus "Nordstern"
    Und keine Kennung eines "Nordstern"-Ereignisses erscheint im Verlauf von "Südstern"

  Szenario: Abgewiesene Aufgabenanlage erzeugt keine Aktivität
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Wenn ich eine Aufgabe ohne Titel im Projekt "Website" dem Agenten "Mira" über die öffentliche API zuweise
    Dann wird die Aufgabenanlage mit einem Feldfehler abgewiesen
    Und im Aktivitätsverlauf von "Nordstern" steht kein Aufgabenereignis

  Szenario: Clientwerte können Akteur und Quelle nicht fälschen
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Wenn ich eine Aufgabenanlage mit den behaupteten Werten "Fremdagent" und "browser" für Akteur und Quelle sende
    Dann wird die Aufgabenanlage ohne neues Ereignis abgewiesen

  Szenario: Formularanlage nennt den tatsächlichen Browserpfad als Quelle
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Wenn ich "Startseite prüfen" über das öffentliche Aufgabenformular für "Website" anlege
    Dann enthält der Aktivitätsverlauf von "Nordstern" genau zwei Ereignisse mit Quelle "browser"

  Szenario: Fehler beim zweiten Ereignis rollt Aufgabe und erste Aktivität zurück
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Und der zweite Ereignisschreibvorgang schlägt kontrolliert fehl
    Wenn ich "Startseite prüfen" im Projekt "Website" dem Agenten "Mira" über die öffentliche API zuweise und der Schreibfehler eintritt
    Dann meldet die Aufgaben-API einen Serverfehler
    Und die Aufgaben-API findet die betroffene Aufgabenkennung nicht
    Und der Aktivitätsverlauf von "Nordstern" enthält kein Ereignis
