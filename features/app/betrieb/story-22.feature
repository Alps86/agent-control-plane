# language: de
@story-22 @ops-05
Funktionalität: Lokale Anwendung ohne App-Konto und Budgetverwaltung nutzen
  Als lokaler Betreiber
  möchte ich die installierte Anwendung ohne App-Anmeldung verwenden,
  damit Modellanbieter-Zugänge nur für Modellaufrufe nötig sind und kein Budget verwaltet wird.

  # Die Szenarien laufen gegen den installierten Go-Server und seine öffentlichen HTTP-Wege.
  # Basis: OPS-01/Story-10 und UI-02/Story-06 sind integriert.
  # Die echten öffentlichen Wege für Startseite mit leerer Organisationsübersicht,
  # Organisationsformular und Organisations-API sind das Gate ORG-01/Story-11.
  # Konkrete Anbieter-Settings und ihre Verbindung werden in MS- und OPS-02-Stories geprüft.
  # Export/Import aus OPS-04/Story-86 gehört nicht zu dieser frühen Abnahme.

  Szenario: Lokaler Server antwortet ohne App-Anmeldung und ohne Budgetvertrag
    Angenommen der installierte Go-Server ist mit einer neuen lokalen Datenbank gestartet
    Wenn ich ohne Cookie oder Authorization-Header GET "/health" an den Server sende
    Dann antwortet der Server mit HTTP 200 und einer Schemaversion
    Und die Antwort verlangt weder Login noch Registrierung
    Und die JSON-Antwort enthält keine Budgetfelder oder Budgetlimits

  @gate-story-11
  Szenario: Startseite ist ohne App-Konto nutzbar
    Angenommen der installierte Go-Server ist mit einer neuen lokalen Datenbank gestartet
    Wenn ich ohne Cookie oder Authorization-Header die App-Startseite öffne
    Dann sehe ich die echte Organisationsübersicht oder ihren Leerzustand
    Und ich werde nicht auf eine App-Anmeldung oder Registrierung geleitet
    Und die Seite enthält keine App-Anmelde- oder Registrierungsaktion
    Und die Seite zeigt keine Budgetfelder oder Budgetlimits

  @gate-story-11
  Szenario: Organisationsformular bleibt ohne App-Konto und Budgeteingabe bedienbar
    Angenommen der installierte Go-Server ist mit einer neuen lokalen Datenbank gestartet
    Wenn ich ohne Cookie oder Authorization-Header das Organisationsformular öffne
    Dann kann ich Name und Beschreibung eingeben
    Und ich sehe keine Eingabe für Budget oder Budgetlimit
    Wenn ich die Organisation "Nordstern" ohne App-Anmeldung anlege
    Dann sehe ich "Nordstern" in der echten Organisationsübersicht
    Und die Antwort enthält keine Budgetfelder oder Budgetlimits

  @gate-story-11
  Szenariogrundriss: Manipulierter Formular-POST weist Budgeteingaben zurück
    Angenommen der installierte Go-Server ist mit einer neuen lokalen Datenbank gestartet
    Und ich habe ohne Cookie oder Authorization-Header das öffentliche Organisationsformular geöffnet
    Wenn ich das Formular mit Name "Nordstern", einer Beschreibung und dem zusätzlich eingeschleusten Feld <Feld> als HTML-Form-POST absende
    Dann wird die Eingabe als ungültig abgewiesen
    Und es wurde keine Organisation "Nordstern" angelegt
    Und die Antwort enthält weder das übermittelte Budget noch ein Budgetlimit

    Beispiele:
      | Feld            |
      | budget=100      |
      | budgetLimit=100 |

  @gate-story-11
  Szenariogrundriss: Öffentliche Organisations-API weist Budgeteingaben zurück
    Angenommen der installierte Go-Server ist mit einer neuen lokalen Datenbank gestartet
    Wenn ich ohne Cookie oder Authorization-Header über die öffentliche Organisations-API "Nordstern" mit dem zusätzlichen JSON-Feld <Feld> anlege
    Dann wird die Eingabe als ungültig abgewiesen
    Und es wurde keine Organisation "Nordstern" angelegt
    Und die Antwort enthält weder das übermittelte Budget noch ein Budgetlimit

    Beispiele:
      | Feld              |
      | "budget":100     |
      | "budgetLimit":100 |

  @gate-story-11
  Szenario: Öffentliche Organisations-API liefert kein Budget aus
    Angenommen eine Organisation "Nordstern" wurde über die öffentliche API ohne App-Anmeldung angelegt
    Wenn ich sie ohne Cookie oder Authorization-Header über die öffentliche Organisations-API lese
    Dann antwortet die API mit der Organisation "Nordstern"
    Und die JSON-Antwort enthält keine Budgetfelder oder Budgetlimits

  @gate-story-11
  Szenario: Fehlende Modellanbieter-Verbindung sperrt die lokale Anwendung nicht
    Angenommen der installierte Go-Server ist mit einer neuen lokalen Datenbank gestartet
    Und es ist noch kein Modellanbieter verbunden
    Wenn ich ohne Cookie oder Authorization-Header die App-Startseite und das Organisationsformular öffne
    Dann kann ich eine Organisation "Nordstern" ohne App-Anmeldung anlegen
    Und ich sehe "Nordstern" in der echten Organisationsübersicht
    Und die fehlende Modellanbieter-Verbindung hat die lokale Organisationsanlage nicht gesperrt

  @gate-story-11
  Szenario: Entfernter Schreibzugriff mit gefälschtem lokalen Host wird abgewiesen
    Angenommen der installierte Go-Server ist mit neuer Datenbank auf allen Netzadressen gestartet
    Wenn ich über eine Nicht-Loopback-Adresse ohne Origin mit Host "localhost" eine Organisation "Fremdzugriff" anlege
    Dann wird der Schreibzugriff mit HTTP 403 abgewiesen
    Und es wurde keine Organisation "Fremdzugriff" angelegt

  @gate-story-11
  Szenario: Wildcard-Listener erlaubt keine lokalen Schreibzugriffe
    Angenommen der installierte Go-Server ist mit neuer Datenbank auf allen Netzadressen gestartet
    Wenn ich über Loopback ohne Origin mit Host "localhost" eine Organisation "Wildcard-Zugriff" anlege
    Dann wird der Schreibzugriff mit HTTP 403 abgewiesen
    Und es wurde keine Organisation "Wildcard-Zugriff" angelegt
