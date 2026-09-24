# language: de
@story-16 @agt-02
Funktionalität: Codex CLI als eigene Agenten-Ausführungsart anlegen
  Als lokale Betreiberin
  möchte ich Codex CLI innerhalb einer Organisation auswählen und seine Grenzen sehen,
  damit ein angelegter Agent nicht fälschlich als ausführbar gilt.

  Szenario: Codex-CLI-Agent wird im Organisationskontext gespeichert und angezeigt
    Angenommen ich starte die Anwendung mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" wurde über die öffentliche Anwendung angelegt
    Wenn ich in "Nordstern" den Agenten "Kai" mit der Ausführungsart "Codex CLI" über HTTP anlege
    Dann antwortet die Anwendung mit einer neuen Agentenkennung
    Und die Agentenübersicht von "Nordstern" enthält "Kai" genau einmal mit der Ausführungsart "Codex CLI"
    Und die Agentendetails nennen "Nordstern" als Organisation
    Und die Agentendetails zeigen die freigegebenen benannten Fachfähigkeiten oder ausdrücklich "Keine Fachfähigkeiten freigegeben"
    Wenn ich die Anwendung mit derselben SQLite-Datenbank neu starte
    Dann enthält die Agentenübersicht von "Nordstern" "Kai" genau einmal mit der Ausführungsart "Codex CLI"

  Szenario: Fehlender Durchsetzungsnachweis hält Codex CLI gesperrt
    Angenommen die Organisation "Nordstern" und ihr Codex-CLI-Agent "Kai" bestehen
    Und für das gewählte Codex-CLI-Profil ist die Sperre direkter Shell-, Terminal-, Prozess-, Interpreter-, HTTP- und Dateisystemwege nicht nachgewiesen
    Wenn ich "Kai" über die öffentliche Agenten-API abfrage
    Dann ist "Kai" nicht startbereit
    Und die Antwort nennt den fehlenden Durchsetzungsnachweis als Grund

  Szenario: Fehlende CLI-Konfiguration wird konkret benannt
    Angenommen die Organisation "Nordstern" und ihr Codex-CLI-Agent "Kai" bestehen
    Und für "Kai" fehlt eine zum Start erforderliche Codex-CLI-Konfiguration
    Wenn ich "Kai" über die öffentliche Agenten-API abfrage
    Dann ist "Kai" nicht startbereit
    Und die Antwort benennt die fehlende Konfiguration
    Und die Antwort behauptet weder einen gestarteten Lauf noch einen verbundenen Modellanbieter

  Szenario: Unzugeordneter Zugriff legt keinen Agenten an
    Angenommen ich starte die Anwendung mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" wurde über die öffentliche Anwendung angelegt
    Wenn ein HTTP-Aufruf ohne eindeutige Betreiberidentität in "Nordstern" den Codex-CLI-Agenten "Kai" anlegt
    Dann wird der Aufruf ohne Agenten-Teilerzeugung verweigert
    Und die berechtigte Agentenübersicht von "Nordstern" enthält "Kai" nicht

  Szenario: Eine Agentenkennung aus einer anderen Organisation wird nicht angezeigt
    Angenommen "Kai" ist ein Codex-CLI-Agent von "Nordstern"
    Und die Organisation "Südlicht" wurde über die öffentliche Anwendung angelegt
    Wenn ich "Kai" über die öffentliche Agenten-URL von "Südlicht" abfrage
    Dann erhalte ich dieselbe datenfreie Ablehnung wie bei einer unbekannten Agentenkennung
    Und der Name "Kai" sowie seine Fachfähigkeiten erscheinen nicht in der Antwort
