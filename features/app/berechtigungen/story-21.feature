# language: de
@story-21 @perm-04
Funktionalität: Privates Codex-CLI-Arbeitsprofil mit gesperrtem Start
  Als lokale Betreiberin
  möchte ich den serverseitig abgeleiteten Arbeitsbereich und vermittelte Schreibrechte konfigurieren,
  damit kein Agent über einen Clientpfad oder eine unbelegte CLI-Grenze Zugriff erhält.

  Grundlage:
    Angenommen ich starte die Anwendung mit einer neuen SQLite-Datenbank und einem privaten Datenverzeichnis
    Und die Organisation "Nordstern" mit dem Codex-CLI-Agenten "Kai" besteht

  Szenario: Neuer Agent hat weder Arbeitsbereich noch Schreibrecht aktiviert
    Wenn ich Kais Codex-Profil über die öffentliche API lese
    Dann ist der Arbeitsbereich ausschließlich serverseitig abgeleitet und gehört zu "Nordstern" und "Kai"
    Und Arbeitsbereich und Schreiben sind deaktiviert
    Und das Profil enthält keine vom Client gesetzten Pfade oder direkten Systemfähigkeiten
    Und Kai ist nicht startbereit

  Szenario: Betreiberin aktiviert den Arbeitsbereich ohne stilles Schreibrecht
    Wenn ich Kais Codex-Profil mit aktiviertem Arbeitsbereich und deaktiviertem Schreiben über die öffentliche API speichere
    Dann zeigt das gespeicherte Profil den aktivierten Arbeitsbereich und deaktiviertes Schreiben
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte
    Dann zeigt Kais Codex-Profil weiterhin den aktivierten Arbeitsbereich und deaktiviertes Schreiben
    Und Kai ist wegen des fehlenden CLI-Durchsetzungsnachweises nicht startbereit

  Szenario: Vermitteltes Schreiben braucht einen aktivierten Arbeitsbereich
    Wenn ich Kais Codex-Profil mit deaktiviertem Arbeitsbereich und aktiviertem Schreiben über die öffentliche API speichere
    Dann wird die Konfiguration am Feld "write_enabled" ohne Änderung abgewiesen
    Und Arbeitsbereich und Schreiben sind deaktiviert

  Szenariogrundriss: Ein Client kann keinen Pfad und keine Systemfähigkeit einschleusen
    Wenn ich Kais Codex-Profil mit dem zusätzlichen JSON-Feld <Feld> über die öffentliche API speichere
    Dann wird die Konfiguration ohne Änderung abgewiesen
    Und Arbeitsbereich und Schreiben sind deaktiviert

    Beispiele:
      | Feld                                      |
      | "workspace_path":"../fremd"             |
      | "workspace_path":"/tmp/fremd"           |
      | "capabilities":["filesystem_write"]    |

  Szenario: Ein übergroßer JSON-Körper ändert das Profil nicht
    Wenn ich Kais Codex-Profil mit gültigen Schaltern und mehr als einem MiB nachfolgendem Leerraum speichere
    Dann wird der übergroße JSON-Körper mit HTTP 400 abgewiesen
    Und Arbeitsbereich und Schreiben sind deaktiviert

  Szenario: Eine fremde Agentenkennung gibt weder Profil noch Arbeitsbereich preis
    Angenommen die Organisation "Südlicht" besteht
    Wenn ich Kais Codex-Profil über die URL von "Südlicht" lese
    Dann entspricht die datenfreie Ablehnung der Antwort auf eine unbekannte Agentenkennung
    Wenn ich Kais Codex-Profil über die URL von "Südlicht" ändere
    Dann entspricht die datenfreie Ablehnung der Antwort auf eine unbekannte Agentenkennung
    Und Kais Profil in "Nordstern" ist unverändert

  Szenariogrundriss: Nicht vertrauenswürdiger Schreibaufruf ändert das Profil nicht
    Wenn <Aufruf> Kais Codex-Profil ändert
    Dann antwortet die Anwendung mit HTTP 403 ohne Profildaten
    Und Arbeitsbereich und Schreiben sind deaktiviert

    Beispiele:
      | Aufruf                              |
      | eine uneindeutige Betreiberidentität |
      | ein fremder Browser-Ursprung        |

  Szenario: Ein Symlink im abgeleiteten Arbeitsbereich sperrt die Nutzung
    Angenommen der serverseitig abgeleitete Arbeitsbereich für Kai enthält einen Symlink nach außerhalb
    Wenn ich Kais Codex-Profil mit aktiviertem Arbeitsbereich und deaktiviertem Schreiben über die öffentliche API speichere
    Dann wird der Arbeitsbereich wegen verletzter Integrität ohne Schreibwirkung abgewiesen
    Und Kai ist nicht startbereit

  Szenario: Eine Profilfreigabe ersetzt keinen CLI-Durchsetzungsnachweis
    Wenn ich Kais Codex-Profil mit aktiviertem Arbeitsbereich und aktiviertem Schreiben über die öffentliche API speichere
    Und ich Kais Bereitschaft über die öffentliche Agenten-API abfrage
    Dann ist Kai wegen des fehlenden CLI-Durchsetzungsnachweises nicht startbereit
    Wenn ich Kai über die öffentliche Agenten-API starte
    Dann wird der Start mit HTTP 409 ohne Laufkennung verweigert

  @story-27 @pending-real-run
  Szenario: Ein echter CLI-Lauf darf weder Pfadausbruch noch direkten Schreibzugriff ausführen
    Angenommen ein nachweisbar begrenztes Codex-CLI-Profil und ein öffentlicher Aufgabenlauf sind verfügbar
    Wenn Kai im öffentlichen Lauf einen Pfadausbruch oder direkten Schreibzugriff anfordert
    Dann wird der Aufruf ohne Dateiänderung und ohne Scheinabschluss verweigert
