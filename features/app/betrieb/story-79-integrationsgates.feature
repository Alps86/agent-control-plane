# language: de
@story-79 @ops-02 @integrationsgate
Funktionalität: Geheimnisgrenze an echten Lauf- und Exportwegen nachweisen
  Diese Szenarien werden nach Montage der jeweiligen öffentlichen Produktwege ausgeführt.
  Ein kontrollierter Modellanbieter ersetzt dabei keinen echten Aufgabenlauf oder Export.

  @story-27
  Szenario: Ein echter Lauf zeigt keine Zugangsdaten und wechselt bei Verbindungsverlust nicht den Anbieter
    Angenommen "Mira" hat eine ausführbare Aufgabe und verwendet ihre freigegebene OpenRouter-Verbindung
    Wenn ich die OpenRouter-Verbindung trenne und die Aufgabe mit "Mira" über die öffentliche Anwendung starte
    Dann wird der Lauf vor dem Modellanbieter mit einem Verbindungsfehler beendet
    Und die Laufansicht und ihre öffentliche Antwort enthalten weder Token noch Schlüssel
    Und kein anderer Modellanbieter wurde für diesen Lauf aufgerufen

  @story-26
  Szenario: Die echte Anbieterregistry startet bei Verbindungsverlust keinen Ersatzanbieter
    Angenommen "Mira" hat OpenRouter als feste Modellroute und zwei kontrollierte Anbieter sind registriert
    Wenn ich die OpenRouter-Verbindung trenne und eine Modellanfrage für "Mira" über die echte Routenbindung starte
    Dann erreicht weder OpenRouter noch der zweite Anbieter einen neuen Request
    Und die Antwort nennt die verlorene OpenRouter-Verbindung ohne Zugangsdaten

  @export
  Szenario: Ein echter Export enthält Referenzen und Status, aber keine Zugangsdaten
    Angenommen ein abgeschlossener Lauf und eine Agentenkonfiguration mit Modellverbindungsreferenz bestehen
    Wenn ich diese Daten über den öffentlichen Export herunterlade
    Dann enthält der Export nur die erlaubten Verbindungsreferenzen und Statusangaben
    Und er enthält weder Token noch Schlüssel noch Refresh-Token
