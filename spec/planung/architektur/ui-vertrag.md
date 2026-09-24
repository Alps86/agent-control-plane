# UI-Struktur und Datenvertrag

Stand: 24. September 2026. Verbindlicher Plan für den späteren Umbau des unvalidierten UI-Vorlaufs. Es findet in der Planungsphase kein Produktumbau statt.

## Einfache Zielstruktur

| Pfad | Inhalt und Grenze |
| --- | --- |
| `ui/web/package.json`, `ui/web/vite.config.js`, `ui/web/index.html`, `ui/web/src/` | Vite, HTMX, JavaScript und CSS; keine Go-Dateien. |
| `ui/web/templates/*.html` | Lesbare Go-HTML-Templates mit sinnvoller Einrückung und benannten Teiltemplates; keine minifizierten Einzeiler. |
| `ui/web/fixtures/*.json` | Einzige Quelle statischer Ansichtsdaten für Vite und Go-Vorschau. Je Ansicht ein JSON-Dokument mit identischen Template-Schlüsseln. Keine fachlichen Go-App-Modelle. |
| `ui/web/dist/` | Temporäres Vite-Buildresultat; ein Buildschritt kopiert `ui/web/templates/` nach `dist/templates/` und `ui/web/fixtures/` nach `dist/fixtures/`. Keine Annahme, dass Vite diese Verzeichnisse selbst kopiert. |
| `ui/bridge/dist/` | Aus dem vollständigen `ui/web/dist/` kopierte Templates, Assets und JSON-Fixtures. `ui/bridge/embed.go` nutzt lokal `//go:embed dist`, ohne `..` oder Symlink-Trick. |
| `ui/bridge/types.go`, `ui/bridge/bridge.go` | Einzige öffentliche Go-Fassade für `app/`: Factory und Receiver-Methoden zum Rendern, Bereitstellen von Assets und Lesen einer synthetischen JSON-Fixture für die Preview. Nur `fs.FS`, `io.Writer`, `http.Handler`, Template-Name und anzeigefertige Map passieren die Rendergrenze. |
| `ui/internal/render/types.go`, `ui/internal/render/*.go` | Geschützter Parser und Renderer für eingebettete Templates. Keine fachlichen App-Typen. |
| `ui/internal/preview/types.go`, `ui/internal/preview/*.go` | Geschützter Go-Vorschauserver, der ausschließlich JSON-Fixtures lädt und dieselbe Render-Fassade verwendet. Keine in Go hardcodierten Beispieldaten. |
| `ui/cmd/preview/main.go` | Minimaler Go-Einstiegspunkt; Verdrahtung der Preview-Struct und des HTTP-Servers. |

`app/` importiert ausschließlich `agentcontrolplane/ui/bridge`; der Go-Internal-Mechanismus verbietet ihm den Import von `ui/internal/`. Die alte UI-Wurzel mit `ui/types.go` und `ui/ui.go` wird in der späteren Umbau-Story aufgelöst. Die Vorschau darf interne UI-Pakete importieren, da sie innerhalb des UI-Moduls liegt. Sie liest Fixture-JSON aus dem eingebetteten Build über die Fassade und decodiert es intern; App-Code ruft diese Preview-Methode nicht auf. Der Embed-Pfad wird durch eine explizite Buildkopie hergestellt; der Vite-Build darf Go-Dateien nicht überschreiben. Der Go-Server rendert sowohl volle Seiten als auch HTMX-Fragmente mit denselben Templates.

## Map-Vertrag als Zwischenstand

Die Fassade erhält `map[string]any` mit dem sichtbaren Schema `PageTitle`, `Navigation`, `Notice`, `Errors`, `View`. `View` enthält ansichtsspezifisch benannte Felder, etwa `Organization`, `Projects`, `Agents`, `Issue`, `Chat`, `Run`, `Settings`, jeweils als einfache Anzeige-Maps oder Listen. Das ist **kein** eigenes UI-Fachmodell: Der App-Webadapter erstellt die Map aus fachlichen Ergebnissen und entfernt nicht freigegebene Felder; die Preview decodiert dieselbe Form aus JSON. Templates lesen nur dokumentierte Schlüssel und treffen keine Rechteentscheidung. `AllowedActions` enthält nur UI-Hinweise; der Server prüft jede Aktion erneut. IDs bleiben stabile Referenzen, angezeigte Texte werden durch `html/template` escaped. Leere Listen, Validierungsfehler und fehlende Verbindungen haben eigene Fixture-Fälle.

Die öffentliche Fassade stellt Rendern eines benannten Templates und Ausliefern der eingebetteten Assets bereit. Sie kennt weder Organisationen als Go-Typ noch Repositories, Provider oder Berechtigungsports. Statische Vorschau und echte App wählen denselben Template-Namen und denselben Map-Vertrag; nur die Datenquelle unterscheidet sich. Eine Fixture darf keine Zugangsdaten, echte Nutzerdaten oder bewusst unberechtigten Inhalte enthalten.

## Abnahme des späteren UI-Umbaus

- `ui/types.go` und `ui/ui.go` liegen nicht mehr auf UI-Wurzelebene; alle Go-Implementierungs- und Preview-Typen liegen in `ui/internal/`, außer der schmalen öffentlichen `ui/bridge/`-Fassade.
- `ui/web/` enthält Vite/JS/CSS/Templates/JSON-Fixtures getrennt von Go. Ein reproduzierbarer Build erzeugt `ui/bridge/dist/` und ein frischer Go-Build findet alle Templates und Assets über lokales `go:embed`.
- Statische UI und Go-Vorschau zeigen dieselben Fixture-Fälle. Eine Änderung nur am JSON verändert die angezeigten Beispieldaten; im Go-Code steht kein Beispiel-Datensatz.
- UI-04 rendert zunächst eine gespeicherte Organisation nach ORG-01 über `ui/bridge`; Projekt-, Agenten- und Issue-Ansichten folgen erst mit ihren Fachstories. Eine fremde Organisation oder nicht erlaubte Aktion erscheint nicht durch eine Template-Abkürzung; HTTP-Endpunkte prüfen Rechte unabhängig von `AllowedActions`.
- Templates sind mit mehreren Zeilen, sinnvoller Einrückung und klaren Teiltemplates lesbar. Volle Seite und HTMX-Fragment rendern ohne doppeltes Fachmodell.
