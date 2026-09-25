# language: de
@story-24 @ms-04 @ui @browser
Funktionalität: OpenRouter-Freigaben in Settings verwalten
  Als Betreiber
  möchte ich je Organisation und Eino-Agent die Nutzung der zentralen Verbindung erkennen und ändern,
  damit eine Schlüsselverbindung keine stillen Nutzungsrechte erzeugt.

  Grundlage:
    Angenommen eine zentrale OpenRouter-Verbindung mit synthetischem Testschlüssel eingerichtet ist
    Und die Organisationen "Nord" und "Süd" sowie die Eino-Agenten "Mira" und "Nora" in "Nord" bestehen

  Szenario: Organisation und Agent bewusst nacheinander freigeben
    Wenn ich "Settings > Modellanbieter > OpenRouter" im echten Browser öffne
    Dann sehe ich die zentrale Verbindungsreferenz ohne Schlüssel
    Und "Nord" und "Süd" werden als nicht freigegeben angezeigt
    Wenn ich "Nord" ausdrücklich freigebe
    Dann sehe ich "Nord" als freigegeben und "Mira" und "Nora" weiterhin als nicht freigegeben
    Wenn ich "Mira" ausdrücklich freigebe
    Dann sehe ich "Mira" als freigegeben und "Nora" weiterhin als nicht freigegeben

  Szenario: Freigabe widerrufen und fehlende Ebene erklären
    Angenommen "Nord", "Mira" und "Nora" sind ausdrücklich freigegeben
    Wenn ich in den OpenRouter-Einstellungen die Freigabe für "Mira" widerrufe
    Dann sehe ich "Mira" als nicht freigegeben und "Nora" weiterhin als freigegeben
    Wenn ich die Freigabe für "Nord" widerrufe
    Dann zeigt die Oberfläche die fehlende Organisationsfreigabe für beide Agenten
    Und eine frühere Agentenfreigabe erscheint nicht als nutzbare Verbindung

  Szenario: Öffentliche Bedienung verrät keinen Schlüssel
    Angenommen "Nord" und "Mira" sind ausdrücklich freigegeben
    Wenn ich die OpenRouter-Einstellungen und die Agentenansicht für "Mira" im echten Browser öffne
    Dann enthalten sichtbare Seite und Browserantworten nur Verbindungsreferenz und Freigabestatus
    Und weder der synthetische Testschlüssel noch ein Teil davon erscheint in HTML, JSON oder Formularfeldern
