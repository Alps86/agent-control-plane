# language: de
@story-14 @proj-01 @browser
Funktionalität: Projekt mit Zielbezug im Browser anlegen
  Als lokale Betreiberin
  möchte ich in meiner Organisation ein Projekt mit Zielbezug anlegen und wiederfinden,
  damit die spätere Arbeit einen sichtbaren Zweck erhält.

  Szenario: Von der Organisation zum Projektformular und zum leeren Projektdetail
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" an
    Wenn ich die Organisationsseite von "Nordstern" öffne
    Und ich die Aktion "Projekte" öffne
    Dann sehe ich die leere Projektübersicht von "Nordstern" mit der Aktion "Projekt anlegen"
    Wenn ich "Projekt anlegen" öffne
    Dann sehe ich die Felder "Name" und "Beschreibung"
    Und ich kann das Ziel "Wissensportal veröffentlichen" aus "Nordstern" auswählen
    Wenn ich als Projektnamen "Website" und als Beschreibung "Öffentlicher Auftritt" eingebe
    Und ich das Ziel "Wissensportal veröffentlichen" auswähle
    Und ich das Projektformular speichere
    Dann sehe ich "Website" mit "Öffentlicher Auftritt" in der Projektübersicht von "Nordstern"
    Wenn ich das Projektdetail von "Website" öffne
    Dann sehe ich die Zuordnung zum Ziel "Wissensportal veröffentlichen"
    Und ich sehe eine leere Aufgabenliste
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich "Website" mit "Öffentlicher Auftritt" und dem Ziel "Wissensportal veröffentlichen" erneut

  Szenariogrundriss: Ein leerer Projektname bleibt mit Feldhinweis im Formular
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" an
    Und ich öffne das Formular "Projekt anlegen" von "Nordstern"
    Wenn ich als Projektnamen <Name> eingebe
    Und ich das Ziel "Wissensportal veröffentlichen" auswähle
    Und ich das Projektformular speichere
    Dann sehe ich einen Hinweis am Projektfeld "Name"
    Und in der Projektübersicht von "Nordstern" erscheint kein Projekt

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Ein fehlendes Ziel bleibt mit Feldhinweis im Formular
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" an
    Und ich öffne das Formular "Projekt anlegen" von "Nordstern"
    Wenn ich als Projektnamen "Website" eingebe
    Und ich das Projektformular ohne Zielauswahl speichere
    Dann sehe ich einen Hinweis am Projektfeld "Ziel"
    Und in der Projektübersicht von "Nordstern" erscheint kein Projekt

  Szenario: Die Projektauswahl zeigt nur Ziele der eigenen Organisation
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisationen "Nordstern" und "Südstern" an
    Und ich lege das Stammziel "Fremdes Ziel" in "Südstern" an
    Wenn ich das Formular "Projekt anlegen" von "Nordstern" öffne
    Dann kann ich "Fremdes Ziel" nicht als Projektziel auswählen
    Und ich sehe, dass das Projektformular zu "Nordstern" gehört

  Szenario: Settings ist von der leeren Organisationsübersicht erreichbar
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Wenn ich in der leeren Organisationsübersicht "Settings" öffne
    Dann sehe ich die globale Seite "/settings" mit Codex-Abo und OpenRouter

  Szenario: Settings ist vom Organisationsdetail erreichbar
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Wenn ich die Organisationsseite von "Nordstern" öffne
    Und ich dort "Settings" öffne
    Dann sehe ich die globale Seite "/settings" mit Codex-Abo und OpenRouter
