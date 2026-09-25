# language: de
@story-37 @agt-04
Funktionalität: Persistenter Berichtsweg innerhalb einer Organisation
  Als lokale Betreiberin
  möchte ich Vorgesetzte für Agenten festlegen und das Organigramm prüfen,
  damit Berichtslinien sichtbar bleiben, ohne bestehende Arbeit umzuhängen.

  Grundlage:
    Angenommen ich starte für Story 37 einen lokalen Server mit neuer SQLite-Datenbank als Betreiberin
    Und die Organisationen "Nordstern" und "Südstern" bestehen
    Und die Eino-Agenten "Kai", "Mira" und "Lena" bestehen in "Nordstern"
    Und der Eino-Agent "Fremd" besteht in "Südstern"

  Szenario: Zwei Agenten derselben Ausführungsart bilden einen dauerhaften Berichtsweg
    Wenn ich "Kai" als Vorgesetzten von "Mira" in "Nordstern" über HTTP speichere
    Dann zeigt das Organigramm von "Nordstern" "Mira" unmittelbar unter "Kai"
    Und das Profil von "Mira" zeigt genau "Kai" als Vorgesetzten
    Und beide Agenten behalten die Ausführungsart "Eino"
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann zeigt das Organigramm von "Nordstern" denselben Berichtsweg genau einmal
    Und das Profil von "Mira" zeigt weiterhin genau "Kai" als Vorgesetzten

  Szenario: Umorganisation ersetzt nur den einen Vorgesetzten und bewahrt Aufgaben
    Angenommen "Mira" berichtet an "Kai" in "Nordstern"
    Und die Aufgabe "Startseite prüfen" ist "Mira" in "Nordstern" zugewiesen
    Wenn ich "Lena" als neue Vorgesetzte von "Mira" in "Nordstern" über HTTP speichere
    Dann zeigt das Organigramm "Mira" unmittelbar unter "Lena" und nicht mehr unter "Kai"
    Und das Profil von "Mira" zeigt genau "Lena" als Vorgesetzte
    Und die Aufgabe "Startseite prüfen" bleibt "Mira" zugewiesen
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann bleiben der neue Berichtsweg und die Zuweisung von "Startseite prüfen" erhalten

  Szenariogrundriss: Ungültige Berichtskante erzeugt keine Teiländerung
    Angenommen "Mira" berichtet an "Kai" in "Nordstern"
    Und "Lena" berichtet an "Mira" in "Nordstern"
    Und die Aufgabe "Startseite prüfen" ist "Mira" in "Nordstern" zugewiesen
    Wenn ich <Vorgesetzte> als Vorgesetzte von <Agent> in "Nordstern" über HTTP zu speichern versuche
    Dann wird die Berichtskante mit dem Grund <Grund> abgewiesen
    Und das Organigramm von "Nordstern" zeigt weiterhin "Kai" über "Mira" über "Lena"
    Und die Aufgabe "Startseite prüfen" bleibt "Mira" zugewiesen
    Und das Organigramm von "Südstern" bleibt unverändert

    Beispiele:
      | Vorgesetzte | Agent  | Grund                    |
      | "Mira"      | "Mira" | "Selbstbezug"            |
      | "Lena"      | "Kai"  | "Nachfahre und Zyklus"   |
      | "Fremd"     | "Mira" | "Agent nicht in dieser Organisation gefunden" |

  Szenario: Fremder Organisationskontext gibt keine Berichtsdaten preis
    Angenommen "Südstern" ist für die aktuelle serverseitige Identität nicht zugeordnet
    Wenn ich das Organigramm und das Agentenprofil von "Südstern" über HTTP abrufe
    Dann erhalte ich datenfreie Antworten wie für eine unbekannte Organisationskennung
    Wenn ich den Berichtsweg von "Fremd" über die direkte URL zu ändern versuche
    Dann erhalte ich eine datenfreie Ablehnung ohne Änderung des Organigramms von "Südstern"

  Szenario: Berichtslinie zeigt Kandidaten ohne Arbeitsfreigabe zu erteilen
    Angenommen "Mira" berichtet an "Kai" in "Nordstern"
    Und in "Nordstern" ist kein Delegationsübergang von "requested" nach "approved" freigegeben
    Wenn ich den Berichtsweg von "Kai" und die Arbeitsregeln von "Nordstern" über HTTP abrufe
    Dann ist "Mira" als Berichtsempfänger erkennbar
    Und der Delegationsübergang von "requested" nach "approved" ist weiterhin nicht freigegeben
