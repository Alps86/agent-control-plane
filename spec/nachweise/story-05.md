# Story-05 / UI-01 – unabhängiger Blackbox-Nachweis

Stand: 24. September 2026. Prüfung der statischen Vite-Vorschau und des ausgelieferten Template-/Fixture-Vertrags. Die Go-Bridge und geschützte Go-Vorschau sind Folgestories.

## Ausgeführte Prüfungen

- `npm run build` in `ui/web/` lief erfolgreich mit Vite 7.3.6. Vite meldete einen Hinweis auf `eval` im eingebundenen `htmx.org`; der Build brach nicht ab.
- Ein frischer Build erzeugte `dist/index.html`, CSS und JS, `assets/htmx.min.js`, fünf Go-HTML-Templates, 16 JSON-Fixtures und `fragments/preview-help.html`. Der Build leerte das Ausgabe-Verzeichnis vor der Erzeugung.
- Ein temporärer, außerhalb des Produkts liegender Go-Blackbox-Aufruf parste alle ausgelieferten `dist/templates/*.html` gemeinsam, decodierte jedes ausgelieferte Fixture als `map[string]any` und führte für alle 16 Fälle sowohl `page` als auch `content` mit `html/template.ExecuteTemplate` erfolgreich aus. Die Ausgaben waren jeweils nicht leer. Damit ist die öffentliche Map-Renderbarkeit geprüft, ohne eine UI-Bridge oder Go-Vorschau vorwegzunehmen.
- Die gebaute Vite-Vorschau wurde unter `http://127.0.0.1:4173/` im echten Browser geprüft. Organisation „Atelier Nord“ mit Übersicht, Projektansicht mit drei Projekten, Agentenansicht mit vier Agenten und Aufgabenansicht mit vier Aufgaben waren sichtbar; die Navigation führte zu den jeweiligen Seiten und zeigte die aktuelle Ansicht.
- Mit `fixture=empty` zeigten Projekt-, Agenten- und Aufgabenansicht jeweils den passenden Text „Noch keine …“ und einen Rückweg zur Übersicht. Mit `view=tasks&fixture=error` erschien „Aufgaben nicht verfügbar“ samt verständlichem Hinweis und Rücklink. Mit `view=organization&fixture=alternate` erschien „Küstenwerk“ statt „Atelier Nord“.
- Bei 390 × 844 Pixeln blieben alle vier Navigationsziele erreichbar; ein Klick auf Aufgaben öffnete die Aufgabenansicht. Die Dokumentbreite lag bei 375 Pixeln bei 390 Pixeln Viewport; die breite Aufgabentabelle hatte einen eigenen horizontalen Scrollbereich. Kein horizontaler Dokumentüberlauf.

## Gezielte Nachprüfung nach Korrektur

Der Hilfebutton „Über diese Vorschau“ löste beim ersten Browserlauf keinen Fragmenttausch aus. Nach der Korrektur und einem frischen Vite-Build lud ein erneuter Klick das HTMX-Fragment sichtbar in `#preview-help`: „Diese Vorschau zeigt ausschließlich lokale JSON-Beispieldaten …“.

Die nachgebesserte Navigation wurde für `fixture=empty` in Projekt-, Agenten- und Aufgabenansicht, für `fixture=error` in der Aufgabenansicht und für `fixture=alternate` in der Organisationsansicht erneut geprüft. Die passenden Zustände waren sichtbar, `aria-current="page"` markierte jeweils die aktuelle Ansicht, und der Navigationslink zur Übersicht behielt die Fixture-Auswahl. Ein tatsächlicher Klick aus dem Aufgaben-Fehlerfall öffnete die Organisationsübersicht mit `fixture=error`. Bei 390 × 844 Pixeln waren erneut alle vier Navigationslinks erreichbar; der Klick auf Aufgaben mit `fixture=empty` öffnete den richtigen Leerzustand ohne horizontalen Dokumentüberlauf.

## Abnahmegrenze

Die fünf fachlichen Gherkin-Szenarien wurden im Browser manuell über die öffentliche Vite-Oberfläche nachvollzogen. Ein ausführbarer Godog-Lauf über die Go-UI-Grenze ist in UI-01 noch nicht möglich, weil `ui/bridge` und `ui/cmd/preview` erst Story-06 und Story-07 gehören. Die erfolgreiche isolierte Go-Template-Map-Prüfung ersetzt diesen späteren öffentlichen End-to-End-Nachweis nicht.
