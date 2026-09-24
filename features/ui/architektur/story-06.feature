# language: de
@story-06 @ui-02
Funktionalität: Öffentliche UI-Bridge für gebaute Ansichten
  Als Anwendungsserver möchte ich die gebaute Oberfläche über eine einzige
  öffentliche Go-Grenze verwenden, damit Vorschau und echte Daten dieselben
  Templates und Assets nutzen.

  Grundlage:
    Angenommen die UI ist aus den Vite-Quellen frisch gebaut und in die Bridge eingebettet

  Szenario: Eingebettete Seite und Assets nach einem frischen Build ausliefern
    Wenn der Anwendungsserver die Organisationsansicht mit der Map der JSON-Fixture "organization.json" als ganze Seite rendert
    Dann enthält die Antwort ein vollständiges HTML-Dokument mit dem Namen der Beispielorganisation
    Und die referenzierten CSS- und HTMX-Assets sind über die öffentliche Bridge erreichbar
    Und die ausgelieferten Assets stammen aus dem eingebetteten Build

  Szenario: Ganze Seite und HTMX-Fragment mit derselben Ansichts-Map rendern
    Angenommen der Anwendungsserver verwendet die Map der JSON-Fixture "projects.json"
    Wenn er daraus die Projektansicht als ganze Seite und als HTMX-Fragment rendert
    Dann zeigen beide Antworten denselben Projekttitel und dieselben Projekte
    Und nur die ganze Seite enthält Dokumentrahmen und Navigation
    Und das Fragment enthält ausschließlich den austauschbaren Inhaltsbereich

  Szenario: Benannte Templates aus einem Unterordner ohne Kollision rendern
    Angenommen der frische Build enthält die neutralen Templates "bridge-check/page.html" und "bridge-check/content.html"
    Und der Anwendungsserver verwendet für alle Templates dieselbe Ansichts-Map
    Wenn er "bridge-check/page" und "bridge-check/content" über die öffentliche Bridge rendert
    Dann enthalten beide Antworten die Werte dieser Ansichts-Map
    Und die benannten Root-Templates "page" und "content" bleiben mit derselben Map renderbar

  Szenario: Eine andere JSON-Fixture ohne Go-Beispieldaten verwenden
    Wenn die Vorschau über die öffentliche Bridge "organization-alternate.json" lädt und damit die Organisationsansicht rendert
    Dann sehe ich den Organisationsnamen der alternativen JSON-Fixture
    Und ich sehe dort nicht den Organisationsnamen der zuerst ausgelieferten Fixture

  Szenario: Anzeigetext wird im HTML maskiert
    Angenommen die Ansichts-Map enthält als Organisationsnamen "<script>alert(1)</script>"
    Wenn der Anwendungsserver die Organisationsansicht als ganze Seite rendert
    Dann steht der Organisationsname nur als maskierter Text im HTML
    Und kein ausführbares Script aus dem Organisationsnamen erscheint im Dokument

  Szenario: Nur eingebettete, benannte Ressourcen ausliefern
    Wenn eine unbekannte Fixture oder ein Pfad außerhalb des Fixture-Verzeichnisses angefordert wird
    Dann verweigert die öffentliche Bridge den Zugriff ohne Dateiinhalte preiszugeben
    Wenn ein unbekanntes Asset angefordert wird
    Dann liefert die öffentliche Bridge eine Nicht-gefunden-Antwort
