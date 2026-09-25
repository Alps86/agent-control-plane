# language: de
@story-33 @goal-03 @browser
Funktionalität: Betreiberin verschiebt Zielzweige im Browser
  Als Betreiberin möchte ich Ziele in der Zielübersicht neu einordnen,
  damit Projektpfade, Status und die neue Elternziel-Kante sichtbar bleiben.

  Szenario: Zielzweig verschieben und nach Neustart mit Projektpfad wiederfinden
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" an
    Und ich lege das Teilziel "Texte" unter "Redaktion" an
    Und ich lege das Stammziel "Archiv" in "Nordstern" an
    Und ich lege das Projekt "Artikel" mit dem Ziel "Texte" in "Nordstern" an
    Wenn ich bei "Redaktion" das übergeordnete Ziel "Archiv" auswähle und die Verschiebung speichere
    Dann sehe ich den Zielpfad "Archiv > Redaktion > Texte" mit dem Projekt "Artikel"
    Und ich sehe "Archiv" als übergeordnetes Ziel von "Redaktion"
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Seite neu lade
    Dann sehe ich den Zielpfad "Archiv > Redaktion > Texte" mit dem Projekt "Artikel"
    Und ich sehe "Archiv" weiterhin als übergeordnetes Ziel von "Redaktion"

  Szenariogrundriss: Eine zyklische Verschiebung zeigt einen Feldhinweis und erhält den Baum
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" an
    Und ich lege das Teilziel "Texte" unter "Redaktion" an
    Wenn ich bei "<Ziel>" das übergeordnete Ziel "<Elternziel>" auswähle und die Verschiebung speichere
    Dann sehe ich einen Hinweis am Zielfeld "Übergeordnetes Ziel"
    Und ich sehe "Wissensportal" weiterhin als übergeordnetes Ziel von "Redaktion"
    Und ich sehe "Redaktion" weiterhin als übergeordnetes Ziel von "Texte"

    Beispiele:
      | Ziel          | Elternziel |
      | Redaktion     | Redaktion  |
      | Wissensportal | Texte      |

  Szenario: Eine aufgelöste Elternziel-Kante ist im Browser dauerhaft sichtbar
    Angenommen ich öffne die Anwendung mit einer neuen SQLite-Datenbank im Browser
    Und ich lege die Organisation "Nordstern" an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" an
    Wenn ich bei "Redaktion" die Auswahl "Stammziel" als übergeordnetes Ziel speichere
    Dann sehe ich "Redaktion" als Stammziel
    Wenn ich die Seite neu lade
    Dann sehe ich "Redaktion" weiterhin als Stammziel
