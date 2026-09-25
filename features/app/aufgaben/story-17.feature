# language: de
@story-17 @task-01
Funktionalität: Aufgabe in einem Projekt genau einem Agenten zuweisen
  Als lokale Betreiberin
  möchte ich eine Aufgabe mit einem zuständigen Agenten in einem Projekt anlegen,
  damit Auftrag, Priorität und anfänglicher Status dauerhaft nachvollziehbar sind.

  Szenariogrundriss: Eine Aufgabe erhält genau einen Eino- oder Codex-CLI-Agenten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" mit dem Projekt "Website" besteht
    Und der Agent <Agent> mit der Ausführungsart <Ausfuehrungsart> besteht in "Nordstern"
    Wenn ich die Aufgabe "Startseite prüfen" mit der Beschreibung "Navigation und Texte prüfen", der Priorität "Hoch" und dem einzigen zuständigen Agenten <Agent> im Projekt "Website" über HTTP anlege
    Dann antwortet der Server mit einer neuen Aufgabenkennung und dem HTTP-Status 201
    Und die Aufgabenliste des Projekts "Website" enthält "Startseite prüfen" genau einmal
    Und das Aufgabendetail zeigt "Startseite prüfen", "Navigation und Texte prüfen", die Priorität "Hoch" und das Projekt "Website"
    Und das Aufgabendetail zeigt genau <Agent> als zuständigen Agenten mit der Ausführungsart <Ausfuehrungsart>
    Und das Aufgabendetail zeigt den anfänglichen Status "Offen"
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann zeigt dasselbe Aufgabendetail unter derselben Aufgabenkennung alle diese Angaben unverändert
    Und die Aufgabenliste des Projekts "Website" enthält "Startseite prüfen" weiterhin genau einmal

    Beispiele:
      | Agent  | Ausfuehrungsart |
      | "Mira" | "Eino"         |
      | "Kai"  | "Codex CLI"    |

  Szenariogrundriss: Ein fehlender Titel erzeugt keine Aufgabe
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" und dem Agenten "Mira" besteht
    Wenn ich eine Aufgabe mit dem Titel <Titel> und dem einzigen zuständigen Agenten "Mira" im Projekt "Website" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Aufgabenfeld "Titel"
    Und die Aufgabenliste des Projekts "Website" bleibt leer

    Beispiele:
      | Titel |
      | ""    |
      | "   " |

  Szenario: Eine Aufgabe ohne zuständigen Agenten wird abgewiesen
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" besteht
    Wenn ich die Aufgabe "Startseite prüfen" ohne zuständigen Agenten im Projekt "Website" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Aufgabenfeld "Zuständig"
    Und die Aufgabenliste des Projekts "Website" bleibt leer

  Szenario: Mehrere zuständige Agenten werden als eine Anfrage abgewiesen
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" besteht
    Und die Agenten "Mira" und "Kai" bestehen in "Nordstern"
    Wenn ich die Aufgabe "Startseite prüfen" mit "Mira" und "Kai" zugleich als zuständigen Agenten im Projekt "Website" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Aufgabenfeld "Zuständig"
    Und die Aufgabenliste des Projekts "Website" bleibt leer

  Szenariogrundriss: Ein unbekannter, fremder oder pausierter Agent erhält keine neue Aufgabe
    Angenommen die Organisation "Nordstern" mit dem Projekt "Website" besteht
    Und die Organisation "Südstern" besteht
    Und der Agent "Fremd" besteht in "Südstern"
    Und der Agent "Pause" besteht pausiert in "Nordstern"
    Wenn ich die Aufgabe "Startseite prüfen" mit <Agent> als einzigem zuständigen Agenten im Projekt "Website" über HTTP anlege
    Dann antwortet der Server mit einem Hinweis am Aufgabenfeld "Zuständig" ohne fremde Agentendaten
    Und die Aufgabenliste des Projekts "Website" bleibt leer
    Und die bestehenden Agenten bleiben unverändert

    Beispiele:
      | Agent                         |
      | einer unbekannten Agentenkennung |
      | "Fremd" aus "Südstern"       |
      | "Pause" aus "Nordstern"      |

  Szenariogrundriss: Ein unbekanntes oder organisationsfremdes Projekt nimmt keine Aufgabe an
    Angenommen die Organisationen "Nordstern" und "Südstern" bestehen
    Und das Projekt "Fremdprojekt" besteht in "Südstern"
    Und der Agent "Mira" besteht in "Nordstern"
    Wenn ich die Aufgabe "Startseite prüfen" mit "Mira" für <Projekt> in "Nordstern" über HTTP anlege
    Dann wird die Projektzuordnung ohne fremde Projektdaten abgewiesen
    Und in "Nordstern" entsteht keine Aufgabe
    Und die Aufgabenliste von "Fremdprojekt" bleibt unverändert

    Beispiele:
      | Projekt                                  |
      | eine unbekannte Projektkennung           |
      | das Projekt "Fremdprojekt" aus "Südstern" |
