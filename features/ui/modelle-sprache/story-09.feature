# language: de
@story-09 @ms-02 @browser @experimental
Funktionalität: Codex-Abo in den Modelleinstellungen verbinden
  Als Betreiber
  möchte ich die Gerätecode-Anmeldung in den Modelleinstellungen bedienen,
  damit ich den aktuellen Verbindungsstatus und den nächsten Schritt erkenne.
  Ein echter Anbieter-Login bleibt ein separates Freigabekriterium.

  Szenario: Gerätecode und Anmeldeseite während eines aktiven Versuchs anzeigen
    Angenommen ich öffne die Modelleinstellungen ohne Codex-Abo-Verbindung
    Wenn ich "Codex-Abo verbinden" wähle
    Dann sehe ich den Anbieterlink und den Gerätecode
    Und ich sehe den Status "Anmeldung läuft"
    Und die Seite zeigt keine Zugangstokens oder Konto-ID

  Szenario: Verbindungsstatus bis zur Bestätigung verfolgen
    Angenommen eine Gerätecode-Anmeldung läuft in den Modelleinstellungen
    Wenn ich den Verbindungsstatus aktualisiere
    Dann sehe ich weiterhin "Anmeldung läuft"
    Wenn der Anbieter die Anmeldung bestätigt und ich den Status aktualisiere
    Dann sehe ich "Verbunden"
    Und Anbieterlink und Gerätecode werden nicht mehr angezeigt
    Und die Seite zeigt keine Zugangstokens oder Konto-ID

  Szenario: Laufende Anmeldung abbrechen
    Angenommen eine Gerätecode-Anmeldung läuft in den Modelleinstellungen
    Wenn ich "Anmeldung abbrechen" wähle
    Dann sehe ich "Anmeldung abgebrochen"
    Und Anbieterlink und Gerätecode werden nicht mehr angezeigt
    Und die Verbindung wird nicht als einsatzbereit angezeigt

  Szenariogrundriss: Fehler und Ablauf verständlich anzeigen
    Angenommen eine Gerätecode-Anmeldung läuft in den Modelleinstellungen
    Wenn der Anbieter die Anmeldung mit "<Ergebnis>" beendet und ich den Status aktualisiere
    Dann sehe ich "<Hinweis>"
    Und Anbieterlink und Gerätecode werden nicht mehr angezeigt
    Und die Verbindung wird nicht als einsatzbereit angezeigt

    Beispiele:
      | Ergebnis   | Hinweis                         |
      | abgelaufen | Gerätecode abgelaufen           |
      | verweigert | Anmeldung nicht zugelassen      |

  Szenario: Deaktivierte Gerätecode-Freigabe erklären
    Angenommen der Anbieter hat die Gerätecode-Anmeldung deaktiviert
    Wenn ich in den Modelleinstellungen "Codex-Abo verbinden" wähle
    Dann sehe ich einen Hinweis zur fehlenden Anbieterfreigabe
    Und ich sehe keinen Gerätecode
    Und die Verbindung wird nicht als einsatzbereit angezeigt

  Szenario: Nach endgültigem Sitzungsfehler erneut verbinden
    Angenommen meine Codex-Abo-Verbindung war verbunden
    Wenn der Anbieter die Sitzung endgültig für ungültig erklärt und ich den Status aktualisiere
    Dann sehe ich "Erneute Anmeldung erforderlich"
    Und ich sehe keinen Gerätecode
    Wenn ich "Erneut verbinden" wähle
    Dann sehe ich einen neuen Anbieterlink und Gerätecode
    Und die Seite zeigt keine Zugangstokens oder Konto-ID

  Szenariogrundriss: Verspäteter Status zeigt keinen abgeschlossenen Gerätecode erneut
    Angenommen ich öffne die Modelleinstellungen ohne Codex-Abo-Verbindung
    Wenn im Browser eine verspätete Anmeldung-läuft-Antwort nach "<Abschluss>" eintrifft
    Dann bleibt im Browser der Status "<Status>" ohne Gerätecode sichtbar

    Beispiele:
      | Abschluss    | Status                |
      | Bestätigung  | Verbunden             |
      | Abbruch      | Anmeldung abgebrochen |
