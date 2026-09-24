# language: de
@story-16 @agt-02 @browser
Funktionalität: Codex-CLI-Agent im Browser auswählen und seine Grenzen prüfen
  Als lokale Betreiberin
  möchte ich die getrennte Ausführungsart und ihren Bereitschaftsgrund sehen,
  damit ich ohne Entwicklerwissen eine richtige Entscheidung treffen kann.

  Szenario: Aus der Organisation einen Codex-CLI-Agenten anlegen
    Angenommen ich öffne die Anwendung für die Organisation "Nordstern" im Browser
    Wenn ich "Agent anlegen" öffne
    Dann kann ich "Codex CLI" als eigene Ausführungsart auswählen
    Wenn ich "Kai" eingebe, "Codex CLI" auswähle und speichere
    Dann erscheint "Kai" in der Agentenübersicht von "Nordstern" mit "Codex CLI"
    Wenn ich "Kai" öffne
    Dann sehe ich "Nordstern" als Organisationskontext
    Und ich sehe seine freigegebenen benannten Fachfähigkeiten oder "Keine Fachfähigkeiten freigegeben"
    Und ich sehe den konkreten Grund, warum er noch nicht startbereit ist

  Szenario: Fehlende Konfiguration bleibt nach dem Speichern sichtbar
    Angenommen ich öffne den Codex-CLI-Agenten "Kai" mit fehlender erforderlicher Konfiguration im Browser
    Dann sehe ich "Nicht startbereit"
    Und ich sehe, welche Konfiguration fehlt
    Wenn ich die Seite neu lade
    Dann sehe ich weiterhin "Nicht startbereit" mit dem fehlenden Konfigurationsgrund

  Szenario: Ohne nachgewiesene Fähigkeitsgrenze zeigt die Oberfläche keine Startbereitschaft
    Angenommen ich öffne den Codex-CLI-Agenten "Kai" im Browser
    Und die Sperre direkter Shell-, Terminal-, Prozess-, Interpreter-, HTTP- und Dateisystemwege ist für sein Profil nicht nachgewiesen
    Dann sehe ich "Nicht startbereit" und den fehlenden Durchsetzungsnachweis
    Und die Oberfläche zeigt keine erfolgreiche Ausführung oder einen abgeschlossenen Lauf
