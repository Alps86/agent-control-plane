# Modellwahl

Quelle: `spec/features/modelle-sprache/modellwahl.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: OpenRouter bewusst für Eino-Agenten wählen

  # MS-03; ergänzt bestehende Codex-Gerätecode-Szenarien unter integrationen/modellanbieter.feature.
  Szenario: Zentralen OpenRouter-Key in Settings hinterlegen
    Angenommen keine OpenRouter-Verbindung ist eingerichtet
    Wenn ich unter "Settings > Modellanbieter" einen OpenRouter-Key eingebe und die Verbindung speichere
    Dann sehe ich den Verbindungsstatus für OpenRouter
    Und ich sehe einen Hinweis auf die separate Abrechnung durch OpenRouter
    Und der Schlüssel erscheint weder in der Einstellungsansicht noch in ihrer öffentlichen Antwort

  # MS-03, MS-07
  Szenariogrundriss: Nicht einsatzbereite OpenRouter-Verbindung erkennen
    Angenommen ich richte die zentrale OpenRouter-Verbindung ein
    Wenn der Schlüssel "<Zustand>" ist
    Dann wird die Verbindung als nicht einsatzbereit angezeigt
    Und ich sehe "<Hinweis>"

    Beispiele:
      | Zustand  | Hinweis                    |
      | fehlend  | Schlüssel fehlt            |
      | ungültig | Verbindung nicht bestätigt |

  # MS-04
  Szenario: Zentrale Verbindung für Organisation und bestimmten Agenten freigeben
    Angenommen die zentrale OpenRouter-Verbindung enthält einen gültigen Key
    Und "Mira" und "Nora" sind Eino-Agenten der Organisation "Nord"
    Wenn ich die Verbindung erst für "Nord" und danach für "Mira" freigebe
    Dann sehe ich in Settings die Verbindungsreferenz und beide Freigabestatus ohne den Key
    Und "Mira" kann OpenRouter als Modellanbieter wählen
    Und "Nora" kann die Verbindung nicht wählen

  # MS-04, MS-05; Codex-Abo-Standard ist bereits im bestehenden Anbieter-Feature beschrieben.
  Szenario: OpenRouter für genau einen Eino-Agenten auswählen
    Angenommen die zentrale OpenRouter-Verbindung ist einsatzbereit
    Und die Verbindung ist für die Organisation von "Mira" freigegeben
    Und "Mira" und "Nora" sind Eino-Agenten mit dem Anbieter "Codex-Abo"
    Und die Verbindung ist für "Mira" freigegeben, für "Nora" aber nicht
    Wenn ich für "Mira" ausdrücklich "OpenRouter" und ein verfügbares Modell auswähle
    Dann sehe ich bei "Mira" den Anbieter "OpenRouter" und das gewählte Modell
    Und "Mira" hat weiter die Ausführungsart "Eino"
    Und "Nora" verwendet weiter ihren bisherigen Anbieter

  # MS-04
  Szenario: Nicht freigegebene zentrale Verbindung nicht auswählen
    Angenommen die zentrale OpenRouter-Verbindung enthält einen gültigen Key
    Und "Mira" ist ein Eino-Agent in der Organisation "Nord"
    Und die Verbindung ist nicht für die Organisation "Nord" freigegeben
    Wenn ich OpenRouter für "Mira" auswählen möchte
    Dann kann ich die Verbindung "Mira" nicht zuordnen
    Und ich sehe einen Hinweis auf die fehlende Organisationsfreigabe

  # MS-04
  Szenariogrundriss: Zentrale Verbindung ohne erforderliche Freigabe nicht verwenden
    Angenommen die zentrale OpenRouter-Verbindung enthält einen gültigen Key
    Und "Mira" ist ein Eino-Agent in der Organisation "Nord"
    Und die Freigabe der Verbindung für "<Ebene>" fehlt
    Und "Mira" ist der Verbindung bereits zugeordnet
    Wenn ich einen neuen Lauf von "Mira" starte
    Dann wird die Verwendung der Verbindung mit einem Hinweis auf die fehlende Freigabe blockiert
    Und der Key wird nicht für einen Modellaufruf verwendet
    Und es erfolgt kein stiller Wechsel zu einer anderen Modellverbindung

    Beispiele:
      | Ebene                |
      | Organisation "Nord" |
      | Agent "Mira"        |

  # MS-04
  Szenario: Freigabe einer genutzten OpenRouter-Verbindung widerrufen
    Angenommen die zentrale OpenRouter-Verbindung ist für die Organisation "Nord" und ihren Eino-Agenten "Mira" freigegeben
    Und "Mira" verwendet diese Verbindung
    Wenn ich die Freigabe für "Mira" widerrufe
    Dann bleibt der zentrale Key eingerichtet
    Und neue und geplante Läufe von "Mira" verwenden die Verbindung nicht mehr
    Und ich sehe die fehlende Agentenfreigabe beim betroffenen Lauf
    Und andere ausdrücklich freigegebene Agenten können die Verbindung weiter verwenden

  # MS-05, VO-01
  Szenario: Nicht belegte Sprachfähigkeit kennzeichnen
    Angenommen ich wähle ein OpenRouter-Modell für "Mira"
    Und die bidirektionale Echtzeitfähigkeit dieser Modellroute ist nicht nachgewiesen
    Wenn ich die Modellfähigkeiten öffne
    Dann sehe ich die einzeln geprüften Fähigkeiten für Text, Tools, Audioeingabe und Audioausgabe
    Und bidirektionale Echtzeit wird als "Nicht nachgewiesen" angezeigt

  # MS-06
  Szenario: Kein stiller Routenwechsel bei nicht unterstützter Fähigkeit
    Angenommen "Mira" verwendet eine ausgewählte OpenRouter-Modellroute
    Wenn diese Route den für ihren Lauf erforderlichen Tool- oder Audioparameter nicht unterstützt
    Dann sehe ich einen verständlichen Fehler beim Lauf
    Und der Lauf verwendet kein anderes Modell und keine andere Provider-Route ohne meine Auswahl

  # MS-06
  Szenario: Anbieterlimit ohne stillen kostenpflichtigen Ersatz behandeln
    Angenommen "Mira" verwendet die ausgewählte OpenRouter-Modellroute
    Wenn OpenRouter den Aufruf wegen eines Limits ablehnt
    Dann sehe ich den Limitgrund beim Lauf
    Und es erfolgt kein Modellaufruf über einen anderen Anbieter oder einen separat konfigurierten API-Zugang

  # MS-07
  Szenario: OpenRouter-Key ersetzen und nach Neustart weiterverwenden
    Angenommen "Mira" verwendet die zentrale OpenRouter-Verbindung
    Wenn ich ihren Key in Settings ersetze und die Anwendung neu starte
    Dann bleibt "Mira" dieser Verbindung zugeordnet
    Und neue Läufe verwenden nur den neuen, weiterhin gültigen Key
    Und der alte und der neue Key werden in keiner Laufansicht angezeigt

  # MS-07
  Szenario: Parallele Eino-Läufe nutzen dieselbe gültige Verbindung geordnet
    Angenommen "Mira" und "Nora" verwenden dieselbe gültige Modellverbindung
    Wenn beide Agenten gleichzeitig einen Lauf starten
    Dann sehe ich für jeden Lauf einen eigenen Status und ein Ergebnis oder einen verständlichen Fehler
    Und eine nötige Sitzungserneuerung verursacht keine widersprüchlichen Verbindungszustände
    Und Zugangsdaten erscheinen in keinem der beiden Läufe

  # MS-07
  Szenario: Zentrale OpenRouter-Verbindung trennen
    Angenommen "Mira" verwendet die zentrale OpenRouter-Verbindung
    Wenn ich die Verbindung in Settings trenne
    Dann sehe ich "Verbindung getrennt"
    Und ein neuer Lauf von "Mira" verlangt eine eingerichtete Verbindung
    Und er wechselt nicht zum Codex-Abo

  # MS-08
  Szenario: Modellzugang im Lauf nachvollziehen ohne Geheimnisse offenzulegen
    Angenommen "Mira" verwendet OpenRouter und darf das Projekt "Website" lesen
    Wenn ich einen abgeschlossenen Lauf von "Mira" öffne
    Dann sehe ich Ausführungsart, Anbieter, Modell und verwendete Route
    Und ich sehe keine Schlüssel oder Tokens in Laufdaten, Agentennachrichten oder Tooldaten
```
