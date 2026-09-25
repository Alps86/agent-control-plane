# language: de
@story-21 @perm-04 @browser
Funktionalität: Codex-CLI-Arbeitsprofil im Browser verstehen und konfigurieren
  Als lokale Betreiberin
  möchte ich den privaten Arbeitsbereich und seine Grenzen auf der Agentenseite sehen,
  damit eine Profiländerung nicht mit einer erfolgreichen CLI-Ausführung verwechselt wird.

  Szenario: Codex-Agent zeigt deaktivierte sichere Vorgaben
    Angenommen ich öffne den Codex-CLI-Agenten "Kai" in "Nordstern" im Browser
    Wenn ich sein Arbeitsprofil öffne
    Dann sehe ich einen serverseitig abgeleiteten privaten Arbeitsbereich ohne Pfadeingabe
    Und die Schalter "Arbeitsbereich" und "Schreiben" sind deaktiviert
    Und der konkrete CLI-Durchsetzungsgrund ist als "Nicht startbereit" sichtbar

  Szenario: Gültige Einstellung bleibt nach erneutem Öffnen sichtbar
    Angenommen ich öffne den Codex-CLI-Agenten "Kai" in "Nordstern" im Browser
    Wenn ich sein Arbeitsprofil öffne und nur "Arbeitsbereich" aktiviere und speichere
    Dann sehe ich "Arbeitsbereich" aktiviert und "Schreiben" deaktiviert
    Wenn ich die Profilseite neu lade
    Dann sehe ich "Arbeitsbereich" aktiviert und "Schreiben" deaktiviert
    Und der CLI-Durchsetzungsgrund bleibt als "Nicht startbereit" sichtbar

  Szenario: Schreiben ohne Arbeitsbereich wird am Feld erklärt
    Angenommen ich öffne den Codex-CLI-Agenten "Kai" in "Nordstern" im Browser
    Wenn ich sein Arbeitsprofil öffne und nur "Schreiben" aktiviere und speichere
    Dann sehe ich einen Feldfehler für "Schreiben"
    Und die Schalter "Arbeitsbereich" und "Schreiben" bleiben deaktiviert

  Szenario: Fremde direkte Profiladresse verrät keinen Agenten
    Angenommen "Kai" gehört zu "Nordstern" und "Südlicht" besteht
    Wenn ich Kais Arbeitsprofil über die Browseradresse von "Südlicht" öffne
    Dann sehe ich weder Kais Profil noch seinen Arbeitsbereich
