# language: de
@story-13 @goal-01 @browser
Funktionalität: Ein Stammziel in der Organisation im Browser anlegen
  Als lokale Betreiberin
  möchte ich aus meiner Organisation ein Ziel anlegen und wiederfinden,
  damit der erste Arbeitszweck sichtbar und dauerhaft ist.

  Szenario: Von der Organisation zur leeren Zielübersicht und zum neuen Stammziel
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Wenn ich die Organisationsseite von "Nordstern" öffne
    Und ich die Aktion "Ziele" öffne
    Dann sehe ich die leere Zielübersicht von "Nordstern" mit der Aktion "Ziel anlegen"
    Wenn ich "Ziel anlegen" öffne
    Dann sehe ich das Feld "Name" für ein Stammziel in "Nordstern"
    Wenn ich als Zielname "Wissensportal veröffentlichen" eingebe
    Und ich das Zielformular speichere
    Dann sehe ich "Wissensportal veröffentlichen" in der Zielübersicht von "Nordstern"
    Und ich sehe das Ziel als Stammziel ohne übergeordnetes Ziel
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich "Wissensportal veröffentlichen" erneut in der Zielübersicht von "Nordstern"

  Szenariogrundriss: Ein leerer Name bleibt mit Feldhinweis im Formular
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich öffne das Formular "Ziel anlegen" von "Nordstern"
    Wenn ich als Zielname <Name> eingebe
    Und ich das Zielformular speichere
    Dann sehe ich einen Hinweis am Feld "Name"
    Und in der Zielübersicht von "Nordstern" erscheint kein Ziel

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Ein Ziel erscheint nicht in der anderen Organisation
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisationen "Nordstern" und "Südstern" an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" an
    Wenn ich die Zielübersicht von "Südstern" öffne
    Dann sehe ich dort "Wissensportal veröffentlichen" nicht
    Und ich sehe, dass die Zielübersicht zu "Südstern" gehört
