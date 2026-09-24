# language: de
@story-15 @agt-01 @browser
Funktionalität: Eino-Agent im Browser aus einer Vorlage anlegen
  Als lokale Betreiberin
  möchte ich einen Agenten mit verständlichem Profil anlegen,
  damit ich seinen Auftrag und den noch fehlenden Modellzugang erkenne.

  Szenario: Von der Organisation zur Vorlage und zum neuen Agenten
    Angenommen ich öffne die Anwendung mit der Organisation "Nordstern" im Browser
    Wenn ich die Agentenübersicht von "Nordstern" öffne
    Dann sehe ich einen leeren Zustand mit der Aktion "Agent anlegen"
    Wenn ich "Agent anlegen" öffne
    Dann kann ich die Agentenvorlage "Recherche" auswählen
    Wenn ich die Vorlage "Recherche" auswähle
    Dann sehe ich vor dem Speichern Rolle, Anweisung und benannte Fachfähigkeiten der Vorlage
    Und die Ausführungsart ist "Eino"
    Wenn ich den Namen "Mira" eingebe und den Agenten speichere
    Dann sehe ich "Mira" in der Agentenübersicht von "Nordstern"
    Wenn ich das Profil von "Mira" öffne
    Dann sehe ich Rolle, Anweisung, benannte Fachfähigkeiten und "Eino"
    Und ich sehe "Modell nicht verbunden" sowie "Nicht startbereit"
    Und ich sehe keine direkte Shell-, Terminal-, Prozess-, Interpreter-, HTTP- oder Dateisystemfähigkeit
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich "Mira" mit derselben Profiladresse und demselben Auftrag erneut

  Szenario: Fehlender Modellzugang erklärt die Startsperre
    Angenommen ich öffne das Profil des unverbundenen Eino-Agenten "Mira" im Browser
    Wenn ich einen Start für "Mira" auslöse
    Dann sehe ich den Hinweis "Modellverbindung fehlt"
    Und ich sehe keine Startbestätigung und keine Laufadresse
    Und "Mira" bleibt "Nicht startbereit"

  Szenario: Aus einer Teamvorlage einen Agenten auswählen
    Angenommen ich öffne das Formular "Agent anlegen" für "Nordstern" im Browser
    Wenn ich die Teamvorlage "Rechercheteam" als Ausgangspunkt auswähle
    Dann sehe ich die Agentenvorlage "Recherche" mit Rolle und Auftrag als Vorschau
    Wenn ich daraus den Agenten "Mira" anlege
    Dann sehe ich nur "Mira" neu in der Agentenübersicht von "Nordstern"

  Szenariogrundriss: Ungültiger Name bleibt im Formular ohne neuen Agenten
    Angenommen ich öffne das Formular "Agent anlegen" für "Nordstern" im Browser
    Wenn ich die Vorlage "Recherche" auswähle
    Und ich als Agentennamen <Name> eingebe
    Und ich das Formular speichere
    Dann sehe ich einen Hinweis am Feld "Name"
    Und in der Agentenübersicht von "Nordstern" erscheint kein neuer Agent

    Beispiele:
      | Name  |
      | ""    |
      | "   " |

  Szenario: Bereits vergebener Name erhält einen verständlichen Feldhinweis
    Angenommen ich öffne die Anwendung mit der Organisation "Nordstern" im Browser
    Und der Agent "Mira" wurde in "Nordstern" aus der Vorlage "Recherche" angelegt
    Wenn ich im Formular "Agent anlegen" erneut die Vorlage "Recherche" und den Namen " mira " speichere
    Dann sehe ich einen Konflikthinweis am Feld "Name"
    Und die Agentenübersicht von "Nordstern" enthält "Mira" genau einmal
