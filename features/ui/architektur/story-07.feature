# language: de
@story-07 @ui-03
Funktionalität: Geschützte lokale Go-Vorschau aus JSON-Fixtures
  Als Betreiber möchte ich die gebauten Ansichten lokal mit Beispieldaten prüfen,
  ohne eine öffentlich erreichbare Vorschau oder echte Anwendungsdaten bereitzustellen.

  Grundlage:
    Angenommen die UI wurde frisch gebaut und die Go-Vorschau wird mit einem freien Vorschauport gestartet

  Szenario: Vorschau nur über die lokale Adresse erreichen
    Dann lauscht die Vorschau ausschließlich auf einer Loopback-Adresse
    Und der konfigurierte Vorschauport ist über diese Adresse erreichbar
    Und eine Verbindung über eine andere lokale Netzwerkschnittstelle wird nicht angenommen

  Szenario: Ungültigen oder belegten Vorschauport melden
    Wenn die Vorschau mit einem ungültigen oder bereits belegten Vorschauport gestartet wird
    Dann endet der Start mit einer verständlichen Fehlermeldung
    Und auf diesem Port wird kein weiterer Vorschauprozess gestartet

  Szenario: Ganze Organisationsseite im Browser öffnen
    Wenn ich im Browser "/?view=organization" der Go-Vorschau öffne
    Dann sehe ich ein vollständiges HTML-Dokument mit der Beispielorganisation "Atelier Nord"
    Und ich kann über die Navigation Projekte, Agenten und Aufgaben öffnen

  Szenario: HTMX erhält nur den Inhalt zur selben Ansicht
    Wenn ich "/?view=projects" als normale Browseranfrage und als HTMX-Anfrage öffne
    Dann zeigen beide Antworten dieselben Projekte aus "projects.json"
    Und nur die normale Antwort enthält Dokumentrahmen und Navigation
    Und die HTMX-Antwort enthält nur den austauschbaren Inhaltsbereich

  Szenario: Auswahl einer anderen JSON-Fixture bleibt bei der Navigation erhalten
    Wenn ich im Browser "/?view=organization&fixture=alternate" öffne
    Dann sehe ich "Küstenwerk" und nicht "Atelier Nord"
    Wenn ich von dort über die Navigation die Projektansicht öffne
    Dann zeigt die Adresse weiterhin die Fixture-Auswahl "alternate"
    Und die Projektansicht verwendet die Daten aus "projects-alternate.json"

  Szenario: Änderung allein an JSON verändert die neu gebaute Vorschau
    Angenommen eine Kopie der Organisations-Fixture enthält einen neuen eindeutigen Namen
    Wenn ich die UI aus dieser Kopie ohne Go-Codeänderung neu baue und die Go-Vorschau starte
    Dann zeigt die Organisationsansicht den neuen Namen aus der JSON-Fixture

  Szenario: Leere und fehlerhafte Beispieldaten anzeigen
    Wenn ich "/?view=projects&fixture=empty" öffne
    Dann sehe ich den verständlichen Leerzustand der Projektansicht
    Wenn ich "/?view=tasks&fixture=error" öffne
    Dann sehe ich den verständlichen Aufgabenfehler mit einem Rückweg zur Übersicht

  Szenario: Gebaute Assets und das Hilfe-Fragment laden
    Wenn ich die Organisationsseite der Go-Vorschau öffne
    Dann laden die referenzierten CSS- und HTMX-Dateien erfolgreich
    Wenn ich die Vorschauhilfe im Browser öffne
    Dann erscheint das eingebettete Hilfe-Fragment ohne vollständigen Seitenrahmen

  Szenario: Unbekannte Ansicht und Fixture nicht durch Beispieldaten ersetzen
    Wenn ich eine unbekannte Ansicht oder Fixture über die Vorschauadresse anfordere
    Dann erhalte ich eine Nicht-gefunden-Antwort ohne Beispieldaten einer anderen Ansicht

  Szenario: Pfadzugriffe außerhalb der eingebetteten Vorschau abweisen
    Wenn ich einen Pfad außerhalb der erlaubten Ansichts-, Asset- und Fragmentrouten anfordere
    Dann erhalte ich eine Nicht-gefunden-Antwort ohne lokale Dateiinhalte
    Wenn ich einen Pfadversuch im Fixture-Namen oder Asset-Pfad anfordere
    Dann erhalte ich keine lokalen Dateiinhalte und keine Fixture-JSON-Antwort
