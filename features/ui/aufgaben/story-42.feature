# language: de
@story-42 @task-05 @browser
Funktionalität: Aufgabenthread im Browser bedienen
  Als lokale Betreiberin
  möchte ich Kommentare einer Aufgabe lesen und schreiben,
  damit ich ihren Verlauf und ihre Herkunft im Aufgabendetail nachvollziehen kann.

  Szenario: Betreiberkommentar erscheint mit Herkunft und bleibt nach Neustart sichtbar
    Angenommen ich öffne die Kommentaransicht "/kommentare" von "Texte prüfen" im Projekt "Website" der Organisation "Nordstern" in einem echten Browser
    Wenn ich im Aufgabenthread "Fachprüfung folgt" eingebe und absende
    Dann sehe ich "Fachprüfung folgt" mit Quelle "Betreiberin" und Zeitpunkt im Thread
    Und die Kommentaransicht bleibt derselben Aufgabenkennung zugeordnet
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte und die Kommentaransicht neu lade
    Dann sehe ich denselben Kommentar mit derselben Herkunft und demselben Zeitpunkt genau einmal

  Szenario: Agentenkommentar und Artefaktreferenz erscheinen im selben Thread
    Angenommen der vertrauenswürdige Metadatenresolver kennt "entwurf-7" als Artefakt "Textentwurf" von "Nordstern"
    Und "Mira" hat als berechtigter Agent "Entwurf ist bereit" mit Artefaktkennung "entwurf-7" an "Texte prüfen" kommentiert
    Wenn ich die Kommentaransicht "/kommentare" von "Texte prüfen" im Browser öffne
    Dann sehe ich "Entwurf ist bereit" mit Quelle "Mira" und Zeitpunkt
    Und ich sehe Kennung, Anzeigenamen "Textentwurf" und den internen Link an genau diesem Kommentar
    Und die Anzeige bietet keinen Datei- oder Downloadlink an

  Szenario: Leerer Inhalt bleibt mit Feldhinweis im Formular
    Angenommen ich öffne die Kommentaransicht "/kommentare" von "Texte prüfen" im Browser
    Wenn ich nur Leerzeichen im Kommentarfeld eingebe und absende
    Dann sehe ich einen Hinweis am Kommentarfeld
    Und im Aufgabenthread erscheint kein neuer Kommentar

  Szenario: Ein fremder Task erscheint nicht über eine direkte Detailadresse
    Angenommen ich betrachte die Organisation "Nordstern" im Browser
    Und "Fremdaufgabe" gehört zur Organisation "Suedstern"
    Wenn ich die direkte Kommentaradresse "/kommentare" von "Fremdaufgabe" im Browser öffne
    Dann sehe ich weder Aufgabeninhalt noch Kommentare oder Artefaktbezüge von "Fremdaufgabe"
