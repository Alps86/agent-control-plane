# language: de
@story-29 @ms-07
Funktionalität: Modellverbindungen über Neustart und Sitzungswechsel sicher verwenden
  Als Betreiber
  möchte ich bestehende Modellverbindungen wiederverwenden und bewusst erneuern oder entfernen,
  damit neue Aufrufe nur mit einem gültigen, serverseitig verwahrten Zugang beginnen.
  Die Szenarien verwenden ausschließlich kontrollierte Anbieter und künstliche Testgeheimnisse.

  Szenario: Codex-Abo-Verbindung nach echtem Serverneustart wiederverwenden
    Angenommen eine gültige Codex-Abo-Sitzung liegt im geschützten Verbindungsspeicher
    Wenn ich den lokalen Serverprozess beende und mit demselben Speicher neu starte
    Und ich die öffentliche Codex-Verbindungsansicht abrufe
    Dann zeigt sie die bestehende Verbindung ohne neuen Gerätecode als verbunden
    Wenn die serverseitige Zugangsgrenze den Zugang auflöst
    Dann erhält sie das gespeicherte Token mit der zugehörigen Konto-ID
    Und die öffentliche Antwort enthält kein Anmeldegeheimnis

  Szenario: Parallele Codex-Zugangsanforderungen erneuern dieselbe Sitzung genau einmal
    Angenommen eine erneuerbare Codex-Abo-Sitzung läuft in Kürze ab
    Und der kontrollierte Codex-Anbieter hält die Erneuerungsantwort zurück
    Wenn zwei serverseitige Zugangsanforderungen gleichzeitig eintreffen
    Und ich die Erneuerungsantwort freigebe
    Dann erhält der Anbieter genau eine Erneuerungsanfrage
    Und beide Anforderungen erhalten dasselbe neue Token mit derselben Konto-ID
    Und das neue Token ist nach einem Neustart weiterverwendbar

  Szenario: Endgültig widerrufene Codex-Sitzung verlangt erneute Anmeldung
    Angenommen eine erneuerbare Codex-Abo-Sitzung läuft in Kürze ab
    Und der kontrollierte Codex-Anbieter verweigert die Erneuerung endgültig
    Wenn die serverseitige Zugangsgrenze den Zugang auflöst
    Dann wird kein Token und keine Konto-ID ausgegeben
    Und die öffentliche Codex-Verbindungsansicht verlangt eine erneute Anmeldung
    Wenn ich den lokalen Serverprozess beende und mit demselben Speicher neu starte
    Dann bleibt der Zugang bis zu einer neuen Gerätecode-Anmeldung gesperrt

  Szenario: Widerrufenen OpenRouter-Schlüssel bewusst ersetzen
    Angenommen die zentrale OpenRouter-Verbindung enthält einen später widerrufenen Testschlüssel
    Wenn ich ihren Status über die öffentliche Settings-Grenze prüfe
    Dann wird der Schlüssel als nicht einsatzbereit mit Ersatzhinweis angezeigt
    Wenn ich über die öffentliche Settings-Grenze einen neuen Testschlüssel speichere
    Und ich den Status erneut prüfe
    Dann bleibt die Verbindungsreferenz gleich und der neue Schlüssel ist einsatzbereit
    Und der kontrollierte OpenRouter-Anbieter erhält nur den neuen Schlüssel bei der zweiten Prüfung

  Szenario: Geprüfte OpenRouter-Bindung über Rotation und Neustart behalten
    Angenommen eine geprüfte OpenRouter-Verbindung ist serverseitig gebunden
    Wenn ich ihren Schlüssel bewusst ersetze und erneut prüfe
    Dann bleibt dieselbe Bindung mit dem neuen Schlüssel auflösbar
    Wenn ich den geschützten Speicher und den Verbindungsdienst neu öffne
    Dann bleibt dieselbe Bindung mit dem neuen Schlüssel auflösbar

  Szenario: Trennen und erneutes Verbinden widerruft alte OpenRouter-Bindungen
    Angenommen eine geprüfte OpenRouter-Verbindung ist serverseitig gebunden
    Wenn ich die Verbindung trenne und mit einem neuen Schlüssel erneut einrichte und prüfe
    Dann ist die alte Bindung endgültig nicht mehr auflösbar
    Und die neue Bindung löst nur den neuen Schlüssel auf

  Szenario: Bestehende geschützte OpenRouter-Verbindung bleibt nach dem Upgrade nutzbar
    Angenommen eine vor der Bindungseinführung geschützte und geprüfte OpenRouter-Verbindung besteht
    Wenn ich den Verbindungsdienst mit demselben geschützten Speicher neu öffne
    Dann kann ich die bestehende Verbindung binden und ihren Schlüssel auflösen
    Und dieselbe Bindung bleibt nach erneutem Öffnen des Dienstes gültig

  @pending-openrouter-aufrufgrenze
  Szenario: Entfernte OpenRouter-Verbindung blockiert neue Anbieteraufrufe
    Angenommen die zentrale OpenRouter-Verbindung ist eingerichtet
    Wenn ich sie über die öffentliche Settings-Grenze trenne
    Und ich einen neuen OpenRouter-Anbieteraufruf anfordere
    Dann wird der Aufruf vor dem Anbieter wegen fehlender Verbindung abgelehnt
    Und der kontrollierte OpenRouter-Anbieter erhält nach der Trennung keine Anfrage
    Und die öffentliche Settings-Ansicht zeigt keine eingerichtete Verbindung

  Szenario: Geheimnisse bleiben an allen öffentlichen Lebenszyklusantworten verborgen
    Angenommen eine Codex-Abo-Sitzung und eine OpenRouter-Verbindung mit eindeutigen Testgeheimnissen bestehen
    Wenn ich beide öffentlichen Settings-Ansichten und JSON-Antworten abrufe
    Und ich den OpenRouter-Schlüssel ersetze und die Codex-Sitzung erneuere
    Dann enthalten keine dieser Antworten Tokens oder Schlüssel
    Und die Verbindungen zeigen nur Referenzen und nichtgeheime Statusdaten
