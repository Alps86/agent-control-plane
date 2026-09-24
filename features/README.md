# Künftige ausführbare Features

Diese Verzeichnisstruktur ist die Zielablage für später entstehende Gherkin-Features. Hier liegen derzeit bewusst keine `.feature`-Dateien. Die 31 früheren Quellen aus `spec/features/` mit 177 Szenarien sind als Markdown im [Archiv der Akzeptanzbeispiele](../spec/planung/akzeptanzbeispiele/README.md) unter `spec/planung/akzeptanzbeispiele/bestand/` gesichert; die ursprünglichen `.feature`-Dateien wurden nach verifiziertem TAR entfernt. Diese README-Dateien belegen weder eine Implementierung noch bestandene Tests.

- [`app/`](app/README.md): fachliche Abnahme über öffentliche Anwendungsgrenzen.
- [`ui/`](ui/README.md): sichtbare Bedienfälle über die reale Oberfläche.

Neue ausführbare `.feature`-Dateien werden ausschließlich unter `features/app/<fachbereich>/` oder `features/ui/<fachbereich>/` abgelegt. Fachliche Unterordner wie `org/`, `project/` und `issue/` entstehen erst mit der jeweils kleinen Story; sie sind keine vorab angelegten Pakete oder Lieferzusagen. Produktcode und Cucumber-Step-Implementierungen bleiben in den dafür vorgesehenen App- und UI-Teilmodulen gemäß den Projektspezifikationen, nicht in diesem Verzeichnis.

## Arbeitsfolge je Story

1. Ein Spezifikationssubagent formuliert für **jede hinreichend kleine Story zuerst** die passende `.feature`-Datei mit beobachtbaren Akzeptanzbeispielen am zutreffenden Pfad.
2. Danach folgen Implementierung und zugehörige Step-Implementierungen in den Teilmodulen.
3. Godog prüft die Szenarien als Blackbox über öffentliche Grenzen; relevante UI-Wege werden zusätzlich im Browser geprüft, sobald die Oberfläche nutzbar ist.
4. Ein separater Review prüft die konkrete Story und ihre Nachweise; Befunde werden behoben und gezielt erneut geprüft.

Diese Folge beschreibt spätere Umsetzungsarbeit nach der vorgesehenen Planungsfreigabe.
