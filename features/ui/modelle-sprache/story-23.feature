# language: de
@story-23 @ms-03 @ui
Funktionalität: OpenRouter-Verbindung in Settings bedienen
  Als Betreiber
  möchte ich den Schlüssel verdeckt eingeben und den Status verstehen,
  damit ich die zusätzliche Abrechnung und die Verbindung bewusst verwalte.

  Grundlage:
    Angenommen der lokale Server verwendet einen neuen geschützten Credentials-Speicher
    Und ein kontrollierter OpenRouter-Testanbieter steht bereit

  Szenario: Separate Abrechnung vor dem Speichern erkennen
    Angenommen ich öffne "Settings > Modellanbieter" ohne OpenRouter-Verbindung
    Dann sehe ich vor der Schlüsseleingabe einen Hinweis auf die separate OpenRouter-Abrechnung
    Und die Oberfläche bietet OpenRouter als bewusste Verbindung an
    Und sie stellt OpenRouter nicht als automatischen Codex-Ersatz dar

  Szenario: Schlüssel verdeckt speichern und Status selbst prüfen
    Angenommen ich öffne "Settings > Modellanbieter" ohne OpenRouter-Verbindung
    Wenn ich einen OpenRouter-Schlüssel in das verdeckte Eingabefeld eingebe und speichere
    Dann zeigt die Oberfläche nur eine maskierte Verbindungsreferenz
    Und sie zeigt den Schlüssel weder im Eingabefeld noch im sichtbaren Seiteninhalt erneut an
    Wenn ich die Statusprüfung auslöse
    Dann sehe ich den verständlichen Verbindungsstatus des kontrollierten Testanbieters

  Szenario: Schlüssel ersetzen und Verbindung trennen
    Angenommen ich öffne "Settings > Modellanbieter" mit bestehender OpenRouter-Verbindung
    Wenn ich den Schlüssel über die Ersetzen-Aktion ändere
    Dann bleibt die zentrale Verbindungsreferenz sichtbar und der neue Schlüssel verborgen
    Wenn ich die Verbindung trenne
    Dann sehe ich "Verbindung getrennt" und kann keinen einsatzbereiten Status annehmen
    Und derselbe vollständige Bedienweg funktioniert in einem echten Browser
