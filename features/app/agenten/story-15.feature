# language: de
@story-15 @agt-01
Funktionalität: Eino-Agent aus einer Vorlage in einer Organisation anlegen
  Als lokale Betreiberin
  möchte ich einen Agenten ohne Entwicklerwissen aus einer Agentenvorlage anlegen,
  damit sein Auftrag und seine zulässigen Fachfähigkeiten sichtbar sind.

  Szenario: Vorlage erzeugt einen dauerhaft identifizierbaren Eino-Agenten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" besteht
    Wenn ich die Agentenvorlage "Recherche" für "Nordstern" über HTTP abrufe
    Dann zeigt die Vorlage eine Rolle, eine Anweisung und benannte Fachfähigkeiten
    Wenn ich den Agenten "Mira" aus der Vorlage "Recherche" für "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einer neuen Agentenkennung und dem HTTP-Status 201
    Und der Agent "Mira" gehört zu "Nordstern" und verwendet die Ausführungsart "Eino"
    Und die Agentendetails zeigen seine Rolle, Anweisung und benannten Fachfähigkeiten
    Und die Agentendetails zeigen den Modellstatus "nicht verbunden" und die Bereitschaft "nicht startbereit"
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann finde ich "Mira" unter derselben Agentenkennung in "Nordstern" genau einmal
    Und Rolle, Anweisung, Fachfähigkeiten und Bereitschaft sind unverändert sichtbar

  Szenario: Die neue Agentenkennung bleibt auf ihre Organisation begrenzt
    Angenommen die Organisationen "Nordstern" und "Suedstern" bestehen
    Und der Agent "Mira" wurde in "Nordstern" aus der Vorlage "Recherche" angelegt
    Wenn ich "Mira" über seine Agentenkennung in "Suedstern" per HTTP abrufe
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Und die Agentenliste von "Suedstern" enthält "Mira" nicht
    Wenn ich dieselbe Agentenkennung in "Nordstern" per HTTP abrufe
    Dann erhalte ich "Mira" mit seiner ursprünglichen Agentenkennung

  Szenario: Teamvorlage bietet einen einzelnen Agenten ohne Teilimport an
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich die Teamvorlage "Rechercheteam" für "Nordstern" über HTTP als Vorschau abrufe
    Dann sehe ich die auswählbare Agentenvorlage "Recherche" mit Rolle und Auftrag
    Wenn ich daraus nur den Agenten "Mira" für "Nordstern" über HTTP anlege
    Dann enthält die Agentenliste von "Nordstern" genau "Mira"
    Und es entstehen keine weiteren Agenten der Teamvorlage

  Szenario: Ohne Modellverbindung darf der angelegte Agent nicht starten
    Angenommen der Agent "Mira" wurde in "Nordstern" aus der Vorlage "Recherche" angelegt
    Und für "Mira" ist keine Modellverbindung zugeordnet
    Wenn ich einen Start für "Mira" über die öffentliche Anwendungsgrenze anfordere
    Dann wird der Start mit dem Grund "Modellverbindung fehlt" abgewiesen
    Und die Startantwort enthält keine Laufkennung und keine Startbestätigung
    Und "Mira" bleibt als "nicht startbereit" sichtbar

  Szenariogrundriss: Direkte Systemfähigkeiten lassen sich nicht in das Agentenprofil einschleusen
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich aus der Vorlage "Recherche" einen Eino-Agenten mit der zusätzlichen Fähigkeit <Faehigkeit> über HTTP anlege
    Dann wird die Fähigkeit als unzulässig abgewiesen
    Und in "Nordstern" entsteht kein Agent aus dieser Anfrage

    Beispiele:
      | Faehigkeit     |
      | "shell"       |
      | "terminal"    |
      | "prozess"     |
      | "interpreter" |
      | "http"        |
      | "dateisystem" |

  Szenariogrundriss: Ein leerer Agentenname wird am Feld abgewiesen
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich einen Agenten mit dem Namen <Name> aus der Vorlage "Recherche" für "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Feld "Name"
    Und die Agentenliste von "Nordstern" bleibt leer

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Eine unbekannte Vorlage erzeugt keinen Agenten
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich den Agenten "Mira" aus der unbekannten Vorlage "Fremdcode" für "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Feld "Vorlage"
    Und die Agentenliste von "Nordstern" bleibt leer

  Szenario: Ein zweiter Agent mit demselben Namen wird in der Organisation abgewiesen
    Angenommen die Organisation "Nordstern" besteht
    Und der Agent "Mira" wurde in "Nordstern" aus der Vorlage "Recherche" angelegt
    Wenn ich den Agenten " mira " erneut aus der Vorlage "Recherche" für "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einem Konflikthinweis am Feld "Name"
    Und die Agentenliste von "Nordstern" enthält "Mira" genau einmal

  Szenario: Ein fremder Organisationskontext kann keinen Agenten anlegen oder lesen
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich als nicht zugeordnete Betreiberin einen Agenten für "Nordstern" über HTTP anlege
    Dann verweigert der Server die Anlage ohne Teilerzeugung
    Wenn ich als nicht zugeordnete Betreiberin die Agentenliste von "Nordstern" über HTTP abrufe
    Dann erhalte ich keine Agentendaten

  Szenariogrundriss: Ein fremder Browser-Ursprung kann keinen Agenten anlegen
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich <Pfad> mit dem Agentennamen "Mira" und der Vorlage "Recherche" von Origin "https://fremd.example" per HTTP POST aufrufe
    Dann antwortet der Server mit dem HTTP-Status 403
    Und die Agentenliste von "Nordstern" bleibt leer

    Beispiele:
      | Pfad                                    |
      | "/api/organisationen/{id}/agenten"     |
      | "/organisationen/{id}/agenten"         |

  Szenario: Eine öffentliche Listener-Adresse verhindert den Agentenserverstart
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich denselben Server an einer öffentlichen Listener-Adresse neu starte
    Dann wird der Serverstart wegen der nicht lokalen APP_ADDR-Konfiguration abgewiesen
    Und auf dem Wildcard-Port ist kein Server erreichbar
    Wenn ich denselben Server mit derselben SQLite-Datenbank auf Loopback neu starte
    Dann bleibt die Agentenliste von "Nordstern" leer

  Szenario: Ein lokal gebundener Server nimmt einen gültigen Agenten-POST an
    Angenommen die Organisation "Nordstern" besteht
    Wenn ich den Agenten "Mira" aus der Vorlage "Recherche" für "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einer neuen Agentenkennung und dem HTTP-Status 201
    Und die Agentenliste von "Nordstern" enthält "Mira" genau einmal

  Szenario: Ein entfernter Peer erhält mit gefälschtem Localhost-Host keinen Schreibzugriff
    Angenommen die Organisation "Nordstern" besteht
    Wenn ein entfernter Peer am öffentlichen Agenten-Handler einen POST mit dem Host "localhost" sendet
    Dann antwortet der Server mit dem HTTP-Status 403
    Und die Agentenliste von "Nordstern" bleibt leer

  Szenario: Eine weitere numerische Loopback-Adresse erlaubt den Agenten-POST
    Angenommen die Organisation "Nordstern" besteht
    Und ich starte denselben Server an der Loopback-Adresse "127.0.0.2" neu
    Wenn ich den Agenten "Mira" aus der Vorlage "Recherche" für "Nordstern" über HTTP anlege
    Dann antwortet der Server mit einer neuen Agentenkennung und dem HTTP-Status 201
    Und die Agentenliste von "Nordstern" enthält "Mira" genau einmal
