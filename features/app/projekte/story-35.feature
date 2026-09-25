# language: de
@story-35 @proj-03
Funktionalität: Ausführungsort eines Projekts sicher binden
  Als lokale Betreiberin
  möchte ich einem Projekt einen überprüften Ausführungsort zuordnen,
  damit ein Lauf nur den freigegebenen privaten Projektbereich erhält.

  Grundlage:
    Angenommen ich starte die Anwendung mit einer neuen SQLite-Datenbank und privatem Datenverzeichnis
    Und ich lege die Organisation "Nordstern" und das Projekt "Website" über die öffentliche API an

  Szenario: Neues Projekt hat noch keinen Ausführungsort
    Wenn ich den Ausführungsort von "Website" über die öffentliche API abrufe
    Dann zeigt die Antwort configured=false, path=null und code=project_location_missing

  Szenario: Betreiberin aktiviert den privaten Projektort
    Wenn ich den privaten Ausführungsort von "Website" über die öffentliche API aktiviere
    Dann ist der Ausführungsort für "Website" aktiv und gehört nur zu "Nordstern"
    Und die öffentliche Startprüfung für "Website" bestätigt den freigegebenen privaten Ort als bereit
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte
    Dann bleibt der Ausführungsort von "Website" aktiv

  Szenario: Ein fehlender Ausführungsort blockiert den Projektstart
    Wenn ich die öffentliche Startprüfung für "Website" anfordere
    Dann wird die Startvorbereitung ohne Laufkennung wegen fehlendem Ausführungsort verweigert
    Und die Antwort erklärt, wie ich den privaten Projektort aktiviere

  Szenariogrundriss: Ein Client kann keinen beliebigen Pfad freigeben
    Wenn ich für "Website" den Ausführungsort mit <Eingabe> über die öffentliche API speichere
    Dann wird die Pfadfreigabe ohne Änderung des Ausführungsorts abgewiesen
    Und die öffentliche Startprüfung für "Website" erhält keinen fremden Pfad

    Beispiele:
      | Eingabe |
      | einem Home-Pfadfeld |
      | einer Repository-URL |
      | einem Pfad mit Aufwärtssegment |
      | einem unbekannten Feld |

  Szenario: Fremde Projektkennung gibt keinen Ausführungsort preis
    Angenommen ich lege die Organisation "Südstern" mit dem Projekt "Intern" über die öffentliche API an
    Wenn ich den Ausführungsort von "Intern" über "Nordstern" abrufe
    Dann erhalte ich eine datenfreie Verweigerung ohne Pfadangabe

  Szenario: Verletzte Integrität des privaten Projektorts blockiert den Start
    Angenommen ich habe den privaten Ausführungsort von "Website" über die öffentliche API aktiviert
    Und der private Ort von "Website" enthält einen Symlink nach außerhalb
    Wenn ich die öffentliche Startprüfung für "Website" anfordere
    Dann wird die Startvorbereitung wegen verletzter Pfadintegrität ohne Pfadübergabe abgewiesen

  Szenario: Ein Hardlink auf eine externe Datei blockiert die Startprüfung
    Angenommen ich habe den privaten Ausführungsort von "Website" über die öffentliche API aktiviert
    Und der private Ort von "Website" enthält einen Hardlink auf eine Datei außerhalb
    Wenn ich die öffentliche Startprüfung für "Website" anfordere
    Dann wird die Startvorbereitung wegen verletzter Pfadintegrität ohne Pfadübergabe abgewiesen
