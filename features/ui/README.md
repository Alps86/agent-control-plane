# Sichtbare UI-Features

Hier entstehen später `.feature`-Dateien für sichtbare Bedienfälle kleiner Stories unter fachlichen Unterordnern, zum Beispiel `org/`, `project/` oder `issue/`. Die Ordner folgen dem fachlichen Zuschnitt und werden erst mit der jeweiligen Story angelegt.

Szenarien beschreiben, was Menschen in der realen Oberfläche sehen und auslösen können. Sie prüfen die UI als Blackbox über öffentliche Bedienwege und behaupten kein Verhalten aufgrund von Templates, Fixtures oder internen Zuständen allein. Cucumber-Runner und Step-Implementierungen bleiben in den gemäß Projektspezifikationen vorgesehenen Teilmodulstrukturen; hier liegen nur Gherkin-Features und Dokumentation. Relevante Wege werden nach der Godog-Prüfung zusätzlich im Browser nachvollzogen, sobald die Oberfläche nutzbar ist.

Für jede Story formuliert ein Spezifikationssubagent zuerst das Feature. Implementierung und Steps folgen danach; anschließend werden Godog, Browserprüfung und separater Review durchgeführt.
