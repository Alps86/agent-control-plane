# language: de
@story-66 @act-02 @browser
Funktionalität: Aktivität im Browser filtern und exportieren
  Als lokale Betreiberin
  möchte ich die sichtbare Aktivität eingrenzen und herunterladen,
  damit der Export genau meiner gewählten Ansicht entspricht.

  Szenario: Kombinierte Filter und Export der sichtbaren Seite
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Und in "Nordstern" besteht der Agent "Noah"
    Und mehrere Aufgabenereignisse mit verschiedenen Agenten, Aktionen und Zeitpunkten bestehen in "Nordstern"
    Wenn ich den Aktivitätsverlauf von "Nordstern" im Browser öffne
    Und ich Agent "Mira", Aktion "assigned", Zeitraum und Aufgabe als Aktivitätsfilter wähle
    Dann sehe ich nur die passenden Ereignisse in stabiler Reihenfolge
    Wenn ich die sichtbare Aktivitätsseite als CSV herunterlade
    Dann enthält der Download genau die sichtbaren Ereignisse in derselben Reihenfolge

  Szenario: Leerzustand nach Filter und Schutz vor fremder Organisation
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Und die Organisation "Südstern" besteht
    Und in "Südstern" besteht eine Aufgabe mit Aktivität
    Wenn ich den Aktivitätsverlauf von "Nordstern" im Browser öffne
    Und ich einen Filter ohne Treffer wähle
    Dann sehe ich einen verständlichen Leerzustand
    Und der CSV-Download enthält nur die Kopfzeile und keine fremde Aktivität
