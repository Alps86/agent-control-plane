# Anbieter erweiterbarkeit

Quelle: `spec/features/modelle-sprache/anbieter-erweiterbarkeit.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@MS-09
Funktionalität: Registrierten Modellanbieter über Konfiguration einbinden
  Als Betreiberin möchte ich einen zusätzlich registrierten Anbieter im bestehenden Eino-Arbeitsweg nutzen.

  Szenario: Konfigurierten Testanbieter für einen Eino-Agenten wählen
    Angenommen ein registrierter Testanbieter ist über externe Anbieterkonfiguration aktiviert
    Und seine Verbindung, Authentifizierungsart, Modelle und geprüften Fähigkeiten sind konfiguriert
    Wenn ich den Modellzugang eines Eino-Agenten öffne
    Dann sehe ich den Testanbieter mit seinen konfigurierten Modellen und Fähigkeiten
    Und ich kann eine zulässige Verbindung und ein Modell auswählen
    Und die Ausführungsart bleibt "Eino"

  Szenario: Berechtigten Aufgabenlauf mit dem Testanbieter ausführen
    Angenommen der Testanbieter ist für die Organisation und den Eino-Agenten freigegeben
    Und der Agent hat ein zulässiges Modell und die erforderlichen Daten- und Toolrechte
    Wenn ich eine zugewiesene Aufgabe mit diesem Agenten starte
    Dann sehe ich eine Modellantwort und ein nachvollziehbares Ergebnis im Lauf
    Und Anbieter, Verbindung und Modell entsprechen meiner Auswahl

  Szenario: Fehlende Verbindungsfreigabe vor dem Anbieteraufruf verweigern
    Angenommen ein Eino-Agent verwendet den konfigurierten Testanbieter
    Und die Verbindung ist für seine Organisation oder für ihn nicht freigegeben
    Wenn ich seine zugewiesene Aufgabe starte
    Dann sehe ich die verweigerte Modellverwendung mit verständlichem Grund
    Und der Testanbieter erhält keinen Modellaufruf

  Szenario: Nicht freigegebenes Tool auch beim Testanbieter nicht ausführen
    Angenommen ein Eino-Agent verwendet den konfigurierten Testanbieter
    Und der Testanbieter schlägt im Aufgabenlauf ein nicht freigegebenes Tool vor
    Wenn der Agent den Toolaufruf bearbeiten soll
    Dann wird die Toolausführung mit verständlichem Grund verweigert
    Und der Aufgabenlauf behauptet keine ausgeführte Toolaktion

  Szenario: Fehler des Testanbieters ohne stillen Wechsel behandeln
    Angenommen ein Eino-Agent verwendet den konfigurierten Testanbieter
    Wenn der Testanbieter seinen Modellaufruf mit einem Fehler beantwortet
    Dann sehe ich den Anbieterfehler am betroffenen Lauf
    Und es erfolgt kein Aufruf über Codex-Abo, OpenRouter oder ein anderes Modell
```
