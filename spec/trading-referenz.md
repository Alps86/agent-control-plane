# Lesend geprüfte Struktur des Trading-Projekts

Referenz: /home/alps86/IdeaProjects/trading (am 24. September 2026 nur lesend geprüft). Dies ist ein Strukturvorbild, keine Freigabe zur Änderung oder Übernahme fachlicher Trading-Modelle.

- app/ und ui/ sind benachbarte Go-Module (app/go.mod, ui/go.mod). app/go.mod bindet trading/ui mit replace auf ../ui lokal ein.
- app/internal/ enthält unmittelbar genau domain/, port/, app/ und adapter/. Unter adapter/ liegen im Trading-Projekt web/, config/ und sqlite/.
- app/cmd/main.go baut den Go-HTTP-Server zusammen und importiert die UI sowie die App- und Adapter-Pakete.
- ui/package.json nutzt Vite und htmx.org. ui/vite.config.js baut nach dist/ und legt Assets unter dist/assets/ ab.
- ui/public/templates/*.html enthält die Go-Templates; gebaute Fassungen liegen unter ui/dist/templates/.
- ui/ui.go verwendet go:embed für dist, template.ParseFS und bietet Render sowie Assets für die App an.
- ui/cmd/preview/main.go startet den Go-Vorschauserver. ui/preview.go rendert Templates für die Vorschau und bietet Assets aus der Vite-Ausgabe an; die Vorschau enthält eigene Beispieldaten.

Die vorhandenen Unit-Tests und konkreten Fachmodelle des Trading-Projekts werden **nicht** als Test- oder Fachvorgabe übernommen. Für Agent Control Plane gelten die separat festgelegten Godog-Szenarien und die eigenen Domänengrenzen.
