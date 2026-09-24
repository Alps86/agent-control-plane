# Story-88 · EXP-01 — Abnahmenachweis und Restgate

Stand: 24.09.2026. Der Nachweis unterscheidet die fachliche Quellenentscheidung, den geprüften öffentlichen Leafhandler und die noch fehlende Verdrahtung im regulären Server.

## Maßstab

Der [Storykatalog](../planung/kern/storykatalog.md) nennt EXP-01 einen „Entscheid“ und fordert zugleich: „Betreiber sieht Inventar der Paperclip-Experimente.“ Für Plugin-Alpha, isolierte Workspaces, Cases, Chat-Style Tasks, Environments, External Objects, Status Cards und weitere Flags müssen Quelle, Reife, Nutzen, Abhängigkeit und Entscheidung erkennbar sein. Nichts darf als Paperclip-Ist-Kern ausgegeben oder still ausgeschlossen werden. Der [kanonische Storyindex](../planung/story-index.json) hat für Story-88 keine harten Vorgänger.

Das [Umfangsinventar](../planung/kern/umfangsinventar.md) führt diese Bereiche bewusst als „offen“, „Experiment“ beziehungsweise „Alpha“. Das [Quellenregister](../planung/kern/quellen-und-abgleich.md) trennt offizielle Ist-Guides, Alpha/Experiment und Drafts. Die anhand aktueller Primärquellen geprüfte Einzelbewertung steht in [Inventar der Experimente](../experimente/inventar.md). Der [Ansichtsvertrag](../experimente/ansichtsvertrag.md) beschreibt den öffentlichen Leseweg und die Trennung von Paperclip-Reife, ACP-Entscheidung und ACP-Implementierungsstand.

## Fachliche Zuordnung ohne stille Umfangsentscheidung

| Bereich | Bestehender fachlicher Anschluss | Grenze dieser Zuordnung |
| --- | --- | --- |
| Isolierte Workspaces | PROJ-03, PERM-04/05, RUN-01 | Belegt keine isolierte Workspace-Funktion oder deren Sicherheitswirkung. |
| Cases | TASK-01–07, ART-01/02 | Ein eigener dokumentenreicher Case-Ablauf ist damit nicht bereits umgesetzt. |
| Chat-Style Tasks | TASK-05, CHV-Stories für Chats | Paperclips experimenteller Aufgabenmodus ist nicht automatisch mit den ACP-Chats gleichzusetzen. |
| Environments | PROJ-03, PERM-04, OPS-01 | Eine eigene Environment-Verwaltung folgt daraus nicht. |
| External Objects | INT-01–05, ART-01/02 | Externe Objektbindung erfordert Einzelentscheidung zu Verbindung, Rechten und Aktualisierung. |
| Status Cards | MON-01/02, ACT-01/02 | Das vorhandene Monitoring ersetzt nicht automatisch Kartenkonfiguration oder ihre API. |
| Plugin-Alpha | INT-04/05, SKL-01–03 | Adapter- und Skillgrenzen sind kein generischer Plugin-Lader. |

Diese Anschlüsse benennen mögliche Abhängigkeiten für spätere Einzelentscheidungen. Ob ein experimenteller Ablauf zusätzlich in ACP umgesetzt wird, entscheidet die zuständige Produktkoordination anhand des Inventars; EXP-01 erteilt dafür keine vorweggenommene Architekturfreigabe. Weitere Flags und ihr jeweiliger Stand sind im Inventar einzeln zu prüfen.

## Abnahmebefund

1. **Dokumentarischer Teil:** Das [Inventar](../experimente/inventar.md) führt die im Katalog genannten Bereiche, weitere offizielle Flags und zusätzliche API-Flächen einzeln mit Primärquelle, Reife, ACP-Nutzen, Abhängigkeiten und begründeter Entscheidung. Offene Kandidaten und nicht experimentelle API-Lücken bleiben ausdrücklich sichtbar. Die Primärquellen wurden geprüft; ein unabhängiger Reviewer bewertete den dokumentarischen Teil nach Korrektur von vier Befunden und gezielter Nachprüfung als fachlich abnehmbar. Das belegt die Quellen- und Entscheidungsarbeit, keine Produktfunktion.
2. **Sichtbarer Betreiberweg:** Der öffentliche Leafhandler für `GET /experimente`, die Vollseiten-/HTMX-Templates und der Inventarexport sind umgesetzt. Der Leseweg zeigt 23 Quellzeilen in zwei Gruppen mit Reife, Nutzen, Abhängigkeiten, Quellen, Entscheidung und ACP-Implementierungsstand. Dies ist eine Leafhandler-Abnahme; die Verknüpfung mit dem zentralen Server-Mux und damit der reguläre Produktweg stehen bis zur Integration mit dem zuständigen Server-Owner noch aus.
3. **Öffentliche Prüfung:** Die [App-Godog-Spezifikation](../../features/app/experimente/story-88.feature) prüft den HTTP-Leseweg; die [UI-Godog-Spezifikation](../../features/ui/experimente/story-88.feature) prüft den sichtbaren Weg mit echtem Chrome bei 375 und 1280 Pixeln einschließlich Assets, Quellen, Feldern, Entscheidungen und Fragment. Der Leafhandler wurde über diese öffentlichen Grenzen geprüft. Der regulär gestartete Gesamtserver ist damit noch nicht abgenommen.

**Prüfstand:** Das unabhängige Review des Dokumentteils und die gezielte Nachprüfung seiner vier korrigierten Befunde sind abgeschlossen. Der Generatorcheck `node spec/experimente/export-inventory.mjs --check` bestätigte 23 Quellzeilen; `npm run build` war erfolgreich. Mit einem temporären `go.work` unter `/tmp` bestand `go test ./cucumber/experimente -v`: App-Godog 4 Szenarien/22 Schritte und UI-Godog 3 Szenarien/20 Schritte mit echtem Chrome. Das unabhängige Produktreview fand drei Befunde (ungebundenes UI-Feature, Assets, Nutzen der Plugins-API); alle wurden korrigiert und gezielt nachgeprüft. Ein regulärer Test ohne temporäres `go.work` sowie der zentrale Server-Mux-/Produktweg stehen noch aus.

**Stand der Story:** Der Dokumentteil und der öffentliche Leafhandler samt HTTP-/Godog- und Browserprüfung sind fachlich abnehmbar. Die Gesamtstory bleibt bis zur Verknüpfung mit dem regulären Server-Mux, einem Test ohne temporäres `go.work` und der Prüfung des durchgängig gestarteten Produktwegs offen. Aus der Inventaransicht folgt keine Freischaltung der inventarisierten Experimental-Funktionen.
