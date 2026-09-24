# Story-88 · EXP-01 — Abnahmenachweis

Stand: 24.09.2026. Der Nachweis umfasst die fachliche Quellenentscheidung und den öffentlichen Betreiberweg im regulären Gesamtserver.

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
2. **Sichtbarer Betreiberweg:** `GET /experimente` ist im regulären Gesamtserver verdrahtet. Vollseite und HTMX-Fragment zeigen 23 Quellzeilen in zwei Gruppen mit Reife, Nutzen, Abhängigkeiten, Quellen, Entscheidung und ACP-Implementierungsstand. Die Montage betrifft ausschließlich `app/cmd/server/main.go`.
3. **Öffentliche Prüfung:** Die [App-Godog-Spezifikation](../../features/app/experimente/story-88.feature) und die [UI-Godog-Spezifikation](../../features/ui/experimente/story-88.feature) bestanden am öffentlichen Leseweg. Der gestartete Gesamtserver mit temporärer Datenbank und `APP_ADDR=127.0.0.1:18788` lieferte HTTP 200 für Vollseite, HX-Fragment, Assets, `/health` und `/organisationen`. Echtes Chrome bestätigte bei 375 und 1280 Pixeln alle geprüften Felder und keinen horizontalen Scroll.

**Prüfstand:** Primärquellen- und Dokumentreview sowie gezielte Nachprüfung der vier Dokumentbefunde sind abgeschlossen. `node spec/experimente/export-inventory.mjs --check` bestätigte 23 Quellzeilen. `ui/web npm run build` war grün. Regulär im `app`-Modul bestand `GOCACHE=/tmp/acp-story88-gocache go test ./cucumber/experimente -v` ohne temporäres `go.work` oder Modfile nach begründeter Loopback-Freigabe: App-Godog 4 Szenarien/22 Schritte und UI-Godog mit echtem Chrome 3 Szenarien/20 Schritte, zusammen 7/42. Der readonly Serverbuild war erfolgreich. Das unabhängige Produktreview fand drei Befunde (ungebundenes UI-Feature, Assets, Nutzen der Plugins-API); alle wurden korrigiert und gezielt nachgeprüft. Die abschließende Montage in `app/cmd/server/main.go` wurde unabhängig ohne Befund reviewt; `git diff --check` und Build waren grün.

**Stand der Story:** Story-88 ist fachlich vollständig abgenommen. Root-Merge, Push und Aktualisierung des zentralen Story-Status stehen noch aus. Aus der Inventaransicht folgt keine Freischaltung der inventarisierten Experimental-Funktionen.
