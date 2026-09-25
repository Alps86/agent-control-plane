# language: de
@story-18 @act-01 @browser
Funktionalität: Aktivitätsverlauf im Browser öffnen
  Als lokale Betreiberin
  möchte ich aus der Organisation den Aktivitätsverlauf öffnen,
  damit ich von einem Ereignis direkt zur Aufgabe gelangen kann.

  Szenario: Anlage und Zuweisung sind im Browser sichtbar und verlinkt
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Und "Startseite prüfen" ist im Projekt "Website" dem Agenten "Mira" zugewiesen
    Wenn ich die Organisation "Nordstern" im Browser öffne
    Und ich von dort den Aktivitätsverlauf öffne
    Dann sehe ich getrennt die Anlage und die Zuweisung von "Startseite prüfen"
    Und ich sehe bei beiden Ereignissen Akteur, Quelle und Zeitpunkt
    Wenn ich den Deep Link eines Ereignisses öffne
    Dann sehe ich das Aufgabendetail von "Startseite prüfen"

  Szenario: Leerer Verlauf bleibt verständlich
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und die Organisation "Nordstern" besteht
    Wenn ich den Aktivitätsverlauf von "Nordstern" im Browser öffne
    Dann sehe ich einen verständlichen Leerzustand ohne Aufgabenereignis
