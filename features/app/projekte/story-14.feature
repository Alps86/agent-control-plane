# language: de
@story-14 @proj-01
Funktionalität: Projekt mit Zielbezug in einer Organisation anlegen
  Als lokale Betreiberin
  möchte ich ein Projekt mit Beschreibung einem Ziel meiner Organisation zuordnen,
  damit der Zweck der späteren Projektarbeit dauerhaft sichtbar ist.

  Szenario: Eine Organisation beginnt mit einer leeren Projektübersicht
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich die Projektübersicht von "Nordstern" über HTTP abrufe
    Dann antwortet der Server erfolgreich mit einer leeren Projektliste
    Wenn ich das Projektformular von "Nordstern" über HTTP abrufe
    Dann bietet es das Ziel "Wissensportal veröffentlichen" zur Auswahl an

  Szenario: Name, Beschreibung und Zielzuordnung werden gespeichert und im Detail sichtbar
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich das Projekt "Website" mit der Beschreibung "Öffentlicher Auftritt" und dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einer neuen Projektkennung
    Und die Projektübersicht von "Nordstern" enthält "Website" mit der Beschreibung "Öffentlicher Auftritt" genau einmal
    Und das Projektdetail von "Website" zeigt das Ziel "Wissensportal veröffentlichen" aus "Nordstern"
    Und das Projektdetail von "Website" zeigt eine leere Aufgabenliste

  Szenario: Projekt und Zielbezug bleiben nach einem Serverneustart erhalten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Website" mit der Beschreibung "Öffentlicher Auftritt" und dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann enthält die Projektübersicht von "Nordstern" "Website" mit der Beschreibung "Öffentlicher Auftritt" genau einmal
    Und das Projektdetail von "Website" zeigt dieselbe Projektkennung und dasselbe Ziel "Wissensportal veröffentlichen"
    Und das Projektdetail von "Website" zeigt eine leere Aufgabenliste

  Szenariogrundriss: Ein leerer Projektname erzeugt keine Teildaten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich ein Projekt mit dem Namen <Name> und dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Projektfeld "Name"
    Und die Projektübersicht von "Nordstern" bleibt leer
    Und das Ziel "Wissensportal veröffentlichen" bleibt unverändert erhalten

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Ohne Zielzuordnung entsteht kein Projekt
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Wenn ich das Projekt "Website" ohne Zielzuordnung in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Projektfeld "Ziel"
    Und die Projektübersicht von "Nordstern" bleibt leer

  Szenariogrundriss: Ein unbekanntes oder organisationsfremdes Ziel wird abgewiesen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisationen "Nordstern" und "Südstern" über HTTP an
    Und ich lege das Stammziel "Fremdes Ziel" in "Südstern" über HTTP an
    Wenn ich das Projekt "Website" mit <Ziel> in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Projektfeld "Ziel"
    Und die Projektübersicht von "Nordstern" bleibt leer
    Und das Ziel "Fremdes Ziel" bleibt unverändert in "Südstern" erhalten

    Beispiele:
      | Ziel                                      |
      | einer unbekannten Zielkennung             |
      | dem Ziel "Fremdes Ziel" aus "Südstern"    |

  Szenario: Projekte erscheinen nur in ihrer Organisation
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisationen "Nordstern" und "Südstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Website" mit der Beschreibung "Öffentlicher Auftritt" und dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich die Projektübersicht von "Südstern" über HTTP abrufe
    Dann antwortet der Server erfolgreich mit einer leeren Projektliste
    Und die Projektübersicht von "Nordstern" enthält "Website" genau einmal

  Szenario: Eine unbekannte Organisation nimmt kein Projekt an
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich das Projekt "Website" mit einer unbekannten Organisationskennung über HTTP anlege
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Und es wird kein Projekt angelegt

  Szenario: Ohne eindeutige Betreiberidentität werden Projekte weder gelesen noch angelegt
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn der Server für einen HTTP-Aufruf keine eindeutige Betreiberidentität ermittelt
    Dann verweigert er das Lesen der Projektübersicht von "Nordstern" ohne Projektdaten
    Und er verweigert das Anlegen eines Projekts in "Nordstern" ohne Teilerzeugung

  Szenario: Eine nicht zugeordnete Organisation bleibt auch für Projekte verborgen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Website" mit der Beschreibung "Öffentlicher Auftritt" und dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich als nicht zugeordnete Betreiberin die Projektübersicht von "Nordstern" über HTTP abrufe
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Wenn ich als nicht zugeordnete Betreiberin das Projektdetail von "Website" über HTTP abrufe
    Dann antwortet der Server mit demselben datenfreien HTTP-Status 404
    Und als nicht zugeordnete Betreiberin kann ich in "Nordstern" kein Projekt anlegen

  Szenariogrundriss: Ein fremder Ursprung oder Host kann kein Projekt anlegen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich <Angriff> das Projekt "Fremdes Projekt" mit dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP anlege
    Dann antwortet der Server mit dem HTTP-Status 403
    Und die Projektübersicht von "Nordstern" bleibt leer

    Beispiele:
      | Angriff                                 |
      | von Origin "https://fremd.example"     |
      | über Host "fremd.example"               |

  Szenariogrundriss: Unsichere APP_ADDR verhindert den Serverstart vor dem Listener
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Website" mit der Beschreibung "Öffentlicher Auftritt" und dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich den Server mit derselben SQLite-Datenbank und APP_ADDR <Adresse> neu starte
    Dann beendet sich der Server vor dem Listener mit einem konkreten APP_ADDR-Loopback-Fehler
    Und die abgewiesene Bindadresse nimmt keine HTTP-Verbindung an
    Wenn ich den Server beende und mit derselben SQLite-Datenbank auf der Loopback-Adresse erneut starte
    Dann enthält die Projektübersicht von "Nordstern" "Website" mit der Beschreibung "Öffentlicher Auftritt" genau einmal
    Und das Projektdetail von "Website" zeigt dieselbe Projektkennung und dasselbe Ziel "Wissensportal veröffentlichen"

    Beispiele:
      | Adresse                  |
      | "0.0.0.0:<Port>"         |
      | "[::]:<Port>"            |
      | "192.0.2.1:<Port>"      |
      | "unbekannt:<Port>"      |
      | "127.0.0.1"             |

  Szenario: Der kanonische Server liefert Organisation, Ziel, Agent, Projekt und Settings aus
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Website" mit der Beschreibung "Öffentlicher Auftritt" und dem Ziel "Wissensportal veröffentlichen" in "Nordstern" über HTTP an
    Wenn ich die öffentlichen Organisations-, Ziel-, Agenten-, Projekt- und Settings-Seiten von "Nordstern" über HTTP abrufe
    Dann antwortet jede dieser Seiten mit HTTP 200
    Und die Organisationsübersicht verlinkt nach "/settings"
    Und die Organisationsdetails verlinken nach "/settings"
