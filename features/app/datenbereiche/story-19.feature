# language: de
@story-19 @perm-01
Funktionalität: Datenbereich eines Eino-Agenten pro Projekt erzwingen
  Als lokale Betreiberin
  möchte ich einem Agenten Projekte ausdrücklich zum Lesen oder Schreiben freigeben,
  damit er nur die Daten seiner Organisation und seiner freigegebenen Projekte erhält oder ändert.

  Hintergrund:
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisationen "Nordstern" und "Suedstern" bestehen
    Und die Projekte "Website" und "Intern" gehören zu "Nordstern"
    Und das Projekt "Fremd" gehört zu "Suedstern"
    Und "Mira" ist ein Eino-Agent von "Nordstern"
    Und der öffentliche Projekt-Datenpfad ermittelt "Mira" über den serverseitigen Identitätsport

  @eino
  Szenario: Eine Agentenvorlage erteilt keinen Datenbereich
    Wenn "Mira" über die öffentliche App-Fassade die Daten von "Website" liest
    Dann wird der Lesezugriff ohne Projektdaten verweigert
    Und die Verweigerung ist mit Agent, Projekt, Aktion und Zeitpunkt auditiert

  @eino
  Szenario: Eine ausdrückliche Lesefreigabe erlaubt nur Lesen im gewählten Projekt
    Wenn ich "Mira" über die öffentliche Betreiber-API "Website" zum Lesen freigebe
    Dann zeigt die Datenbereichs-API "Website" mit Lesen und ohne Schreiben
    Wenn "Mira" über die öffentliche App-Fassade die Daten von "Website" liest
    Dann erhält "Mira" ausschließlich die Daten von "Website"
    Wenn "Mira" über die öffentliche App-Fassade Daten in "Website" schreibt
    Dann wird der Schreibzugriff ohne Datenänderung verweigert
    Und die Verweigerung ist mit Agent, Projekt, Aktion und Zeitpunkt auditiert

  @eino
  Szenario: Schreibfreigabe gilt nur für das eigene freigegebene Projekt
    Angenommen ich habe "Mira" "Website" zum Lesen und Schreiben freigegeben
    Wenn "Mira" über die öffentliche App-Fassade Daten in "Website" schreibt
    Dann ist die Änderung beim Lesen von "Website" sichtbar
    Wenn "Mira" über die öffentliche App-Fassade Daten in "Intern" schreibt
    Dann wird der Schreibzugriff ohne Datenänderung verweigert
    Und die Antwort enthält weder Kennung noch Inhalt von "Intern"
    Und die Verweigerung ist mit Agent, Projekt, Aktion und Zeitpunkt auditiert

  @eino
  Szenario: Unbekanntes und nicht freigegebenes Projekt verraten keine Projektdaten
    Angenommen ich habe "Mira" "Website" zum Lesen freigegeben
    Wenn "Mira" über die öffentliche App-Fassade die Daten von "Intern" liest
    Dann wird der Lesezugriff ohne Projektdaten verweigert
    Wenn "Mira" über die öffentliche App-Fassade die Daten eines unbekannten Projekts liest
    Dann sind Fehlerklasse und Projektausgabe gleich der vorherigen Verweigerung
    Und beide Verweigerungen sind getrennt auditiert

  @eino
  Szenario: Ein fremdes Projekt wird ohne Teiländerung abgewiesen
    Angenommen ich habe "Mira" "Website" zum Lesen freigegeben
    Wenn ich über die öffentliche Betreiber-API "Website" und "Fremd" für "Mira" zum Lesen freigebe
    Dann wird die gesamte Änderung ohne fremde Projektdaten abgewiesen
    Und die Datenbereichs-API zeigt weiterhin nur "Website" zum Lesen
    Und "Mira" kann die Daten von "Fremd" weder lesen noch schreiben
    Und beide Verweigerungen sind ohne fremde Projektkennung auditiert

  @eino
  Szenario: Fremde Agenten- und Organisationskennungen gewähren keinen Datenbereich
    Angenommen ich habe "Mira" "Website" zum Lesen freigegeben
    Wenn ich "Mira" unter "Suedstern" über die öffentliche Datenbereichs-API abrufe
    Dann erhalte ich eine datenfreie Antwort mit HTTP-Status 404
    Wenn ein Betreiberakteur mit "Miras" Kennung über die öffentliche App-Fassade "Website" lesen möchte
    Dann bleibt die serverseitig ermittelte Identität für den Projekt-Datenpfad maßgeblich
    Und der Aufruf erhält keine Daten von "Website"

  @eino
  Szenario: Auch abgeschaltete fremde Projektwerte werden vor dem Ersetzen geprüft
    Angenommen ich habe "Mira" "Website" zum Lesen freigegeben
    Wenn ich "Website" zum Lesen und "Fremd" ohne Lesen und Schreiben über die öffentliche Betreiber-API sende
    Dann wird die gesamte Änderung ohne fremde Projektdaten abgewiesen
    Und die Datenbereichs-API zeigt weiterhin nur "Website" zum Lesen
    Wenn ich "Website" zum Lesen und ein unbekanntes Projekt ohne Lesen und Schreiben über die öffentliche Betreiber-API sende
    Dann wird die gesamte Änderung ohne fremde Projektdaten abgewiesen
    Und die Datenbereichs-API zeigt weiterhin nur "Website" zum Lesen

  @eino
  Szenario: Entzug gilt beim nächsten Aufruf und bleibt nach Neustart erhalten
    Angenommen ich habe "Mira" "Website" zum Lesen und Schreiben freigegeben
    Wenn ich "Mira" über die öffentliche Betreiber-API das Schreiben in "Website" entziehe
    Und ich den Server mit derselben SQLite-Datenbank neu starte
    Dann zeigt die Datenbereichs-API "Website" mit Lesen und ohne Schreiben
    Wenn "Mira" über die öffentliche App-Fassade Daten in "Website" schreibt
    Dann wird der Schreibzugriff ohne Datenänderung verweigert
    Und die Verweigerung bleibt nach einem weiteren Neustart im Audit sichtbar

  @codex-cli
  Szenario: Codex-CLI-Agent nutzt dieselbe serverseitige Projektgrenze
    Angenommen "Kai" ist ein Codex-CLI-Agent von "Nordstern"
    Und der öffentliche Projekt-Datenpfad ermittelt "Kai" über den serverseitigen Identitätsport
    Wenn "Kai" über die öffentliche App-Fassade die Daten von "Website" liest
    Dann wird der Lesezugriff ohne Projektdaten verweigert
    Und die Verweigerung ist mit Agent, Projekt, Aktion und Zeitpunkt auditiert
    Wenn ich "Kai" über die öffentliche Betreiber-API "Website" zum Lesen und Schreiben freigebe
    Dann zeigt die Datenbereichs-API für "Kai" "Website" mit Lesen und Schreiben
    Wenn "Kai" über die öffentliche App-Fassade Daten in "Website" schreibt
    Dann ist die Änderung beim Lesen von "Website" durch "Kai" sichtbar
    Wenn "Kai" über die öffentliche App-Fassade die Daten von "Intern" liest
    Dann wird der Lesezugriff ohne Projektdaten verweigert
    Und die Verweigerung ist mit Agent, Projekt, Aktion und Zeitpunkt auditiert
    Und "Kai" bleibt ohne Codex-CLI-Durchsetzungsnachweis nicht startbereit

  # Die App-Fassade belegt die Datenbereichsentscheidung für den konfigurierten
  # Codex-CLI-Agenten. Ein tatsächlicher CLI-Lauf und sein Toolprofil bleiben
  # gesonderte Abnahme von Story-21.
