# language: de
@story-32 @goal-02
Funktionalität: Betreiberin pflegt Zielbaum und Status über HTTP
  Als lokale Betreiberin
  möchte ich Projektziele in einem Baum sehen und ihren Status ausdrücklich ändern,
  damit Projektzwecke und ihr tatsächlicher Stand nachvollziehbar bleiben.

  Szenario: Zwei Projekte zeigen ihre verschiedenen Zielpfade in derselben Organisation
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Technik" unter "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Artikel" mit dem Ziel "Redaktion" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Website" mit dem Ziel "Technik" in "Nordstern" über HTTP an
    Wenn ich den Zielbaum von "Nordstern" über HTTP abrufe
    Dann enthält er den Pfad "Wissensportal > Redaktion" für das Projekt "Artikel"
    Und er enthält den Pfad "Wissensportal > Technik" für das Projekt "Website"
    Und beide Teilziele haben "Wissensportal" als übergeordnetes Ziel

  Szenario: Eine ausdrückliche Aktion ändert nur den gewählten Zielstatus
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Technik" unter "Wissensportal" in "Nordstern" über HTTP an
    Wenn ich den Status von "Redaktion" in "Nordstern" über HTTP ausdrücklich auf "erreicht" setze
    Dann zeigt der Zielbaum "Redaktion" mit dem Status "erreicht"
    Und "Wissensportal" und "Technik" behalten ihren bisherigen Status
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann zeigt der Zielbaum "Redaktion" weiterhin mit dem Status "erreicht"
    Und "Wissensportal" und "Technik" haben weiterhin ihren bisherigen Status

  Szenario: Der Status des Stammziels wird nicht aus abgeschlossenen Kindern geraten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Technik" unter "Wissensportal" in "Nordstern" über HTTP an
    Wenn ich den Status von "Redaktion" in "Nordstern" über HTTP ausdrücklich auf "erreicht" setze
    Und ich den Status von "Technik" in "Nordstern" über HTTP ausdrücklich auf "erreicht" setze
    Dann behalten beide Teilziele den Status "erreicht"
    Und "Wissensportal" behält seinen bisherigen Status
    Wenn ich den Status von "Wissensportal" in "Nordstern" über HTTP ausdrücklich auf "erreicht" setze
    Dann zeigt der Zielbaum auch "Wissensportal" mit dem Status "erreicht"

  Szenariogrundriss: Ein leerer Teilzielname erzeugt keinen neuen Baumknoten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" über HTTP an
    Wenn ich das Teilziel mit dem Namen <Name> unter "Wissensportal" in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Zielfeld "Name"
    Und der Zielbaum von "Nordstern" enthält nur das unveränderte Stammziel "Wissensportal"

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Ein organisationsfremder Elternknoten wird ohne Teilerzeugung abgewiesen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisationen "Nordstern" und "Südstern" über HTTP an
    Und ich lege das Stammziel "Fremder Baum" in "Südstern" über HTTP an
    Wenn ich das Teilziel "Redaktion" unter "Fremder Baum" in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Zielfeld "Übergeordnetes Ziel"
    Und der Zielbaum von "Nordstern" bleibt leer
    Und der Zielbaum von "Südstern" enthält nur das unveränderte Stammziel "Fremder Baum"

  Szenario: Ein Projekt aus einer anderen Organisation erscheint nicht im Zielpfad
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisationen "Nordstern" und "Südstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Stammziel "Südportal" in "Südstern" über HTTP an
    Und ich lege das Projekt "Südprojekt" mit dem Ziel "Südportal" in "Südstern" über HTTP an
    Wenn ich den Zielbaum von "Nordstern" über HTTP abrufe
    Dann enthält er weder "Südportal" noch "Südprojekt"
    Und der Zielbaum von "Südstern" zeigt "Südprojekt" nur am Ziel "Südportal"

  Szenario: Eine nicht zugeordnete Betreiberin sieht und ändert keinen fremden Zielbaum
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" über HTTP an
    Wenn ich als nicht zugeordnete Betreiberin den Zielbaum von "Nordstern" über HTTP abrufe
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Wenn ich als nicht zugeordnete Betreiberin den Status von "Wissensportal" in "Nordstern" über HTTP ändere
    Dann antwortet der Server mit demselben datenfreien HTTP-Status 404
    Und der Status von "Wissensportal" bleibt unverändert
