# language: de
@story-37 @agt-04 @browser
Funktionalität: Organigramm und Berichtsweg im Browser bedienen
  Als lokale Betreiberin
  möchte ich Berichtslinien aus der Organisation und dem Agentenprofil erreichen,
  damit ich eine Änderung und ihre Wirkung unmittelbar prüfen kann.

  Grundlage:
    Angenommen ich öffne Story 37 mit "Nordstern", den Eino-Agenten "Kai", "Mira" und "Lena" im Browser

  Szenario: Organigramm ist vom Organisationseinstieg erreichbar
    Wenn ich das Detail von "Nordstern" vom regulären Einstieg aus öffne
    Dann sehe ich die Aktion "Organigramm"
    Wenn ich das Organigramm darüber öffne
    Dann sehe ich "Kai", "Mira" und "Lena" mit ihren Berichtslinien oder als Agenten ohne Vorgesetzten
    Wenn ich "Mira" auswähle
    Dann erreiche ich ihr Agentenprofil mit dem sichtbaren Vorgesetzten

  Szenario: Vorgesetztenwechsel ist im Organigramm und Agentenprofil nach Neuladen sichtbar
    Angenommen "Mira" berichtet an "Kai" in "Nordstern"
    Und die Aufgabe "Startseite prüfen" ist "Mira" zugewiesen
    Wenn ich im Profil von "Mira" den Berichtsweg bearbeite
    Und ich "Lena" als neue Vorgesetzte speichere
    Dann sehe ich "Lena" als einzige Vorgesetzte von "Mira"
    Und das Organigramm zeigt "Mira" unmittelbar unter "Lena" und nicht mehr unter "Kai"
    Und die Aufgabe "Startseite prüfen" zeigt weiterhin "Mira" als zuständigen Agenten
    Wenn ich die Seite neu lade
    Dann bleiben der neue Berichtsweg und die Aufgabenzuweisung sichtbar

  Szenariogrundriss: Ungültiger Vorgesetzter wird verständlich abgewiesen
    Angenommen "Mira" berichtet an "Kai" in "Nordstern"
    Und "Lena" berichtet an "Mira" in "Nordstern"
    Wenn ich für <Agent> <Vorgesetzte> als Vorgesetzte zu speichern versuche
    Dann sehe ich den Ablehnungsgrund <Grund> im Berichtswegformular
    Und das Organigramm zeigt weiterhin "Kai" über "Mira" über "Lena"

    Beispiele:
      | Agent  | Vorgesetzte | Grund                  |
      | "Mira" | "Mira"      | "Selbstbezug"          |
      | "Kai"  | "Lena"      | "Nachfahre und Zyklus" |

  Szenario: Berichtslinie und Arbeitsfreigabe werden getrennt erklärt
    Angenommen "Mira" berichtet an "Kai" in "Nordstern"
    Und in "Nordstern" ist kein Delegationsübergang von "requested" nach "approved" freigegeben
    Wenn ich den Berichtsweg von "Mira" im Organigramm öffne
    Dann sehe ich "Kai" als Vorgesetzten von "Mira"
    Wenn ich die Arbeitsregeln von "Nordstern" öffne
    Dann sehe ich den Delegationsübergang von "requested" nach "approved" weiterhin nicht als freigegeben
