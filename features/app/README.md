# Fachliche App-Features

Hier entstehen später `.feature`-Dateien für kleine Stories unter fachlichen Unterordnern, zum Beispiel `org/`, `project/` oder `issue/`. Die Namen beschreiben Fachbereiche, keine zusätzliche Ebene im Go-Modul `app/`. Unterordner und Szenarien werden erst beim Zuschnitt der jeweiligen Story angelegt.

Szenarien prüfen beobachtbares Verhalten über öffentliche HTTP-, CLI- oder andere freigegebene Anwendungsgrenzen. Sie greifen weder auf `app/internal/` noch auf private Speicher- oder Adapterdetails zu. Der Godog-Runner und die zugehörigen Step-Implementierungen liegen gemäß `spec/technik.md` in `app/cucumber/`; dieses Verzeichnis enthält nur Gherkin-Features und Dokumentation.

Für jede Story schreibt ein Spezifikationssubagent zuerst das Feature. Erst danach werden Produktverhalten und Steps umgesetzt, mit Godog geprüft und separat reviewed. Die [archivierten Akzeptanzbeispiele](../../spec/planung/akzeptanzbeispiele/README.md) sind Planungsmaterial und keine Nachweise für bestandene Tests.
