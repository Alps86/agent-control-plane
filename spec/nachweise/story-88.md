# Story-88 · EXP-01 — Abnahmenachweis und Restgate

Stand: 24.09.2026. Der Nachweis unterscheidet die fachliche Quellenentscheidung von der im Katalog verlangten Sicht des Betreibers. Er ist keine Behauptung einer implementierten Anwendung.

## Maßstab

Der [Storykatalog](../planung/kern/storykatalog.md) nennt EXP-01 einen „Entscheid“ und fordert zugleich: „Betreiber sieht Inventar der Paperclip-Experimente.“ Für Plugin-Alpha, isolierte Workspaces, Cases, Chat-Style Tasks, Environments, External Objects, Status Cards und weitere Flags müssen Quelle, Reife, Nutzen, Abhängigkeit und Entscheidung erkennbar sein. Nichts darf als Paperclip-Ist-Kern ausgegeben oder still ausgeschlossen werden. Der [kanonische Storyindex](../planung/story-index.json) hat für Story-88 keine harten Vorgänger.

Das [Umfangsinventar](../planung/kern/umfangsinventar.md) führt diese Bereiche bewusst als „offen“, „Experiment“ beziehungsweise „Alpha“. Das [Quellenregister](../planung/kern/quellen-und-abgleich.md) trennt offizielle Ist-Guides, Alpha/Experiment und Drafts. Die anhand aktueller Primärquellen geprüfte Einzelbewertung steht in [Inventar der Experimente](../experimente/inventar.md). Diese Dokumente allein sind noch keine Produktabnahme.

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
2. **Sichtbarer Betreiberweg:** Eine nur im Repository lesbare Markdown-Datei erfüllt den Satz „Betreiber sieht Inventar“ für die geplante Nichtentwickler-Oberfläche nicht. Root hat deshalb die öffentliche, lesende ACP-Inventaransicht als Restgate festgelegt. Sie ist nach dem UI-Bootstrap mit den Besitzern von `app/` und `ui/` einzubinden und muss dieselben Felder und offenen Entscheidungen anzeigen. Der Betreiber muss keine Entwicklungsdateien öffnen.
3. **Öffentliche Prüfung:** Die [fachliche Godog-Spezifikation](../../features/app/experimente/story-88.feature) beschreibt den HTTP-Leseweg: vollständige Liste, Quelle, Reife, Nutzen, Abhängigkeiten und Entscheidung je Eintrag; offene Punkte sind als offen erkennbar. Nach Einbindung der Ansicht werden dieser öffentliche Anwendungsweg und der sichtbare Bedienpfad im Browser geprüft. Aktuell sind weder Step-Implementierung noch Produkt- oder Browsertest erbracht.

**Prüfstand:** Das unabhängige Review des Dokumentteils und die gezielte Nachprüfung der vier korrigierten Befunde sind abgeschlossen. Die drei noch ungetrackten Story-Dateien wurden jeweils mit `git diff --no-index --check /dev/null <Datei>` auf Formatfehler geprüft; die Befehle lieferten keine Ausgabe. Godog-, HTTP- und Browserergebnisse liegen noch nicht vor.

**Stand der Story:** Der Dokumentteil ist fachlich abnehmbar. Die Gesamtstory bleibt bis zur öffentlichen Inventaransicht, ihrer HTTP-/Godog-Prüfung und der Browserprüfung offen. Der Integrationsbedarf beschränkt sich auf die Darstellung des vorhandenen Inventars und den öffentlichen Lesepfad; neue Experimental-Funktionen sind daraus nicht automatisch abzuleiten.
