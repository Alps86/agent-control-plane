# Voice interviews

Quelle: `spec/features/modelle-sprache/voice-interviews.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Sprachliche Anforderungsinterviews für spätere Fachmodule

  # VO-06: Die Fachmodule Project Discovery und GrillMe erhalten eigene Stories.
  Szenario: Project Discovery als mehrstufiges Sprachinterview
    Angenommen ein berechtigter Eino-Agent unterstützt geprüfte Sprach-Turns
    Wenn ich ein Project-Discovery-Interview per Sprache starte
    Dann stellt der Agent Rückfragen zu Ziel, Beteiligten und offenen Annahmen
    Und ich sehe Antworten und ungeklärte Punkte als prüfbaren Entwurf
    Und ohne meine Bestätigung wird kein Projekt angelegt oder geändert

  # VO-06
  Szenario: GrillMe hinterfragt Anforderungen kritisch
    Angenommen ein berechtigter Eino-Agent unterstützt geprüfte Sprach-Turns
    Wenn ich ein GrillMe-Interview zu einem Vorhaben per Sprache starte
    Dann fragt der Agent nach Risiken, Widersprüchen und fehlenden Entscheidungen
    Und ich sehe die offenen Entscheidungen getrennt von bestätigten Aussagen
    Und die Zusammenfassung wird erst nach meiner Prüfung übernommen

  # VO-06
  Szenario: Unterbrochenes Interview berechtigt fortsetzen
    Angenommen ich habe ein begonnenes Sprachinterview mit offenen Fragen
    Wenn ich es unterbreche und später berechtigt fortsetze
    Dann sehe ich die bisherigen bestätigten Antworten und offenen Fragen
    Und der Agent setzt ohne doppelte Fachaktion fort

  # VO-07: Gate für ein noch nicht belegtes OpenRouter-Realtime-Angebot.
  Szenario: Echte bidirektionale Echtzeit erst nach technischem Nachweis anbieten
    Angenommen für die gewählte OpenRouter-Modellroute ist keine bidirektionale Realtime-Verbindung nachgewiesen
    Wenn ich die Sprachmodi öffne
    Dann sehe ich keine als einsatzbereit bezeichnete bidirektionale Echtzeitoption
    Und ein geprüfter Sprach-Turn wird als eigener Modus angezeigt

  # VO-07
  Szenario: Unterbrechung und Toolrückgabe in nachgewiesener Echtzeitsitzung
    Angenommen eine OpenRouter-Realtime-Verbindung mit gleichzeitigem Audioein- und -ausgang ist technisch nachgewiesen
    Und "Mira" darf nur den Projektstatus lesen
    Wenn ich die hörbare Antwort unterbreche und eine neue Frage zum Projektstatus stelle
    Dann stoppt die laufende Wiedergabe
    Und "Mira" verarbeitet die neue Spracheingabe in derselben berechtigten Sitzung
    Und ein Toolaufruf wird von Eino geprüft und sein Ergebnis der Sitzung zurückgegeben

  # VO-07: Die folgenden Gate-Tests setzen eine erst noch zu belegende Kandidatenroute voraus.
  Szenario: Berechtigte Echtzeitsitzung mit erhaltenem Kontext fortsetzen
    Angenommen eine konkrete OpenRouter-Realtime-Kandidatenroute wird im technischen Gate geprüft
    Und ich habe mit "Mira" eine berechtigte Sitzung mit einer beantworteten Frage begonnen
    Wenn ich die Verbindung unterbreche und die Sitzung mit ihrer Referenz berechtigt fortsetze
    Dann ist die vorherige Frage im fortgesetzten Sitzungskontext verfügbar
    Und "Mira" beantwortet meine nächste Frage ohne einen neuen unverbundenen Dialog zu beginnen
    Und keine bereits ausgeführte Fachaktion wird erneut ausgelöst

  # VO-07
  Szenario: Echtzeitsitzung wirksam abbrechen
    Angenommen eine konkrete OpenRouter-Realtime-Kandidatenroute wird im technischen Gate geprüft
    Und eine berechtigte Sitzung mit "Mira" empfängt und sendet gerade Audio
    Wenn ich die Sitzung abbreche
    Dann enden Mikrofonübertragung und Wiedergabe
    Und der Anbietertransport bestätigt das Ende der Sitzung
    Und nach dem Abbruch beginnt kein neuer Toolaufruf aus dieser Sitzung
    Und die Oberfläche behauptet keine Rücknahme bereits ausgeführter Aktionen

  # VO-07
  Szenariogrundriss: Fehler einer Echtzeit-Kandidatenroute sichtbar behandeln
    Angenommen eine konkrete OpenRouter-Realtime-Kandidatenroute wird im technischen Gate geprüft
    Und eine berechtigte Sitzung mit "Mira" läuft
    Wenn "<Fehler>" während der Sitzung eintritt
    Dann sehe ich "<Hinweis>" mit dem tatsächlichen Sitzungszustand
    Und es erfolgt kein stiller Wechsel zu einem anderen Modell oder Anbieter
    Und eine unklare Fachaktion wird vor einem erneuten Versuch anhand ihrer Auftragskennung geklärt

    Beispiele:
      | Fehler                           | Hinweis                         |
      | die Verbindung abbricht          | Verbindung unterbrochen         |
      | der Anbieter die Anfrage ablehnt | Anbieterfehler                  |
      | der OpenRouter-Key ungültig wird | OpenRouter-Key ersetzen         |
```
