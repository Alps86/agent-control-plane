# language: de
Funktionalität: Betreiber liest das Inventar der Paperclip-Experimente
  Damit offene Paritätsentscheidungen sichtbar bleiben
  möchte der Betreiber das geprüfte Inventar über die öffentliche Anwendung lesen.

  Szenario: Alle bewerteten Experimente sind über HTTP lesbar
    Angenommen die Anwendung verwendet das freigegebene Inventar der Paperclip-Experimente
    Wenn der Betreiber das Inventar über den öffentlichen HTTP-Leseweg öffnet
    Dann enthält die Antwort jeden Eintrag des freigegebenen Inventars genau einmal
    Und sie enthält Plugin-Alpha, isolierte Workspaces, Cases, Chat-Style Tasks, Environments, External Objects und Status Cards
    Und zu jedem Eintrag sind Quelle, Reifegrad, Nutzen, Abhängigkeiten und begründete Entscheidung lesbar

  Szenario: Offene Experimente bleiben als offen kenntlich
    Angenommen die Anwendung verwendet das freigegebene Inventar der Paperclip-Experimente
    Wenn der Betreiber das Inventar über den öffentlichen HTTP-Leseweg öffnet
    Dann sind noch offene Entscheidungen ausdrücklich als offen gekennzeichnet
    Und kein als Alpha, Experiment oder Entwurf bewerteter Eintrag wird als fertiger Ist-Kern bezeichnet
    Und jeder Eintrag zeigt seinen tatsächlich nachgewiesenen ACP-Implementierungsstatus
    Und ein noch nicht implementierter Eintrag wird nicht als verfügbare ACP-Funktion bezeichnet
