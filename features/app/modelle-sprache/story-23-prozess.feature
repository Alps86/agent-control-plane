# language: de
@story-23 @ms-03 @prozess
Funktionalität: Zentrale OpenRouter-Verbindung am gebauten Server dauerhaft bedienen
  Als Betreiber
  möchte ich den eingebauten Settings-Weg nach einem Neustart wiederfinden,
  damit die Installation ihre eine geschützte Verbindung behält.

  @prozess-kern
  Szenario: Synthetischen Schlüssel ohne Anbieteraufruf speichern, neu starten und trennen
    Angenommen ein gebauter lokaler Server mit temporärer SQLite-Datei und getrennten Credential- und Masterkey-Dateien ist gestartet
    Wenn ich die OpenRouter-Einstellungen über die öffentliche Settings-Seite öffne
    Dann erreiche ich die OpenRouter-Settings-Ansicht und ihr eingebautes Stylesheet
    Und der echte Browser zeigt die montierte Settings-Seite ohne Schlüssel im DOM
    Und die öffentliche JSON-Antwort zeigt "nicht eingerichtet"
    Wenn ich einen synthetischen OpenRouter-Schlüssel über das öffentliche Formular speichere
    Dann zeigen HTML und JSON dieselbe zentrale Verbindungsreferenz ohne den Schlüssel
    Und die geschützte Credential-Datei enthält den Schlüssel nicht im Klartext
    Wenn ich den Server mit denselben Daten- und Secret-Pfaden neu starte
    Dann zeigt die öffentliche JSON-Antwort dieselbe zentrale Verbindungsreferenz ohne den Schlüssel
    Wenn ich die Verbindung über die öffentliche Settings-API trenne
    Dann zeigt die öffentliche JSON-Antwort "nicht eingerichtet"
    Und es wurde keine OpenRouter-Statusprüfung ausgelöst

  @navigation
  Szenario: OpenRouter-Settings über die Organisationsnavigation erreichen
    Angenommen ein gebauter lokaler Server mit temporärer SQLite-Datei und getrennten Credential- und Masterkey-Dateien ist gestartet
    Wenn ich Settings über die öffentliche Navigation öffne
    Dann erreiche ich die OpenRouter-Settings-Ansicht und ihr eingebautes Stylesheet
