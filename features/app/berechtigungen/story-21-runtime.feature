# language: de
@story-21 @perm-04 @runtime-proof @requires-bwrap
Funktionalität: Tatsächliche Codex-CLI-Grenze über eine öffentliche Profilprüfung nachweisen
  Als lokale Betreiberin
  möchte ich ein festes, serverseitiges Prüfprofil mit einer synthetischen Gegenstelle ausführen,
  damit Startbereitschaft nur aus beobachteten Tool- und Arbeitsbereichsgrenzen folgt.

  Grundlage:
    Angenommen ich starte die Anwendung mit einer neuen SQLite-Datenbank und einem privaten Datenverzeichnis
    Und die Organisation "Nordstern" mit dem Codex-CLI-Agenten "Kai" besteht
    Und die Runtime-Prüfung verwendet eine kontrollierte synthetische Gegenstelle ohne echte Zugangsdaten

  Szenario: Der Prüfaufruf nimmt keine ausführbaren Clientparameter an
    Wenn ich Kais Runtime-Prüfung mit einem Clientbefehl, Pfad oder URL anfordere
    Dann antwortet die öffentliche API mit HTTP 400 ohne CLI-Start und ohne Dateiänderung

  @runtime-positive
  Szenario: Feste benannte Go-Aktion schreibt nur ein privates Artefakt
    Angenommen Kais Arbeitsbereich und vermitteltes Schreiben sind aktiviert
    Und die synthetische Gegenstelle fordert ausschließlich die fest benannte Go-Artefaktaktion an
    Wenn ich Kais Runtime-Prüfung mit einem leeren JSON-Objekt über die öffentliche API anfordere
    Dann wurde die echte Codex CLI innerhalb der bwrap-Grenze gestartet
    Und die Prüfantwort nennt die ausgeführte benannte Fachaktion und ihren Nachweis
    Und nur im privaten Arbeitsbereich von "Nordstern" und "Kai" liegt das erwartete Artefakt
    Und der Prüfaufruf meldet keinen abgeschlossenen Aufgabenlauf

  @runtime-repeat
  Szenario: Dieselbe private Prüfaktion kann sicher wiederholt werden
    Angenommen Kais Arbeitsbereich und vermitteltes Schreiben sind aktiviert
    Und die synthetische Gegenstelle fordert ausschließlich die fest benannte Go-Artefaktaktion an
    Wenn ich Kais Runtime-Prüfung mit einem leeren JSON-Objekt über die öffentliche API anfordere
    Dann nennt die Prüfantwort die erfolgreich ausgeführte benannte Fachaktion
    Wenn ich Kais Runtime-Prüfung erneut mit einem leeren JSON-Objekt über die öffentliche API anfordere
    Dann nennt die Prüfantwort erneut die erfolgreich ausgeführte benannte Fachaktion
    Und im privaten Arbeitsbereich von "Nordstern" und "Kai" liegt genau ein unverändertes Prüfartefakt
    Und außerhalb dieses Arbeitsbereichs entstand kein Prüfartefakt

  Szenariogrundriss: Direkte System- und Netzaufrufe werden in der tatsächlichen CLI verweigert
    Angenommen Kais Arbeitsbereich ist aktiviert und Schreiben ist deaktiviert
    Und die synthetische Gegenstelle fordert <Nebenweg> an
    Wenn ich Kais Runtime-Prüfung mit einem leeren JSON-Objekt über die öffentliche API anfordere
    Dann wurde die echte Codex CLI innerhalb der bwrap-Grenze gestartet
    Und der Nebenweg <Nebenweg> wurde ohne Seitenwirkung verweigert
    Und die Prüfantwort enthält einen konkreten Nachweis der Verweigerung
    Und Kai bleibt ohne vollständigen Profilnachweis nicht startbereit

    Beispiele:
      | Nebenweg       |
      | "exec_command" |
      | "externer HTTP-Aufruf" |

  Szenariogrundriss: Fremde Daten und unsichere Artefaktpfade werden verweigert
    Angenommen Kais Arbeitsbereich ist aktiviert und Schreiben ist <Schreibrecht>
    Und die synthetische Gegenstelle fordert <Angriff> an
    Wenn ich Kais Runtime-Prüfung mit einem leeren JSON-Objekt über die öffentliche API anfordere
    Dann wurde die echte Codex CLI innerhalb der bwrap-Grenze gestartet
    Und <Angriff> wurde ohne Seitenwirkung verweigert
    Und die Prüfantwort enthält einen konkreten Nachweis der Verweigerung
    Und Kai bleibt ohne vollständigen Profilnachweis nicht startbereit

    Beispiele:
      | Angriff                         | Schreibrecht |
      | "fremde Organisation"          | aktiviert    |
      | "fremder Agent"                | aktiviert    |
      | "Pfad mit .."                  | aktiviert    |
      | "nicht freigegebenes Schreiben" | deaktiviert  |

  Szenario: Ein Symlink im privaten Arbeitsbereich stoppt schon die Runtime-Freigabe
    Angenommen Kais Arbeitsbereich und vermitteltes Schreiben sind aktiviert
    Und der private Arbeitsbereich für Kai enthält einen Symlink nach außerhalb
    Wenn ich Kais Runtime-Prüfung mit einem leeren JSON-Objekt über die öffentliche API anfordere
    Dann wird die Prüfung mit HTTP 409 ohne CLI-Start und ohne Dateiänderung verweigert

  Szenario: Ein fehlender Runtime-Nachweis gibt keine Startbereitschaft
    Angenommen Kais Arbeitsbereich ist aktiviert und Schreiben ist deaktiviert
    Und für Kais Profil kann die tatsächliche CLI-Grenze nicht vollständig geprüft werden
    Wenn ich Kais Runtime-Prüfung mit einem leeren JSON-Objekt über die öffentliche API anfordere
    Dann enthält die Prüfantwort "ready":false und "code":"codex_cli_enforcement_unverified"
    Und Kais öffentliche Bereitschaft bleibt "ready":false
    Und ein Start von Kai bleibt mit HTTP 409 ohne Laufkennung gesperrt

  @story-27 @pending-real-run
  Szenario: Die nachgewiesene Grenze gilt auch im späteren echten Aufgabenlauf
    Angenommen ein öffentlicher Aufgabenlauf mit demselben verifizierten Runtime-Profil ist verfügbar
    Wenn Kai im Aufgabenlauf einen direkten Systemweg anfordert
    Dann wird der Aufruf ohne Seitenwirkung und ohne Scheinabschluss verweigert
