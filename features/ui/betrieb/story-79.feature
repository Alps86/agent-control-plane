# language: de
@story-79 @ops-02 @ui
Funktionalität: Verbindungen in Settings gemeinsam und redigiert überblicken
  Als Betreiber
  möchte ich den Status meiner Modellverbindungen an einer Stelle sehen,
  damit ich einen verlorenen Zugang vor dem nächsten Lauf erkenne.

  Grundlage:
    Angenommen ein lokaler Server mit kontrollierten Modellanbietern läuft

  @browser
  Szenario: Gemeinsame Settings-Ansicht zeigt beide Verbindungen ohne Geheimnisse
    Angenommen die zentrale OpenRouter-Verbindung wurde mit einem synthetischen Schlüssel eingerichtet und geprüft
    Wenn ich "Settings > Modellanbieter" im echten Browser öffne
    Dann sehe ich Codex-Abo und OpenRouter mit ihren getrennten Status und nur vorhandenen Verbindungsreferenzen
    Und ich sehe für Codex-Abo nur die Gerätecode-Anmeldung als Authentifizierungsart
    Und ich sehe für OpenRouter nur den API-Schlüssel als Authentifizierungsart
    Und weder der sichtbare Seiteninhalt noch der DOM enthalten Token, Schlüssel oder Refresh-Token

  @browser
  Szenario: Trennen wird in der gemeinsamen Ansicht und nach Neustart sichtbar
    Angenommen die gemeinsame Settings-Ansicht zeigt eine geprüfte OpenRouter-Verbindung
    Wenn ich OpenRouter über die vorhandene Settings-Aktion trenne
    Dann zeigt die gemeinsame Ansicht OpenRouter als getrennt und nicht einsatzbereit
    Und sie zeigt Codex-Abo weiterhin mit seinem tatsächlichen Status
    Wenn ich den Server mit denselben Daten- und Credential-Pfaden neu starte
    Und ich "Settings > Modellanbieter" im echten Browser öffne
    Dann bleibt OpenRouter als getrennt und nicht einsatzbereit sichtbar
    Und weder der sichtbare Seiteninhalt noch der DOM enthalten Token, Schlüssel oder Refresh-Token
