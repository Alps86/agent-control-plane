# MS-09: öffentliche Akzeptanzbeispiele zur Anbieter-Erweiterbarkeit

Stand: 24. September 2026. Dies sind Planungsbeispiele, keine fertigen Feature-Dateien und keine bestandenen Tests. Bei Umsetzung werden die öffentlichen Anwendungsfälle unter `features/app/modelle-sprache/` und sichtbare Bedienfälle unter `features/ui/modelle-sprache/` konkretisiert. Ein registrierter Testadapter und externe Anbieterkonfiguration dienen nur als Testgegenstelle; die fachliche Abnahme prüft öffentliche Wirkung statt interne Klassen.

## Auswahl aus Konfiguration

**Gegeben:** Ein registrierter Testanbieter ist über externe Anbieterkonfiguration aktiviert; Authentifizierungsart, Verbindung, Modelle und geprüfte Fähigkeiten sind dort angegeben.

**Wenn:** Die Bedienperson den Modellzugang eines Eino-Agenten öffnet.

**Dann:** Der Testanbieter erscheint mit den konfigurierten Modellen und Fähigkeiten. Eine zulässige Verbindung und ein Modell lassen sich auswählen; die Ausführungsart bleibt Eino.

## Berechtigter Aufgabenlauf

**Gegeben:** Die Testverbindung ist für Organisation und Agent freigegeben. Der Eino-Agent hat ein zulässiges Modell sowie die erforderlichen Daten- und Toolrechte.

**Wenn:** Die Bedienperson eine zugewiesene Aufgabe mit diesem Agenten startet.

**Dann:** Modellantwort, Resultat, ausgewählter Anbieter, Verbindung und Modell sind im öffentlichen Lauf nachvollziehbar.

## Fehlende Verbindungsgenehmigung

**Gegeben:** Der Agent hat den Testanbieter konfiguriert, aber die Verbindung ist für seine Organisation oder ihn nicht freigegeben.

**Wenn:** Die Bedienperson die zugewiesene Aufgabe startet.

**Dann:** Die Modellverwendung wird mit verständlichem Grund verweigert; die Testgegenstelle erhält keinen Modellaufruf.

## Nicht freigegebenes Tool

**Gegeben:** Der Testanbieter schlägt im Aufgabenlauf ein Tool vor, für das der Eino-Agent keine Freigabe besitzt.

**Wenn:** Der Agent den Toolaufruf verarbeiten soll.

**Dann:** Die Toolausführung wird mit verständlichem Grund verweigert; der Lauf behauptet keine ausgeführte Toolaktion.

## Anbieterfehler ohne Ersatzroute

**Gegeben:** Ein Eino-Agent verwendet den konfigurierten Testanbieter.

**Wenn:** Die Testgegenstelle den Modellaufruf mit einem Fehler beantwortet.

**Dann:** Der Fehler ist am Lauf sichtbar; es erfolgt kein Aufruf über Codex-Abo, OpenRouter oder ein anderes Modell.

Die frühere Datei `spec/features/modelle-sprache/anbieter-erweiterbarkeit.feature` wurde als [Markdown-Planungsbeispiel](../akzeptanzbeispiele/bestand/modelle-sprache/anbieter-erweiterbarkeit.md) gesichert; der alte Feature-Pfad besteht im Arbeitsbaum nicht mehr.
