# language: de
@story-12 @ui-04
Funktionalität: Gespeicherte Organisationen sicher in der Oberfläche projizieren
  Als lokale Betreiberin
  möchte ich nur meine gespeicherten Organisationen in vollständigen Seiten und Fragmenten sehen,
  damit die Oberfläche echte Fachdaten anzeigt und direkte HTTP-Aufrufe dieselben Rechte beachten.

  @browser
  Szenario: Browser und API zeigen dieselbe gespeicherte Organisation nach einem Neustart
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" mit der Beschreibung "Echte Fachdaten" über die öffentliche JSON-API an
    Wenn ich die Organisationsübersicht im Browser öffne
    Dann enthält die HTML-Übersicht "Nordstern" mit "Echte Fachdaten" genau einmal
    Und die Detailseite dieser Organisation zeigt denselben Namen und dieselbe Beschreibung
    Wenn ich den Server mit derselben SQLite-Datenbank neu starte
    Dann enthält die HTML-Übersicht "Nordstern" mit "Echte Fachdaten" erneut genau einmal
    Und die öffentliche JSON-API liefert dieselbe Organisationskennung

  Szenario: Vollseite und HTMX-Fragment projizieren dieselben freigegebenen Daten
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" mit der Beschreibung "Nur eigene Daten" über die öffentliche JSON-API an
    Wenn ich die Organisationsübersicht als vollständige HTML-Seite und mit "HX-Request: true" abrufe
    Dann enthalten beide Antworten "Nordstern" und "Nur eigene Daten"
    Und die vollständige Antwort enthält ein HTML-Dokument mit Navigation
    Und das Fragment enthält nur den Inhalt der Organisationsübersicht ohne HTML-Dokument

  Szenario: Fachtexte werden in Übersicht und Detail sicher ausgegeben
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "<script>alert(1)</script>" mit der Beschreibung "<img src=x onerror=alert(2)>" über die öffentliche JSON-API an
    Wenn ich die Organisationsübersicht und die Detailseite im Browser öffne
    Dann sehe ich beide Fachtexte als Text
    Und es wurde weder ein Script ausgeführt noch ein Bild-Fehlerhandler ausgelöst
    Und die HTML-Antworten enthalten kein ausführbares Script oder Bild aus diesen Fachtexten

  Szenario: HTML-Ansichten verraten einer fremden Betreiberin keine Organisation
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und ich lege die Organisation "Nordstern" mit der Beschreibung "Vertraulich" über die öffentliche JSON-API an
    Wenn eine andere serverseitige Betreiberidentität die Organisationsübersicht und Detailseite direkt über HTTP abruft
    Dann enthält ihre Übersicht weder "Nordstern" noch "Vertraulich"
    Und die Detailseite antwortet wie für eine unbekannte Kennung ohne Organisationsdaten

  Szenario: Direkter HTML-Schreibaufruf ohne gültige Betreiberidentität bleibt wirkungslos
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich ohne gültige serverseitige Betreiberidentität "POST /organisationen" mit Name "Unerlaubt" direkt aufrufe
    Dann antwortet die HTML-Route mit HTTP-Status 403 ohne Organisationsdaten
    Und die öffentliche JSON-API der Betreiberin zeigt keine neu angelegte Organisation

  Szenario: Ein unbekanntes Formularfeld ändert den Organisationsvertrag nicht
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Wenn ich "POST /organisationen" mit gültigem Namen "Budgetversuch" und unbekanntem Feld "budgetLimit" direkt aufrufe
    Dann antwortet die HTML-Route mit HTTP-Status 400 oder 422 ohne neue Organisation
    Und die öffentliche JSON-API der Betreiberin zeigt keine neu angelegte Organisation
