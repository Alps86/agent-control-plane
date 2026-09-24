# language: de
@story-09 @ms-02 @experimental
Funktionalität: Codex-Abo per Gerätecode verbinden
  Als Betreiber
  möchte ich die Gerätecode-Anmeldung über die Modelleinstellungen verfolgen,
  damit ich den Verbindungszustand und nötige nächste Schritte erkenne.
  Diese Szenarien verwenden einen simulierten Anbieter. Ein echter Konto-Login bleibt ein separates Freigabekriterium.

  Szenario: Gerätecode starten und laufenden Versuch verfolgen
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich die Gerätecode-Anmeldung starte
    Dann erhalte ich die Anmeldeseite des Anbieters und einen Gerätecode
    Und die Verbindung ist noch nicht einsatzbereit
    Wenn ich den Verbindungsstatus prüfe
    Dann wird der Versuch als laufend angezeigt

  Szenario: Erfolgreiche Anmeldung als verbundene Referenz anzeigen
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich die Anmeldung beim Anbieter erfolgreich abschließe
    Und ich den Verbindungsstatus prüfe
    Dann wird die Codex-Abo-Verbindung als verbunden angezeigt
    Und die Antwort enthält keine Anmeldegeheimnisse
    Und nur der serverseitige Zugangs-Port kann Token und Konto-ID zusammen auflösen
    Und das Tokenbündel liegt verschlüsselt im Verbindungsspeicher

  Szenario: Laufenden Gerätecode-Versuch abbrechen
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich den Verbindungsversuch abbreche
    Dann wird die Anmeldung als abgebrochen angezeigt
    Und die Verbindung ist noch nicht einsatzbereit

  Szenario: Anbieter bestätigt Anmeldung unmittelbar vor dem Abbruch
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn der Anbieter die Anmeldung unmittelbar vor meinem Abbruch bestätigt
    Und ich den Verbindungsversuch abbreche
    Dann wird die Codex-Abo-Verbindung als verbunden angezeigt
    Und der bestätigte Vorgang wird beim Anbieter nicht abgebrochen

  Szenario: Verbundene Sitzung bei erneutem Startversuch behalten
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn ich die Gerätecode-Anmeldung erneut starte
    Dann wird die Codex-Abo-Verbindung als verbunden angezeigt
    Und ich erhalte keinen neuen Gerätecode vom Anbieter
    Wenn ich den Verbindungsversuch abbreche
    Dann wird die Codex-Abo-Verbindung als verbunden angezeigt
    Und der serverseitige Zugangs-Port liefert weiterhin Token und Konto-ID

  Szenario: Abgelaufenen Gerätecode erklären
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn der Gerätecode beim Anbieter abgelaufen ist
    Und ich den Verbindungsstatus prüfe
    Dann wird der Gerätecode als abgelaufen angezeigt
    Und die Verbindung ist noch nicht einsatzbereit

  Szenario: Deaktivierte Gerätecode-Freigabe erklären
    Angenommen der Anbieter hat die Gerätecode-Anmeldung deaktiviert
    Wenn ich die Gerätecode-Anmeldung starte
    Dann sehe ich einen Hinweis zur fehlenden Anbieterfreigabe
    Und die Verbindung ist noch nicht einsatzbereit

  Szenario: Verweigerte Anmeldung erklären
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn der Anbieter die Anmeldung verweigert
    Und ich den Verbindungsstatus prüfe
    Dann wird die Anmeldung als verweigert angezeigt
    Und die Verbindung ist noch nicht einsatzbereit

  Szenario: Nach endgültigem Sitzungsfehler erneut anmelden
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn der Anbieter die Sitzung endgültig widerruft
    Und ich den Verbindungsstatus prüfe
    Dann sehe ich, dass eine erneute Anmeldung erforderlich ist
    Wenn ich die Gerätecode-Anmeldung erneut starte
    Und ich die Anmeldung beim Anbieter erfolgreich abschließe
    Und ich den Verbindungsstatus prüfe
    Dann wird die Codex-Abo-Verbindung als verbunden angezeigt
    Und die Antwort enthält keine Anmeldegeheimnisse

  Szenario: Endgültigen Sitzungsfehler des Anbieters fachlich anzeigen
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn der Anbieter eine erneute Anmeldung ausdrücklich verlangt
    Und ich den Verbindungsstatus prüfe
    Dann sehe ich, dass eine erneute Anmeldung erforderlich ist
    Und die Antwort enthält keine Anmeldegeheimnisse

  Szenario: Bald ablaufendes Zugangstoken serverseitig erneuern
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn das Zugangstoken kurz vor Ablauf steht und der Anbieter neue Tokens liefert
    Und der serverseitige Zugangs-Port Token und Konto-ID anfordert
    Dann erhält nur der serverseitige Port das erneuerte Token mit derselben Konto-ID
    Und die Antwort enthält keine Anmeldegeheimnisse

  Szenario: Endgültig verweigerte Erneuerung verlangt erneute Anmeldung
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn der Anbieter die Token-Erneuerung mit "invalid_grant" verweigert
    Und der serverseitige Zugangs-Port Token und Konto-ID anfordert
    Und ich den Verbindungsstatus prüfe
    Dann sehe ich, dass eine erneute Anmeldung erforderlich ist
    Und die Antwort enthält keine Anmeldegeheimnisse

  Szenario: Erneuerung ohne frisches Zugangstoken ablehnen
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn der Anbieter bei der Erneuerung kein frisches Zugangstoken liefert
    Und der serverseitige Zugangs-Port Token und Konto-ID anfordert
    Dann liefert der Port weder Token noch Konto-ID
    Und das bisherige Tokenbündel wird nicht durch ein neues ersetzt
    Wenn ich den Verbindungsstatus prüfe
    Dann sehe ich, dass eine erneute Anmeldung erforderlich ist

  Szenario: Erneuerung für ein anderes Konto ablehnen
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn der Anbieter bei der Erneuerung ein Zugangstoken für ein anderes Konto liefert
    Und der serverseitige Zugangs-Port Token und Konto-ID anfordert
    Dann liefert der Port weder Token noch Konto-ID
    Und das bisherige Tokenbündel wird nicht durch ein neues ersetzt
    Wenn ich den Verbindungsstatus prüfe
    Dann sehe ich, dass eine erneute Anmeldung erforderlich ist

  Szenario: Erneuerung ohne neues ID-Token für dasselbe Konto zulassen
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn der Anbieter ein neues Zugangstoken für dasselbe Konto ohne ID-Token liefert
    Und der serverseitige Zugangs-Port Token und Konto-ID anfordert
    Dann erhält nur der serverseitige Port das erneuerte Token mit derselben Konto-ID
    Und das erneuerte Tokenbündel wird geschützt gespeichert

  Szenario: Tokens bleiben aus Oberfläche Protokollen und UI-Fixtures heraus
    Angenommen die Codex-Abo-Verbindung war zuvor gültig
    Wenn ich den Verbindungsstatus prüfe
    Dann enthalten Statusantwort Protokoll und UI-Fixtures keine Zugangstokens

  Szenario: Geschützte Verbindung nach erneutem Öffnen verwenden
    Angenommen ein geschützter lokaler Verbindungsspeicher ist geöffnet
    Wenn ich die Codex-Anmeldedaten speichere
    Dann stehen die Anmeldedaten nicht im Klartext in der Secret-Datei
    Wenn ich den Speicher mit demselben Masterschlüssel erneut öffne
    Dann erhalte ich die gespeicherten Anmeldedaten zurück

  Szenario: Fehlenden Masterschlüssel nicht durch Klartext ersetzen
    Angenommen ich habe Codex-Anmeldedaten geschützt gespeichert
    Wenn der Masterschlüssel fehlt
    Und ich den Verbindungsspeicher erneut öffne
    Dann wird der Verbindungsspeicher nicht geöffnet
    Und die Secret-Datei enthält keine Anmeldedaten im Klartext

  Szenariogrundriss: Zu weit freigegebene Credential-Datei ablehnen
    Angenommen ich habe Codex-Anmeldedaten geschützt gespeichert
    Wenn die Berechtigungen der "<Datei>" zu weit geöffnet werden
    Und ich den Verbindungsspeicher erneut öffne
    Dann wird der Verbindungsspeicher nicht geöffnet
    Und die Secret-Datei enthält keine Anmeldedaten im Klartext

    Beispiele:
      | Datei           |
      | Masterschlüssel |
      | Secret-Datei    |

  Szenario: Fremde Browser-Herkunft startet keine Gerätecode-Anmeldung
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich die Gerätecode-Anmeldung mit fremder Browser-Herkunft starte
    Dann wird die Settings-Anfrage untersagt
    Und beim Anbieter wurde kein Gerätecode angefordert
    Und die Antwort enthält keine Anmeldegeheimnisse

  Szenario: Fremder Host startet keine Gerätecode-Anmeldung
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich die Gerätecode-Anmeldung mit fremdem Host starte
    Dann wird die Settings-Anfrage untersagt
    Und beim Anbieter wurde kein Gerätecode angefordert

  Szenario: Fremde Browser-Herkunft bricht keine Anmeldung ab
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich den Verbindungsversuch mit fremder Browser-Herkunft abbreche
    Dann wird die Settings-Anfrage untersagt
    Und der Verbindungsversuch bleibt laufend

  Szenario: Fremder Host bricht keine Anmeldung ab
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich den Verbindungsversuch mit fremdem Host abbreche
    Dann wird die Settings-Anfrage untersagt
    Und der Verbindungsversuch bleibt laufend

  Szenario: Ohne Browser-Herkunft keine Anmeldung starten
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich die Gerätecode-Anmeldung ohne Browser-Herkunft starte
    Dann wird die Settings-Anfrage untersagt
    Und beim Anbieter wurde kein Gerätecode angefordert

  Szenario: Ohne Browser-Herkunft keine Anmeldung abbrechen
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich den Verbindungsversuch ohne Browser-Herkunft abbreche
    Dann wird die Settings-Anfrage untersagt
    Und der Verbindungsversuch bleibt laufend

  Szenario: Fremder Host liest keinen laufenden Gerätecode
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich den Verbindungsstatus mit fremdem Host prüfe
    Dann wird die Settings-Anfrage untersagt
    Und die Antwort enthält keinen Gerätecode
    Und der Verbindungsversuch bleibt laufend

  Szenario: Entfernter Peer startet trotz lokaler Header keine Anmeldung
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich die Gerätecode-Anmeldung von einem entfernten Peer mit lokalen Headern starte
    Dann wird die Settings-Anfrage untersagt
    Und beim Anbieter wurde kein Gerätecode angefordert

  Szenario: Entfernter Peer bricht trotz lokaler Header keine Anmeldung ab
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich den Verbindungsversuch von einem entfernten Peer mit lokalen Headern abbreche
    Dann wird die Settings-Anfrage untersagt
    Und der Verbindungsversuch bleibt laufend

  Szenario: Entfernter Peer liest trotz lokaler Header keinen Gerätecode
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn ich den Verbindungsstatus von einem entfernten Peer mit lokalen Headern prüfe
    Dann wird die Settings-Anfrage untersagt
    Und die Antwort enthält keinen Gerätecode
    Und der Verbindungsversuch bleibt laufend

  Szenariogrundriss: Unsichere Listener-Bind-Adresse vor Browser-Anfragen ablehnen
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich den Settings-Handler mit der Bind-Adresse "<Adresse>" initialisiere
    Dann wird die unsichere Bind-Adresse abgelehnt
    Und beim Anbieter wurde kein Gerätecode angefordert

    Beispiele:
      | Adresse            |
      | 0.0.0.0:8080       |
      | [::]:8080          |
      | 192.0.2.1:8080     |
      | :8080              |
      | 127.0.0.1:ungueltig |

  Szenario: Lokaler Browser verwendet vertrauenswürdige Listener-Bind-Adresse
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich den Settings-Handler mit der Bind-Adresse "localhost:8080" initialisiere
    Und ich die Gerätecode-Anmeldung über localhost starte
    Dann erhalte ich die Anmeldeseite des Anbieters und einen Gerätecode

  Szenario: Weitere numerische Loopback-Adresse erlaubt Codex-Status und Start
    Angenommen die Codex-Abo-Verbindung ist nicht eingerichtet
    Wenn ich den Settings-Handler mit der Bind-Adresse "127.0.0.2:8080" initialisiere
    Und ich die Gerätecode-Anmeldung über 127.0.0.2 starte
    Dann erhalte ich die Anmeldeseite des Anbieters und einen Gerätecode
    Wenn ich den Verbindungsstatus über 127.0.0.2 prüfe
    Dann wird der Versuch als laufend angezeigt
