# language: de
@story-03 @arc-03
Funktionalität: Organisationsressourcen an der öffentlichen Rechtegrenze lesen
  Als Betreiber oder zugeordneter Agent
  möchte ich nur die mir erlaubte Beispielressource lesen,
  damit eine fremde oder ungeklärte Zuordnung keine Daten preisgibt.

  Grundlage:
    Angenommen der Go-Server kennt die Beispielressource "res-eigen" mit dem Namen "Eigene Ressource" in Organisation "org-eigen"
    Und der Go-Server kennt die Beispielressource "res-fremd" mit dem Namen "Fremde Ressource" in Organisation "org-fremd"

  Szenario: Betreiber liest eine Ressource seiner Organisation
    Angenommen der Server ermittelt eindeutig den Betreiber "betreiber-a" für Organisation "org-eigen"
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-eigen" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 200
    Und die JSON-Antwort enthält die Ressource "res-eigen" mit dem Namen "Eigene Ressource" und der Organisation "org-eigen"

  Szenario: Zugeordneter Agent liest seine Ressource
    Angenommen der Server ermittelt eindeutig den Agenten "agent-a" für Organisation "org-eigen" mit Zuordnung zur Ressource "res-eigen"
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-eigen" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 200
    Und die JSON-Antwort enthält die Ressource "res-eigen" mit dem Namen "Eigene Ressource" und der Organisation "org-eigen"

  Szenario: Eine fremde Ressource ist von einer unbekannten nicht zu unterscheiden
    Angenommen der Server ermittelt eindeutig den Betreiber "betreiber-a" für Organisation "org-eigen"
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-fremd" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 404
    Und der JSON-Fehler ist genau "resource_not_found" ohne Ressourcendaten
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-unbekannt" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 404
    Und Status und Antwortkörper stimmen mit der vorherigen Antwort überein

  Szenario: Gefälschte Identität in Headern und Query erweitert keine Rechte
    Angenommen der Server ermittelt eindeutig den Betreiber "betreiber-a" für Organisation "org-eigen"
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-fremd" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 404
    Und der JSON-Fehler ist genau "resource_not_found" ohne Ressourcendaten
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-fremd?actor=betreiber-fremd&organization=org-fremd" mit diesen Headern an den Server sende:
      | X-Actor-ID        | betreiber-fremd |
      | X-Organization-ID | org-fremd       |
    Dann antwortet der Server mit dem HTTP-Status 404
    Und Status und Antwortkörper stimmen mit der vorherigen Antwort überein

  Szenario: Ein Agent ohne Ressourcenzuordnung erhält keine Daten
    Angenommen der Server ermittelt eindeutig den Agenten "agent-a" für Organisation "org-eigen" ohne Ressourcenzuordnung
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-eigen" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 404
    Und der JSON-Fehler ist genau "resource_not_found" ohne Ressourcendaten

  Szenariogrundriss: Fehlende oder mehrdeutige Identität und Zuordnung verweigern das Lesen
    Angenommen der Server hat <Kontext>
    Wenn ich über HTTP GET "/api/architektur/rechte/ressourcen/res-eigen" an den Server sende
    Dann antwortet der Server mit dem HTTP-Status 403
    Und der JSON-Fehler ist genau "access_denied" ohne Ressourcendaten

    Beispiele:
      | Kontext |
      | keine Akteuridentität |
      | zwei Akteuridentitäten |
      | einen eindeutigen Betreiber ohne Organisationszuordnung |
      | einen eindeutigen Betreiber mit zwei Organisationszuordnungen |
