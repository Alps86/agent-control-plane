# language: de
@story-88 @exp-01 @browser
Funktionalität: Betreiber liest die Entscheidungen zum Paperclip-Inventar
  Die öffentliche Inventaransicht zeigt den freigegebenen Quellenstand vollständig
  und trennt Paperclip-Reife, ACP-Entscheidung und tatsächlichen ACP-Stand.

  Grundlage:
    Angenommen die Anwendung verwendet das freigegebene Paperclip-Inventar

  Szenario: Vollständiges Inventar auf einem schmalen Bildschirm lesen
    Angenommen ich betrachte die Anwendung auf einem schmalen Bildschirm
    Wenn ich die öffentliche Seite /experimente öffne
    Dann sehe ich den Quellenstand vom 24.09.2026
    Und ich sehe zwei getrennte Gruppen mit 14 experimentellen und 9 weiteren API-Quellzeilen
    Und ich kann alle 23 Einträge ohne horizontales Scrollen lesen
    Und jede Quellzeile nennt Reifegrad, Nutzen, ACP-Abhängigkeiten, Entscheidung und ACP-Implementierungsstand
    Und Plugin-Alpha, isolierte Workspaces, Cases, Chat-Style Tasks, Environments, External Objects und Status Cards sind lesbar

  Szenario: Offene Entscheidungen sind verständlich von implementierten Funktionen getrennt
    Wenn ich die öffentliche Seite /experimente öffne
    Dann erklärt die Seite Zuordnen, Später evaluieren, Abgelöst und Verweis getrennt
    Und ein später zu evaluierender Eintrag nennt Nutzen, Risiko und erforderlichen Folgeentscheid
    Und eine Story-Zuordnung wird nicht als implementierte ACP-Funktion dargestellt
    Und die zweite Quellzeile zu Status Cards verweist sichtbar auf die zugehörige Entscheidung

  Szenario: Quellen und Fragment bleiben nutzbar
    Wenn ich die öffentliche Seite /experimente öffne
    Dann führen die Quellenlinks zu den freigegebenen HTTPS-Seiten unter docs.paperclip.ing
    Und die Seite hat eine Hauptüberschrift und erreichbare Gruppenüberschriften
    Wenn ich denselben Weg als HTMX-Fragment öffne
    Dann sehe ich dieselben 23 Quellzeilen ohne zweite Dokumenthülle
