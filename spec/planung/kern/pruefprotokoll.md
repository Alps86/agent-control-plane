# Abgegrenzter Planungscheck

Stand 24.09.2026. Der unabhängige lesende Review prüfte die neuen Kern-Dateien gegen `spec/produkt.md`, `spec/technik.md`, `spec/paperclip-feature-map.md`, `spec/stories.md` und die vorhandenen Features. Danach wurden seine konkreten Befunde einmal gezielt behoben; es gab keine Produktimplementierung und keine Implementierungsreview.

| Befund | Behebung |
| --- | --- |
| Zielanlage und Projektanlage setzten sich gegenseitig voraus. | ORG-01 legt nur die Organisation an; GOAL-01 legt danach eigenständig das Ziel an; PROJ-01 verknüpft es. |
| OPS-01 war zugleich Startvoraussetzung und vollständige POC-Neustartabnahme. | ARC-02 liefert die technische SQLite-Basis; OPS-01 prüft den Start mit vorhandener/unzugänglicher DB; OPS-06 prüft nach RUN-01 den vollständigen Nutzungsweg und Neustart. |
| Aktivitäts- und Rechte-Abhängigkeiten standen entgegen der POC-Folge. | ACT-01 prüft früh Anlage/Zuweisung; PERM-02 baut darauf auf; ACT-03 ergänzt spätere Ereignisarten. Die POC-Folge wurde entsprechend geordnet. |
| AGT-01 setzte früher eine falsch zugeordnete Modellstory voraus, obwohl die Agentenanlage zunächst ohne Zugang möglich sein muss. | AGT-01 legt einen noch nicht lauffähigen Eino-Agenten an. MS-05 ordnet später Verbindung und konkretes Modell zu; davor gelten je nach bewusst gewähltem Pfad MS-02 oder MS-03/04. RUN-01 verlangt MS-05/06 und den gewählten Zugang. MOD-03 ist nur Legacy-Alias für MS-05. |
| Gherkin enthielt Neustart als `Dann`, eine nicht angelegte Aufgabe, vermischte DST-Jahreszeiten und ein `oder` bei ungültigen Eingaben. | Neustart und Taskanlage sind eigene `Wenn`-Schritte; Frühlings-/Herbstfall sind getrennt; Cron- und Zeitzonenfehler sind Beispiele des Szenariogrundrisses. |
| Quellenverzeichnis enthielt zwei veraltete URL-Pfade. | Connector-Access und Members-&-Access verweisen auf die tatsächlich erreichbaren offiziellen Seiten. |
| Alte Matrix deckte Work Modes, Execution Policy, Routine-Issue-Herkunft, Decisions und weitere Navigation nicht ab. | Quellenregister, 72 Story-IDs und Umfangsinventar führen diese Bereiche mit Ist-/Experiment-/Draft-Einstufung und offener Entscheidung. |

Abschließender statischer Check nach den Korrekturen: **72 eindeutige Kern-Story-IDs**, keine unbekannten internen Kern-Abhängigkeiten und kein Zyklus; **22 registrierte Quellenkürzel**, keine unregistrierte Referenz; **4 neue deutsche Feature-Dateien mit 36 Szenario-/Grundriss-Deklarationen**, alle Story-Tags referenzieren bekannte IDs; keine defekten lokalen Links und keine Wellen-/Batch-Merge-Gates im Katalog. Die technischen Vorgänger ARC-02/04 und die Modellstories MS-01 bis MS-09 sind externe Planungsreferenzen, keine Kern-Story-IDs. Die 14 zuvor vorhandenen Feature-Dateien wurden im Kernplanungsreview nicht geändert; der anschließende Integrationsabgleich präzisierte ein ZA-02-Szenario. Die neuen Szenarien sind Planung, keine ausgeführten oder bestandenen Godog-Tests.

Nicht als bestanden behauptet: MS-01 (direkter Codex-Abo-Modellzugang für Eino; MOD-01 ist nur Legacy-Alias), konkrete Codex-CLI-Isolation und gewählte Eino-Daten-/Toolquellen. Sie bleiben nachweispflichtige Voraussetzungen der jeweiligen Ausführungsstory. Für weitere Paperclip-Adapter, einzelne Connectoren und Experimente stehen INT-05/EXP-01 als sichtbare Zuschnittentscheidungen; keine davon wurde still aus dem Umfang entfernt.
