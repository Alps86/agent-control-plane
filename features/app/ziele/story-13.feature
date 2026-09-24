# language: de
@story-13 @goal-01
Funktionalität: Stammziele einer Organisation dauerhaft anlegen und auflisten
  Als lokale Betreiberin
  möchte ich ein Stammziel in meiner Organisation anlegen,
  damit Projekte später einem ausdrücklichen Zweck zugeordnet werden können.

  Szenario: Eine Organisation beginnt mit einer leeren Zielübersicht
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn ich die Zielübersicht von "Nordstern" über HTTP abrufe
    Dann antwortet der Server erfolgreich mit einer leeren Zielliste

  Szenario: Ein Stammziel wird in seiner Organisation gespeichert
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn ich das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einer neuen Zielkennung
    Und die Zielübersicht von "Nordstern" enthält "Wissensportal veröffentlichen" genau einmal
    Und das Ziel hat keinen übergeordneten Ziel- oder Projektbezug

  Szenario: Das Stammziel bleibt nach einem Serverneustart erhalten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann enthält die Zielübersicht von "Nordstern" "Wissensportal veröffentlichen" genau einmal
    Und das Ziel hat dieselbe Kennung und keinen übergeordneten Ziel- oder Projektbezug

  Szenariogrundriss: Ein leerer Zielname erzeugt keine Teildaten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn ich ein Stammziel mit dem Namen <Name> in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Feld "Name"
    Und die Zielübersicht von "Nordstern" bleibt leer
    Und die Organisation "Nordstern" bleibt unverändert erhalten

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Jede Organisation sieht nur ihre eigenen Ziele
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisationen "Nordstern" und "Südstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich die Zielübersicht von "Südstern" über HTTP abrufe
    Dann antwortet der Server erfolgreich mit einer leeren Zielliste
    Und die Zielübersicht von "Nordstern" enthält "Wissensportal veröffentlichen" genau einmal

  Szenario: Eine unbekannte Organisation nimmt kein Ziel an
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich das Stammziel "Unerlaubtes Ziel" mit einer unbekannten Organisationskennung über HTTP anlege
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Und es wird kein Ziel angelegt

  Szenario: Ohne eindeutige Betreiberidentität werden Ziele weder gelesen noch angelegt
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn der Server für einen HTTP-Aufruf keine eindeutige Betreiberidentität ermittelt
    Dann verweigert er das Lesen der Zielübersicht von "Nordstern" ohne Zieldaten
    Und er verweigert das Anlegen eines Stammziels in "Nordstern" ohne Teilerzeugung

  Szenario: Eine nicht zugeordnete Organisation bleibt auch für Ziele verborgen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn ich als nicht zugeordnete Betreiberin die Zielübersicht von "Nordstern" über HTTP abrufe
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Wenn ich als nicht zugeordnete Betreiberin ein Stammziel in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit demselben datenfreien HTTP-Status 404
    Und es wird kein Ziel angelegt

  Szenariogrundriss: Ein fremder Ursprung oder Host kann kein Ziel anlegen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn ich <Angriff> ein Stammziel "Fremdes Ziel" in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit dem HTTP-Status 403
    Und die Zielübersicht von "Nordstern" bleibt leer

    Beispiele:
      | Angriff                                 |
      | von Origin "https://fremd.example"       |
      | über Host "fremd.example"                 |

  Szenario: Ein Wildcard-Listener erlaubt auch mit gefälschtem localhost-Host keinen Zielschreibzugriff
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn ich den Server beende und mit derselben SQLite-Datenbank auf der Wildcard-Adresse "0.0.0.0" erneut starte
    Und ein entfernter Client das Stammziel "Fremdes Ziel" in "Nordstern" mit dem gefälschten HTTP-Host "localhost" samt tatsächlichem Listener-Port anlegt
    Dann antwortet der Server mit dem HTTP-Status 403
    Wenn ein lokaler Client das Stammziel "Lokales Ziel" in "Nordstern" über den Wildcard-Listener anlegt
    Dann antwortet der Server ebenfalls mit dem HTTP-Status 403
    Wenn ich den Server beende und mit derselben SQLite-Datenbank auf der Loopback-Adresse erneut starte
    Dann bleibt die Zielübersicht von "Nordstern" leer
