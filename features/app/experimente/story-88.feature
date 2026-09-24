# language: de
Funktionalität: Betreiber liest das Inventar der Paperclip-Experimente
  Damit offene Paritätsentscheidungen sichtbar bleiben
  liest der Betreiber das freigegebene Inventar in der öffentlichen Anwendung.

  Szenario: Das vollständige freigegebene Inventar ist als Seite lesbar
    Angenommen die Anwendung verwendet das freigegebene Inventar der Paperclip-Experimente
    Wenn der Betreiber die Inventarseite über GET /experimente öffnet
    Dann antwortet die Anwendung mit Status 200 und einer deutschsprachigen HTML-Seite
    Und die Seite zeigt den Quellenstand des Inventars
    Und sie zeigt die 14 experimentellen Bedien- und Laufzeitflächen sowie die 9 gesonderten API-Flächen jeweils genau einmal
    Und zu jedem Eintrag sind Name, sichere Primärquelle, tatsächlicher Paperclip-Reifegrad, Nutzen, ACP-Abhängigkeiten und begründete Entscheidung lesbar
    Und Plugin-Alpha, isolierte Workspaces, Cases, Chat-Style Tasks, Environments, External Objects und Status Cards sind enthalten

  Szenario: Die öffentliche Fragmentantwort enthält denselben Inventarinhalt
    Angenommen die Anwendung verwendet das freigegebene Inventar der Paperclip-Experimente
    Wenn der Betreiber die Inventaransicht über GET /experimente mit HX-Request öffnet
    Dann antwortet die Anwendung mit Status 200 und einem HTML-Fragment ohne zweite Dokumenthülle
    Und das Fragment zeigt dieselben 23 freigegebenen Einträge und Entscheidungen wie die Vollseite

  Szenario: Experimentelle und offene Entscheidungen werden nicht als fertiger ACP-Kern ausgegeben
    Angenommen die Anwendung verwendet das freigegebene Inventar der Paperclip-Experimente
    Wenn der Betreiber die Inventarseite über GET /experimente öffnet
    Dann sind Zuordnen, Später evaluieren und Abgelöst als verschiedene Entscheidungen erkennbar
    Und die später zu evaluierenden Funktionen bleiben mit Nutzen, Risiko und benötigtem Folgeentscheid sichtbar
    Und Alpha, Experiment, API-dokumentiert und zum Kern befördert bleiben als unterschiedliche Paperclip-Reifegrade erkennbar
    Und jeder Eintrag nennt den ausdrücklich nachgewiesenen ACP-Implementierungsstatus oder kennzeichnet ihn als noch nicht nachgewiesen
    Und keine zugeordnete oder experimentelle Funktion wird allein wegen der Inventarisierung als bereits in ACP verfügbar bezeichnet

  Szenario: Freigegebene Primärquellen werden sicher verlinkt
    Angenommen die Anwendung verwendet das freigegebene Inventar der Paperclip-Experimente
    Wenn der Betreiber die Inventarseite über GET /experimente öffnet
    Dann führen alle externen Quellenlinks ausschließlich zu den freigegebenen HTTPS-Primärquellen
    Und angezeigter Quellentext wird als Text statt als ausführbares HTML ausgegeben
