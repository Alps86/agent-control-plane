# Lokale Go-Vorschau der Oberfläche

Die Go-Vorschau zeigt die gebauten UI-Templates mit synthetischen JSON-Fixtures. Sie ist ein Entwicklungswerkzeug. Der echte Anwendungsserver bezieht seine Ansichtsdaten aus der Anwendung und verwendet Fixtures nicht als Ersatz für fehlende Daten.

## Bauen und starten

Voraussetzung sind Node.js mit npm und Go. Im Projektverzeichnis:

```sh
cd ui/web
npm ci
npm run build
cd ..
ACP_PREVIEW_PORT=4174 go run ./cmd/preview
```

`npm run build` erzeugt `ui/web/dist/` und kopiert den vollständigen Build nach `ui/bridge/dist/`. Erst danach kann Go die Templates, Assets und Fixtures über die Bridge einbetten. Bei einer Änderung an einer JSON-Fixture, einem Template oder einem Asset den Build wiederholen und den Go-Vorschauprozess neu starten. Die Vorschau lädt den eingebetteten Stand; sie bietet keinen Dateiwächter.

`ACP_PREVIEW_PORT` wählt den TCP-Port der Vorschau. Beispiel: `4174`. Der Prozess bindet ausschließlich an die Loopback-Adresse `127.0.0.1`; er darf nicht an `0.0.0.0` oder eine Netzwerkschnittstelle gebunden werden. Ein ungültiger oder bereits belegter Port führt zu einem Startfehler. Im Browser `http://127.0.0.1:4174/?view=organization` öffnen. Die Vorschau benötigt keinen Anwendungslogin. Loopback begrenzt den Zugriff auf den lokalen Rechner; andere lokale Prozesse beziehungsweise Nutzer mit Zugriff auf diesen Rechner können die synthetischen Daten ebenfalls sehen.

## Ansichten und Fixtures

| Adresse hinter `http://127.0.0.1:4174` | Anzeige |
| --- | --- |
| `/?view=organization` | Organisationsübersicht aus `organization.json` |
| `/?view=projects` | Projekte aus `projects.json` |
| `/?view=agents` | Agenten aus `agents.json` |
| `/?view=tasks` | Aufgaben aus `tasks.json` |
| `/?view=organization&fixture=alternate` | Zweite Organisation aus `organization-alternate.json` |
| `/?view=projects&fixture=empty` | Leere Projektansicht aus `projects-empty.json` |
| `/?view=tasks&fixture=error` | Fehlerzustand aus `tasks-error.json` |

Für jede der vier Ansichten gibt es die Varianten `alternate`, `empty` und `error`. Die Navigation behält die gewählte Variante bei. Normale GET-Anfragen liefern eine vollständige Seite. Eine GET-Anfrage mit `HX-Request: true` erhält für dieselbe Ansicht das austauschbare Inhaltsfragment. CSS und HTMX liegen unter `/assets/`, die Vorschauhilfe unter `/fragments/preview-help.html`. Unbekannte Ansichten, Varianten und andere Pfade werden nicht auf eine andere Fixture umgeleitet. Fixture-Dateien sind keine öffentliche Daten-API der Go-Vorschau.

Die Beispieldaten liegen ausschließlich in `ui/web/fixtures/*.json`. In diese Dateien gehören keine echten Nutzerdaten, Secrets oder fremden Organisationsdaten. Der Vorschauprozess darf diese Fixtures nicht in die Routen des echten Anwendungsservers einspeisen.
