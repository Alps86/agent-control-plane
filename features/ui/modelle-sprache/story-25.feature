# language: de
@story-25 @ms-05 @ui
Funktionalität: Modellroute eines Eino-Agenten im Browser wählen
  Als Ersteller sehe ich verfügbare Anbieter, Verbindungen und Modelle.

  Grundlage:
    Angenommen ein lokaler Server mit kontrollierten Modellkatalogen läuft
    Und die Organisation "Nord" mit dem Eino-Agenten "Mira" besteht

  Szenario: Codex-Abo als Standard und Eino-Ausführungsart sehen
    Wenn ich Miras Modellwahl im echten Browser öffne
    Dann sehe ich "Codex-Abo" als vorausgewählten Anbieter
    Und ich sehe "Eino" als Ausführungsart
    Und die Seite zeigt nur nicht geheime Verbindungsreferenzen

  Szenario: OpenRouter erst nach beiden Freigaben wählen
    Angenommen die zentrale OpenRouter-Verbindung ist einsatzbereit und für "Nord" und "Mira" freigegeben
    Wenn ich im echten Browser OpenRouter und "openai/gpt-4" für "Mira" auswähle und speichere
    Dann sehe ich "OpenRouter" und "openai/gpt-4" in Miras Agentenansicht
    Und ich sehe "Eino" als Ausführungsart
    Und die Seite zeigt keine Zugangsdaten

  Szenario: Nicht belegte Fähigkeit sichtbar einordnen
    Angenommen die zentrale OpenRouter-Verbindung ist einsatzbereit und für "Nord" und "Mira" freigegeben
    Wenn ich Miras Modellfähigkeiten im echten Browser öffne
    Dann sehe ich Modellquelle, Stand und Prüfstatus für "openai/gpt-4"
    Und ich sehe den Textbeleg mit Quelle, Prüfzeit und Status
    Und ich sehe Text, Tools, Audioeingabe und Audioausgabe als getrennte Fähigkeiten
    Und bidirektionale Echtzeit ist als "Nicht nachgewiesen" markiert
