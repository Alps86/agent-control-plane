# language: de
@story-17 @task-01 @browser
Funktionalität: Aufgabe im Projekt im Browser anlegen und wiederfinden
  Als lokale Betreiberin
  möchte ich eine Aufgabe aus einem Projekt heraus anlegen,
  damit ich Auftrag und Zuständigkeit sofort im Projektdetail sehe.

  Szenariogrundriss: Eino oder Codex CLI ist als einzelner Empfänger wählbar
    Angenommen ich öffne die Anwendung mit der Organisation "Nordstern" und dem Projekt "Website" im Browser
    Und der Agent <Agent> mit der Ausführungsart <Ausfuehrungsart> besteht in "Nordstern"
    Wenn ich das Projektdetail von "Website" öffne
    Dann sehe ich die leere Aufgabenliste und die Aktion "Aufgabe anlegen"
    Wenn ich "Aufgabe anlegen" öffne
    Dann sehe ich die Felder "Titel", "Beschreibung", "Priorität", "Projekt" und "Zuständig"
    Und "Website" ist als Projekt ausgewählt
    Wenn ich den Titel "Startseite prüfen" und die Beschreibung "Navigation und Texte prüfen" eingebe
    Und ich die Priorität "Hoch" und genau <Agent> als zuständigen Agenten auswähle
    Und ich das Aufgabenformular speichere
    Dann sehe ich "Startseite prüfen" genau einmal in der Aufgabenliste von "Website"
    Wenn ich das Aufgabendetail von "Startseite prüfen" öffne
    Dann sehe ich Beschreibung, Priorität "Hoch", Projekt "Website", <Agent> mit <Ausfuehrungsart> und Status "Offen"
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich dieselbe Aufgabe unter derselben Detailadresse mit denselben Angaben

    Beispiele:
      | Agent  | Ausfuehrungsart |
      | "Mira" | "Eino"         |
      | "Kai"  | "Codex CLI"    |

  Szenariogrundriss: Fehlender Titel oder Empfänger bleibt mit Feldhinweis im Formular
    Angenommen ich öffne das Formular "Aufgabe anlegen" für das Projekt "Website" in "Nordstern" im Browser
    Und der Agent "Mira" besteht in "Nordstern"
    Wenn ich die Aufgabe "Startseite prüfen" mit "Mira" als Empfänger vorbelege
    Und ich <Feld> im Aufgabenformular leere
    Und ich das Aufgabenformular speichere
    Dann sehe ich einen Hinweis am Aufgabenfeld <Feld>
    Und in der Aufgabenliste von "Website" erscheint keine neue Aufgabe

    Beispiele:
      | Feld        |
      | "Titel"     |
      | "Zuständig" |

  Szenario: Die Empfängerauswahl bleibt auf aktive Agenten der Organisation begrenzt
    Angenommen ich öffne das Formular "Aufgabe anlegen" für das Projekt "Website" in "Nordstern" im Browser
    Und der Agent "Mira" besteht in "Nordstern"
    Und der Agent "Fremd" besteht in "Südstern"
    Und der Agent "Pause" besteht pausiert in "Nordstern"
    Wenn ich die Auswahl "Zuständig" öffne
    Dann kann ich "Mira" genau einmal auswählen
    Und ich kann "Fremd" und "Pause" nicht auswählen
    Und ich kann für eine Aufgabe nicht zwei Agenten zugleich auswählen

  Szenario: Ein ungültiger Projektbezug erzeugt keinen Aufgabeneintrag
    Angenommen ich öffne das Formular "Aufgabe anlegen" für das Projekt "Website" in "Nordstern" im Browser
    Und der Agent "Mira" besteht in "Nordstern"
    Wenn ich eine Aufgabe mit dem Titel "Startseite prüfen" und dem Empfänger "Mira" für ein ungültiges Projekt speichere
    Dann sehe ich einen Hinweis am Aufgabenfeld "Projekt"
    Und in der Aufgabenliste von "Website" erscheint keine neue Aufgabe
