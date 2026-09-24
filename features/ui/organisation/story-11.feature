# language: de
@story-11 @org-01 @browser
Funktionalität: Erste Organisation im Browser anlegen
  Als lokale Betreiberin
  möchte ich die leere Übersicht und ein verständliches Formular bedienen,
  damit die erste Organisation mit einem sicheren Arbeitsstandard entsteht.

  Szenario: Aus der leeren Übersicht eine Organisation anlegen
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Dann sehe ich eine leere Organisationsübersicht mit der Aktion "Organisation anlegen"
    Wenn ich "Organisation anlegen" öffne
    Dann sehe ich die Felder "Name" und "Beschreibung"
    Und ich sehe, dass Arbeits- und Delegationsübergänge zunächst restriktiv sind
    Wenn ich als Namen "Nordstern" und als Beschreibung "Wissensarbeit für das Team" eingebe
    Und ich das Formular speichere
    Dann sehe ich "Nordstern" und "Wissensarbeit für das Team" in der Organisationsübersicht
    Und ich sehe den restriktiven Standard für Arbeits- und Delegationsübergänge
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich "Nordstern" und "Wissensarbeit für das Team" erneut
    Und ich sehe weiterhin den restriktiven Standard für Arbeits- und Delegationsübergänge

  Szenariogrundriss: Ungültiger Name bleibt im Formular ohne neue Organisation
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich öffne das Formular "Organisation anlegen"
    Wenn ich als Namen <Name> und als Beschreibung "Nur Versuch" eingebe
    Und ich das Formular speichere
    Dann sehe ich einen Hinweis am Feld "Name"
    Und in der Organisationsübersicht erscheint keine neue Organisation

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Gleicher Name zeigt den Fehler am Feld und bewahrt die vorhandene Organisation
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege "Nordstern" mit der Beschreibung "Bestehend" an
    Wenn ich erneut "Organisation anlegen" öffne
    Und ich als Namen "  nordstern  " und als Beschreibung "Duplikat" eingebe
    Und ich das Formular speichere
    Dann sehe ich einen Konflikthinweis am Feld "Name"
    Und in der Organisationsübersicht steht "Nordstern" mit "Bestehend" genau einmal
    Und ich sehe dort keine Organisation mit der Beschreibung "Duplikat"
