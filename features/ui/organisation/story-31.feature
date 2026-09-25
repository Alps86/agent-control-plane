# language: de
@story-31 @org-02 @browser
Funktionalität: Organisationen wechseln und Arbeitsregeln sichtbar prüfen
  Als lokale Betreiberin
  möchte ich Organisationssuche, Wechsel und Regelvorschau im Browser bedienen,
  damit ich die Wirkung vor dem Speichern erkenne.

  Grundlage:
    Angenommen ich öffne Story 31 mit den Organisationen "Nordstern" und "Südwind" im Browser

  Szenario: Vom regulären Einstieg sind Organisationswahl und Arbeitsregeln erreichbar
    Wenn ich den regulären Einstieg der Anwendung öffne
    Dann sehe ich die Aktion "Organisation wechseln"
    Wenn ich die Organisationswahl über diese Aktion öffne
    Und ich "Nordstern" auswähle
    Dann sehe ich im Organisationsdetail die Aktion "Arbeitsregeln bearbeiten"
    Wenn ich die Arbeitsregeln über diese Aktion öffne
    Dann sehe ich getrennte Bereiche für Aufgabenstatus, Zuweisung und Delegation

  Szenario: Die Suche findet eine Organisation und der Wechsel öffnet ihr Detail
    Wenn ich in der Organisationsübersicht nach "Süd" suche
    Dann sehe ich "Südwind" und nicht "Nordstern" in den Suchergebnissen
    Wenn ich "Südwind" auswähle
    Dann zeigt die Oberfläche "Südwind" als gewählte Organisation
    Und ich sehe Namen und Beschreibung im Detail
    Wenn ich die Detail-URL neu lade
    Dann bleibt "Südwind" als gewählte Organisation sichtbar

  Szenario: Der direkte Aufruf einer fremden Organisation zeigt keine Inhalte
    Angenommen "Südwind" ist für die aktuelle serverseitige Identität nicht zugeordnet
    Wenn ich die Detail-URL von "Südwind" direkt im Browser öffne
    Dann sehe ich keine Daten von "Südwind"
    Und die Oberfläche zeigt denselben datenfreien Zustand wie für eine unbekannte Kennung

  Szenario: Regelvorschau erklärt die drei getrennten Bereiche vor dem Speichern
    Wenn ich die Arbeitsregeln von "Nordstern" öffne
    Dann sehe ich getrennte Bereiche für Aufgabenstatus, Zuweisung und Delegation
    Wenn ich einen zulässigen Aufgabenstatus-Übergang und die zuständige Freigabe auswähle
    Und ich die Vorschau öffne
    Dann sehe ich betroffene Aufgabenstatus-Abläufe und keine Änderung der übrigen Bereiche
    Wenn ich die Regel speichere und die Seite neu lade
    Dann sehe ich den gespeicherten Übergang bei "Nordstern"
    Und bei "Südwind" bleibt die bisherige Regel sichtbar

  Szenariogrundriss: Ungültige Regel und Rechteausweitung sind sichtbar abgewiesen
    Wenn ich bei "Nordstern" <Art> vorschauprüfe
    Dann sehe ich einen konkreten Ablehnungsgrund
    Und ich kann die abgewiesene Regel nicht speichern
    Und nach erneutem Laden bleiben die bisherigen Regeln sichtbar

    Beispiele:
      | Art                  |
      | einen ungültigen Übergang |
      | eine Rechteausweitung     |

  Szenario: Eine gespeicherte Regel wird mit sichtbarer Wirkung widerrufen
    Wenn ich die Arbeitsregeln von "Nordstern" öffne
    Und ich einen zulässigen Aufgabenstatus-Übergang und die zuständige Freigabe auswähle
    Und ich die Vorschau öffne
    Und ich die Regel speichere und die Seite neu lade
    Wenn ich den gespeicherten Übergang widerrufe
    Dann sehe ich den Übergang bei "Nordstern" nicht mehr als freigegeben
    Und die Bereiche Zuweisung und Delegation bleiben sichtbar unverändert
    Und bei "Südwind" bleibt die bisherige Regel sichtbar
