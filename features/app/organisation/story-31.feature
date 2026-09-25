# language: de
@story-31 @org-02
Funktionalität: Organisationskontext und getrennte Arbeitsregeln über öffentliche HTTP-Routen
  Als lokale Betreiberin
  möchte ich zwischen zugeordneten Organisationen wechseln und ihre Regeln getrennt pflegen,
  damit Suche, Details und spätere Arbeitsaktionen an die gewählte Organisation gebunden sind.

  Grundlage:
    Angenommen ich starte für Story 31 einen lokalen Server mit neuer SQLite-Datenbank
    Und ich lege über die öffentliche API die Organisationen "Nordstern" und "Südwind" an

  Szenario: Suche, Detail und Organisationswechsel halten den gewählten Kontext
    Wenn ich "Nord" in der Organisationsliste über HTTP suche
    Dann enthält die Suchantwort "Nordstern" und nicht "Südwind"
    Wenn ich über HTTP zu "Südwind" wechsle
    Dann ist "Südwind" der gewählte Organisationskontext
    Und die direkte Detail-URL von "Südwind" zeigt dessen Namen und Beschreibung
    Wenn ich über HTTP zu "Nordstern" wechsle
    Dann ist "Nordstern" der gewählte Organisationskontext
    Und die direkte Detail-URL von "Nordstern" zeigt dessen Namen und Beschreibung

  Szenario: Fremdreferenz und unbekannte Kennung verraten dieselben wenigen Daten
    Angenommen "Südwind" ist für die aktuelle serverseitige Identität nicht zugeordnet
    Wenn ich dessen direkte API-Detail-URL, HTML-Detail-URL und Wechselroute über HTTP abrufe
    Dann erhalte ich datenfreie Antworten ohne "Südwind" und dessen Beschreibung
    Und die Antworten sind nicht von einer unbekannten Organisationskennung unterscheidbar
    Und "Nordstern" bleibt der gewählte Organisationskontext

  Szenario: Aufgabenstatus, Zuweisung und Delegation sind getrennte Regelbereiche
    Wenn ich die Arbeitsregeln von "Nordstern" über HTTP abfrage
    Dann sehe ich getrennte Regeln für Aufgabenstatus, Zuweisung und Delegation mit zuständiger Freigabe
    Wenn ich für "Nordstern" den Aufgabenstatus-Übergang "todo" nach "in_progress" mit Freigabe "betreiber" vorschauprüfe und speichere
    Dann nennt die Vorschau betroffene Aufgabenstatus-Abläufe und keine Zuweisungs- oder Delegationsänderung
    Und die gespeicherte Aufgabenstatus-Regel ist nur bei "Nordstern" sichtbar
    Und Zuweisungs- und Delegationsregeln bleiben unverändert
    Und die Arbeitsregeln von "Südwind" bleiben unverändert

  Szenariogrundriss: Zulässige Zuweisungs- und Delegationsregeln benötigen ihre eigene Freigabe
    Wenn ich für "Nordstern" den <Bereich>-Übergang <Von> nach <Nach> mit Freigabe "betreiber" vorschauprüfe und speichere
    Dann nennt die Vorschau betroffene <Bereich>-Abläufe und keine Änderung anderer Regelbereiche
    Und die gespeicherte <Bereich>-Regel ist nur bei "Nordstern" sichtbar
    Und die Arbeitsregeln von "Südwind" bleiben unverändert

    Beispiele:
      | Bereich    | Von           | Nach         |
      | Zuweisung  | unassigned    | assigned     |
      | Delegation | requested     | approved     |

  Szenario: Eine gespeicherte Regel kann revisionsgebunden widerrufen werden
    Wenn ich für "Nordstern" den Aufgabenstatus-Übergang "todo" nach "in_progress" mit Freigabe "betreiber" vorschauprüfe und speichere
    Und ich die gespeicherte Aufgabenstatus-Regel für "Nordstern" über HTTP widerrufe
    Dann ist diese Regel bei "Nordstern" nicht mehr freigegeben
    Und Zuweisungs- und Delegationsregeln bleiben unverändert
    Und die Arbeitsregeln von "Südwind" bleiben unverändert
    Wenn ich den Widerruf mit einer veralteten Revision wiederhole
    Dann erhalte ich einen Revisionskonflikt ohne Regeländerung

  Szenariogrundriss: Ungültiger Übergang oder Rechteausweitung erzeugt keine Teiländerung
    Wenn ich für "Nordstern" eine <Art> Regeländerung über HTTP vorschauprüfe und zu speichern versuche
    Dann benennt die Antwort den Ablehnungsgrund ohne Regeländerung
    Und alle drei Regelbereiche von "Nordstern" bleiben unverändert
    Und die Arbeitsregeln von "Südwind" bleiben unverändert

    Beispiele:
      | Art                         |
      | unbekannter Aufgabenstatus  |
      | fehlende zuständige Freigabe|
      | Rechteausweitung            |
      | Workflow-DSL                |

  Szenario: Eine fremde Organisation kann keine Regeländerung erhalten
    Angenommen "Südwind" ist für die aktuelle serverseitige Identität nicht zugeordnet
    Wenn ich eine Delegationsregel für "Südwind" über eine direkte URL vorschauprüfe und zu speichern versuche
    Dann erhalte ich datenfreie Antworten wie für eine unbekannte Organisationskennung
    Und die Arbeitsregeln von "Südwind" bleiben unverändert
