# language: de
@story-66 @act-02
Funktionalität: Aktivität filtern und als CSV exportieren
  Als lokale Betreiberin
  möchte ich nur passende Ereignisse sehen und exportieren,
  damit die Auswertung nachvollziehbar und auf meine Organisation begrenzt bleibt.

  Grundlage:
    Angenommen der lokale Server läuft mit einer neuen SQLite-Datenbank
    Und in "Nordstern" bestehen das Projekt "Website" und der Agent "Mira"
    Und in "Nordstern" besteht der Agent "Noah"
    Und die Organisation "Südstern" besteht

  Szenario: Agent, Aktion, Zeitraum und Objekt wirken gemeinsam
    Angenommen mehrere Aufgabenereignisse mit verschiedenen Agenten, Aktionen und Zeitpunkten bestehen in "Nordstern"
    Wenn ich den Aktivitätsverlauf von "Nordstern" nach Agent "Mira", Aktion "assigned", dem Zeitpunkt des mittleren Ereignisses und dessen Aufgabe filtere
    Dann enthält die gefilterte API-Antwort ausschließlich das passende Ereignis
    Und jedes Ereignis enthält Kennung, Zeitpunkt, Aktion, Agent und Deep Link zur Aufgabe

  Szenario: Sortierung und Seiten bleiben stabil
    Angenommen mehrere Aufgabenereignisse mit verschiedenen Agenten, Aktionen und Zeitpunkten bestehen in "Nordstern"
    Wenn ich den Aktivitätsverlauf von "Nordstern" seitenweise mit zwei Ereignissen je Seite abrufe
    Dann sind die Ereignisse über alle Seiten vollständig, überschneidungsfrei und stabil sortiert
    Und ein erneuter Abruf liefert dieselben Seiten in derselben Reihenfolge

  Szenario: CSV enthält exakt die sichtbare gefilterte Seite
    Angenommen mehrere Aufgabenereignisse mit verschiedenen Agenten, Aktionen und Zeitpunkten bestehen in "Nordstern"
    Wenn ich eine gefilterte Seite des Aktivitätsverlaufs von "Nordstern" abrufe und als CSV exportiere
    Dann entsprechen die CSV-Datensätze in Inhalt und Reihenfolge exakt der API-Seite
    Und die CSV-Antwort ist als CSV-Download gekennzeichnet

  Szenario: CSV-Quoting erhält Unicode, Trennzeichen, Anführungszeichen und Zeilenumbruch
    Angenommen ein Aufgabenereignis in "Nordstern" enthält Unicode, Komma, Anführungszeichen und Zeilenumbruch in einem sichtbaren Feld
    Wenn ich den Aktivitätsverlauf von "Nordstern" als CSV exportiere
    Dann liest ein standardkonformer CSV-Parser den ursprünglichen Feldwert unverändert
    Und die CSV enthält keinen zusätzlichen Ereignisdatensatz

  Szenario: Leere Filterergebnisse ergeben leere Liste und gültige CSV
    Wenn ich den Aktivitätsverlauf von "Nordstern" nach einem nicht passenden Objekt filtere
    Dann ist die gefilterte API-Antwort leer
    Und der Export enthält nur die CSV-Kopfzeile

  Szenario: Organisationsgrenze und ungültige Filter bleiben geschützt
    Angenommen in "Südstern" besteht eine Aufgabe mit Aktivität
    Wenn ich den Aktivitätsverlauf von "Nordstern" filtere und als CSV exportiere
    Dann enthält der Export keine Ereignisse aus "Südstern"
    Und enthält weder Token noch Geheimnisse oder vollständige Prompts
    Wenn ich einen ungültigen Zeitraum oder eine ungültige Seitengröße an die Aktivitäts-API sende
    Dann werden Liste und Export mit einem verständlichen Clientfehler abgewiesen

  Szenario: Filter und Export bleiben nach Neustart identisch
    Angenommen mehrere Aufgabenereignisse mit verschiedenen Agenten, Aktionen und Zeitpunkten bestehen in "Nordstern"
    Wenn ich eine gefilterte Seite des Aktivitätsverlaufs von "Nordstern" abrufe und als CSV exportiere
    Und ich den Server beende und mit derselben SQLite-Datenbank neu starte
    Dann liefern dieselben Filter dieselbe API-Seite und dieselben CSV-Datensätze
