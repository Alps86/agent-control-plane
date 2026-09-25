# language: de
@story-32 @goal-02 @browser
Funktionalität: Betreiberin pflegt Zielbaum und Status im Browser
  Als lokale Betreiberin
  möchte ich Zielpfade und Projekte sehen und den Zielstatus gezielt pflegen,
  damit ich den Fortschritt ohne angenommene Elternstatus ablesen kann.

  Szenario: Zwei Projektzielpfade sind im Zielbaum sichtbar und bleiben nach Neustart erhalten
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" an
    Wenn ich die Zielübersicht von "Nordstern" öffne
    Und ich das Teilziel "Redaktion" unter "Wissensportal" anlege
    Und ich das Teilziel "Technik" unter "Wissensportal" anlege
    Und ich das Projekt "Artikel" mit dem Ziel "Redaktion" in "Nordstern" anlege
    Und ich das Projekt "Website" mit dem Ziel "Technik" in "Nordstern" anlege
    Dann sehe ich den Zielpfad "Wissensportal > Redaktion" mit dem Projekt "Artikel"
    Und ich sehe den Zielpfad "Wissensportal > Technik" mit dem Projekt "Website"
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich beide Zielpfade und ihre Projektzuordnungen erneut

  Szenario: Zielstatus ändert sich nur über die ausdrückliche Statusaktion
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" an
    Wenn ich die Zielübersicht von "Nordstern" öffne
    Dann sehe ich den bisherigen Status von "Wissensportal" und "Redaktion"
    Wenn ich bei "Redaktion" die Statusaktion öffne und "erreicht" speichere
    Dann sehe ich "Redaktion" mit dem Status "erreicht"
    Und im Browser behält "Wissensportal" seinen bisherigen Status
    Wenn ich die Seite neu lade
    Dann sehe ich denselben Status beider Ziele

  Szenariogrundriss: Das Teilzielformular weist leere Namen am Feld zurück
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" an
    Wenn ich das Formular für ein Teilziel unter "Wissensportal" öffne
    Und ich als Teilzielname <Name> eingebe
    Und ich das Teilzielformular speichere
    Dann sehe ich einen Hinweis am Zielfeld "Name"
    Und unter "Wissensportal" erscheint kein Teilziel

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Der Zielbaum zeigt keine Ziele oder Projekte einer anderen Organisation
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisationen "Nordstern" und "Südstern" an
    Und ich lege das Stammziel "Südportal" in "Südstern" an
    Und ich lege das Projekt "Südprojekt" mit dem Ziel "Südportal" in "Südstern" an
    Wenn ich die Zielübersicht von "Nordstern" öffne
    Dann sehe ich weder "Südportal" noch "Südprojekt"
    Und ich sehe, dass der Zielbaum zu "Nordstern" gehört
