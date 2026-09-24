# language: de
@story-11 @org-01
Funktionalität: Erste Organisation dauerhaft anlegen
  Als lokale Betreiberin
  möchte ich eine Organisation mit Name und Beschreibung anlegen,
  damit spätere Ziele, Projekte und Agenten einen eindeutigen Organisationskontext haben.

  Szenario: Leere Übersicht und neues Formular sind erreichbar
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich die Organisationsübersicht über HTTP abrufe
    Dann antwortet der Server erfolgreich mit einer leeren Organisationsliste
    Wenn ich das Formular für eine neue Organisation über HTTP abrufe
    Dann enthält das Formular die Felder "Name" und "Beschreibung"
    Und das Formular zeigt keine vorgegebene Organisation

  Szenario: Name und Beschreibung werden gespeichert und nach Neustart wieder gelesen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich die Organisation "Nordstern" mit der Beschreibung "Wissensarbeit für das Team" über HTTP anlege
    Dann antwortet der Server mit einer neuen Organisationskennung
    Und die Organisationsübersicht enthält "Nordstern" mit der Beschreibung "Wissensarbeit für das Team" genau einmal
    Und die Organisationsdetails zeigen einen restriktiven Standard für Arbeits- und Delegationsübergänge
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann enthält die Organisationsübersicht "Nordstern" mit der Beschreibung "Wissensarbeit für das Team" genau einmal
    Und die Organisationsdetails zeigen weiterhin den restriktiven Standard für Arbeits- und Delegationsübergänge

  Szenariogrundriss: Ein leerer Name erzeugt einen Feldhinweis und keine Organisation
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich eine Organisation mit dem Namen <Name> und der Beschreibung "Nur Versuch" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Feld "Name"
    Und die Organisationsübersicht bleibt leer

    Beispiele:
      | Name    |
      | ""      |
      | "   "   |

  Szenario: Ein bereits verwendeter Name wird als Feldkonflikt abgewiesen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" mit der Beschreibung "Bestehend" wurde über HTTP angelegt
    Wenn ich die Organisation "  nordstern  " mit der Beschreibung "Duplikat" über HTTP anlege
    Dann antwortet der Server mit einem Konflikthinweis am Feld "Name"
    Und die Organisationsübersicht enthält "Nordstern" mit der Beschreibung "Bestehend" genau einmal
    Und keine Organisation enthält die Beschreibung "Duplikat"

  Szenario: Ein unzugeordneter Aufruf darf keine Organisation anlegen oder lesen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" mit der Beschreibung "Bestehend" wurde über HTTP angelegt
    Wenn der Server für einen HTTP-Aufruf keine eindeutige Betreiberidentität ermittelt
    Dann verweigert er das Lesen der Organisationsübersicht ohne Organisationsdaten
    Und er verweigert das Anlegen einer Organisation ohne Teilerzeugung

  Szenario: Unbekannte und nicht zugeordnete Organisationskennung bleiben ununterscheidbar
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" mit der Beschreibung "Bestehend" wurde über HTTP angelegt
    Wenn ich eine unbekannte Organisationskennung über HTTP abfrage
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Wenn ich die vorhandene Organisationskennung als nicht zugeordnete Betreiberin über HTTP abfrage
    Dann antwortet der Server mit demselben datenfreien HTTP-Status 404

  Szenariogrundriss: Ein fremder Browser-Origin kann keine Organisation anlegen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich <Pfad> mit dem Namen "Fremder Ursprung" und der Beschreibung "Unerlaubt" von Origin "https://fremd.example" per HTTP POST aufrufe
    Dann antwortet der Server mit dem HTTP-Status 403
    Und die Organisationsübersicht bleibt leer

    Beispiele:
      | Pfad                  |
      | "/api/organisationen" |
      | "/organisationen"     |

  Szenariogrundriss: Ein unerlaubter Host kann keine Organisation anlegen
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich <Pfad> mit dem Namen "Fremder Host" und der Beschreibung "Unerlaubt" über Host "fremd.example" per HTTP POST aufrufe
    Dann antwortet der Server mit dem HTTP-Status 403
    Und die Organisationsübersicht bleibt leer

    Beispiele:
      | Pfad                  |
      | "/api/organisationen" |
      | "/organisationen"     |
