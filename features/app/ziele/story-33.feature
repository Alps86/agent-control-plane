# language: de
@story-33 @goal-03
Funktionalität: Betreiberin verschiebt und verknüpft bestehende Ziele über HTTP
  Als Betreiberin möchte ich die Elternziel-Kante ausdrücklich ändern,
  damit Zielpfade aktualisiert werden, ohne Status oder Projektbeziehungen zu verlieren.

  Hintergrund:
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" über HTTP an
    Und ich lege das Stammziel "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Redaktion" unter "Wissensportal" in "Nordstern" über HTTP an
    Und ich lege das Teilziel "Texte" unter "Redaktion" in "Nordstern" über HTTP an
    Und ich lege das Stammziel "Archiv" in "Nordstern" über HTTP an
    Und ich lege das Projekt "Artikel" mit dem Ziel "Texte" in "Nordstern" über HTTP an

  Szenario: Ein vorhandener Teilbaum wird unter ein anderes Ziel verschoben und bleibt nach Neustart erhalten
    Wenn ich den Status von "Redaktion" in "Nordstern" über HTTP ausdrücklich auf "erreicht" setze
    Und ich "Redaktion" in "Nordstern" über HTTP unter "Archiv" verschiebe
    Dann enthält der Zielbaum von "Nordstern" den Pfad "Archiv > Redaktion > Texte" für das Projekt "Artikel"
    Und "Redaktion" hat "Archiv" als übergeordnetes Ziel
    Und "Redaktion" hat weiterhin den Status "erreicht"
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann enthält der Zielbaum von "Nordstern" weiterhin den Pfad "Archiv > Redaktion > Texte" für das Projekt "Artikel"
    Und "Redaktion" hat weiterhin "Archiv" als übergeordnetes Ziel

  Szenario: Ein bestehendes Stammziel wird verknüpft und kann wieder zum Stammziel werden
    Wenn ich "Archiv" in "Nordstern" über HTTP unter "Wissensportal" verknüpfe
    Dann hat "Archiv" das übergeordnete Ziel "Wissensportal"
    Wenn ich die Elternziel-Verknüpfung von "Archiv" in "Nordstern" über HTTP löse
    Dann ist "Archiv" wieder ein Stammziel
    Und der Zielbaum von "Nordstern" enthält weiterhin alle vier Ziele genau einmal

  Szenariogrundriss: Selbst- und Nachfahrenzyklen lassen die bisherige Struktur unberührt
    Wenn ich "<Ziel>" in "Nordstern" über HTTP unter "<Elternziel>" verschiebe
    Dann antwortet der Server mit einem Hinweis an der Kante "Übergeordnetes Ziel"
    Und "Redaktion" hat weiterhin "Wissensportal" als übergeordnetes Ziel
    Und "Texte" hat weiterhin "Redaktion" als übergeordnetes Ziel
    Und der Zielbaum von "Nordstern" enthält weiterhin den Pfad "Wissensportal > Redaktion > Texte" für das Projekt "Artikel"

    Beispiele:
      | Ziel           | Elternziel |
      | Redaktion      | Redaktion  |
      | Wissensportal  | Texte      |

  Szenario: Fremde Organisation kann weder Elternziel noch verschobenes Ziel liefern
    Angenommen ich lege die Organisation "Südstern" über HTTP an
    Und ich lege das Stammziel "Südportal" in "Südstern" über HTTP an
    Wenn ich "Redaktion" in "Nordstern" über HTTP unter "Südportal" verschiebe
    Dann antwortet der Server mit einem Hinweis an der Kante "Übergeordnetes Ziel"
    Und "Redaktion" hat weiterhin "Wissensportal" als übergeordnetes Ziel
    Und "Südportal" bleibt ein Stammziel in "Südstern"
    Wenn ich "Südportal" über die Organisation "Nordstern" unter "Archiv" verschiebe
    Dann antwortet der Server ohne fremde Zieldaten mit HTTP-Status 404
    Und "Südportal" bleibt ein Stammziel in "Südstern"
    Und der Zielbaum von "Nordstern" enthält weiterhin den Pfad "Wissensportal > Redaktion > Texte" für das Projekt "Artikel"

  Szenario: Eine nicht zugeordnete Betreiberin kann keine Elternziel-Kante ändern
    Wenn ich als nicht zugeordnete Betreiberin "Redaktion" in "Nordstern" unter "Archiv" verschiebe
    Dann antwortet der Server mit einem datenfreien HTTP-Status 404
    Und "Redaktion" hat weiterhin "Wissensportal" als übergeordnetes Ziel
