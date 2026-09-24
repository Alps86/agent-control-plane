# language: de
@story-05 @ui-01
Funktionalität: Statische Arbeitsoberfläche aus JSON-Beispieldaten
  Als Betreiber möchte ich die geplante Organisation, Projekte, Agenten und Aufgaben
  in einer Vorschau erkunden, bevor diese Ansichten an echte Anwendungsdaten gebunden sind.

  Grundlage:
    Angenommen die statische Vorschau ist mit den ausgelieferten JSON-Beispieldaten geöffnet

  Szenario: Von der Organisation zu Projekten, Agenten und Aufgaben wechseln
    Wenn ich die Organisationsübersicht öffne
    Dann sehe ich den Namen der Beispielorganisation und ihre Übersicht
    Und ich kann über die Navigation die Projekte, Agenten und Aufgaben öffnen
    Wenn ich nacheinander die Projekt-, Agenten- und Aufgabenansicht öffne
    Dann zeigt jede Ansicht ihren eigenen Titel und die zugehörigen Beispieldaten
    Und die Navigation zeigt mir die aktuell geöffnete Ansicht an

  Szenario: Leere Listen verständlich darstellen
    Angenommen ich wähle die Beispieldaten für eine Organisation ohne Projekte, Agenten und Aufgaben
    Wenn ich nacheinander die Projekt-, Agenten- und Aufgabenansicht öffne
    Dann zeigt jede Ansicht einen passenden Leerzustand statt einer leeren Fläche
    Und die Navigation zur Organisation bleibt benutzbar

  Szenario: Fehlerhafte Ansicht mit nutzbarem Rückweg darstellen
    Angenommen ich wähle die Beispieldaten für eine fehlerhafte Aufgabenansicht
    Wenn ich die Aufgabenansicht öffne
    Dann sehe ich einen verständlichen Fehlerhinweis ohne rohe technische Fehlermeldung
    Und ich kann über die Navigation zur Organisationsübersicht zurückkehren

  Szenario: Ansichten beziehen ihre Beispieldaten aus der gewählten JSON-Fixture
    Angenommen ich wähle eine zweite JSON-Fixture mit einem anderen Organisationsnamen
    Wenn ich die Organisationsübersicht öffne
    Dann sehe ich den Organisationsnamen dieser zweiten Fixture
    Und ich sehe dort nicht den Organisationsnamen der zuerst ausgelieferten Fixture

  Szenario: Navigation auf einem schmalen Bildschirm verwenden
    Angenommen ich betrachte die statische Vorschau auf einem schmalen Bildschirm
    Wenn ich die Organisationsübersicht öffne
    Dann kann ich die Navigation zu Projekten, Agenten und Aufgaben erreichen
    Und die Inhalte sind ohne horizontales Scrollen lesbar
