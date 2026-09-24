# Voice auftraege

Quelle: `spec/features/modelle-sprache/voice-auftraege.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Aufgaben und Agenten über einen Eino-Sprach-Turn vorbereiten

  # VO-01: Dies ist eine fachliche Gate-Abnahme; echte Requests erfolgen erst nach der Planungsphase.
  Szenario: Sprachmodus nur für vollständig geprüftes Sprachprofil anbieten
    Angenommen "Mira" ist ein Eino-Agent mit einem OpenRouter-Chatmodell
    Und Eingabe-, Eino-Chat- und Ausgabe-Teilroute ihres Sprachprofils sind einzeln nachgewiesen
    Und die gesamte Kette einschließlich nötiger Tools ist nachgewiesen
    Wenn ich die Sprachfunktion für "Mira" öffne
    Dann sehe ich "Sprach-Turn mit gestreamter Antwort" als verfügbaren Modus
    Und ich sehe die geprüften Modelle und verwendeten Routen der drei Teilstrecken

  # VO-01, VO-02
  Szenario: Sprach-Turn über geprüfte STT-, Chat- und TTS-Teilstrecken führen
    Angenommen "Mira" besitzt ein geprüftes OpenRouter-Sprachprofil mit STT, Eino-Chat und TTS
    Und ich habe den Mikrofonzugriff erlaubt
    Wenn ich eine Frage einspreche und die Aufnahme beende
    Dann sehe ich den erkannten Text
    Und ich sehe eintreffende Textantwortteile während der Antwort
    Und ich kann die gestreamte TTS-Sprachantwort anhören
    Und die Oberfläche behauptet keine bidirektionale Echtzeitverbindung

  # VO-01, VO-02
  Szenario: Sprach-Turn über geprüfte Chat-Audioeingabe und -ausgabe führen
    Angenommen "Mira" besitzt ein geprüftes OpenRouter-Sprachprofil mit Chat-Audioeingabe und gestreamter Chat-Audioausgabe
    Und ich habe den Mikrofonzugriff erlaubt
    Wenn ich eine Frage einspreche und die Aufnahme beende
    Dann sehe ich, dass der Audioeingang an die geprüfte Chatroute übergeben wurde
    Und ich kann eintreffende Audioantwortteile anhören
    Und ich sehe das zugehörige Antworttranskript
    Und die Oberfläche behauptet keine bidirektionale Echtzeitverbindung

  # VO-01
  Szenario: Toolergebnis in der geprüften Sprachkette verwenden
    Angenommen "Mira" besitzt ein geprüftes OpenRouter-Sprachprofil mit STT, Eino-Chat und TTS
    Und "Mira" darf nur den Status des Projekts "Website" lesen
    Wenn ich "Wie ist der Status von Website?" einspreche und das Transkript sende
    Dann wird der Statuszugriff über Einos freigegebenes Tool ausgeführt
    Und das Toolergebnis fließt in die hörbare Antwort ein
    Und kein Zugriff auf ein anderes Projekt wird ausgeführt

  # VO-02
  Szenario: Verweigerte Mikrofonfreigabe behandeln
    Angenommen ich öffne den Sprach-Turn von "Mira"
    Wenn ich den Mikrofonzugriff verweigere
    Dann beginnt keine Aufnahme
    Und ich kann meine Anfrage als Text eingeben

  # VO-02, VO-03
  Szenario: Ohne prüfbares Eingabetranskript keine sprachliche Schreibaktion
    Angenommen die gewählte Chat-Audioeingabe liefert kein prüfbares Eingabetranskript
    Und keine geprüfte STT-Teilstrecke ist verfügbar
    Wenn ich per Sprache eine neue Aufgabe anfordere
    Dann kann ich den Auftrag nicht als Sprachauftrag zum Anlegen senden
    Und keine Aufgabe wird angelegt

  # VO-03
  Szenario: Eindeutigen mündlichen Auftrag nach Transkriptprüfung ausführen
    Angenommen "Mira" darf Aufgaben im Projekt "Website" anlegen
    Wenn ich "Lege im Projekt Website eine Aufgabe für die Startseite an" einspreche
    Dann sehe ich das korrigierbare Auftragstranskript vor dem Senden
    Und die Aufgabe besteht noch nicht
    Wenn ich das eindeutige Transkript als Auftrag sende
    Dann sehe ich genau eine neue Aufgabe im Projekt "Website"
    Und ich kann sie aus dem Sprachergebnis öffnen

  # VO-03
  Szenario: Unklares Projekt erfragen statt falsch anlegen
    Angenommen "Mira" darf Aufgaben in mehreren Projekten anlegen
    Wenn ich einen mündlichen Auftrag ohne eindeutiges Zielprojekt als Transkript sende
    Dann fragt "Mira" nach dem Projekt
    Und ohne meine geklärte Antwort wird keine Aufgabe angelegt

  # VO-04
  Szenario: Eindeutigen mündlichen Agentenauftrag nach Transkriptprüfung ausführen
    Angenommen ich darf in "Nordstern" Agenten anlegen
    Und die Recherche-Vorlage enthält eine zulässige Eino-Ausführungsart, Modellverbindung und Rechte
    Wenn ich "Lege in Nordstern die Recherche-Agentin Mira mit der Recherche-Vorlage an" einspreche
    Dann sehe ich das korrigierbare Auftragstranskript vor dem Senden
    Und noch kein neuer Agent besteht
    Wenn ich das eindeutige Transkript als Auftrag sende
    Dann sehe ich genau einen neuen Agenten in "Nordstern"
    Und ich sehe seine Vorlage, Ausführungsart, Modellverbindung und Rechte

  # VO-04
  Szenario: Nicht freigegebene Rechte durch Sprache nicht übernehmen
    Angenommen ich darf für "Nordstern" keine Shell-Rechte vergeben
    Wenn ich einen mündlichen Auftrag für einen Eino-Agenten mit Shell-Recht als Transkript sende
    Dann wird das fehlende Recht verständlich abgelehnt
    Und aus der Sprachanfrage entsteht kein Agent mit Shell-Recht

  # VO-05
  Szenario: Abbruch vor dem Senden erzeugt keine Fachaktion
    Angenommen ein Auftragstranskript aus einem Sprach-Turn ist noch nicht gesendet
    Wenn ich den Turn abbreche
    Dann endet die Aufnahme oder Antwort soweit vom Anbieter unterstützt
    Und keine Aufgabe wird angelegt

  # VO-05, ZA-06
  Szenario: Abbruch nach Annahme behauptet keine Rücknahme
    Angenommen ein gesendeter Sprachauftrag zur Aufgabenanlage wurde bereits angenommen
    Wenn ich den Sprach-Turn abbreche
    Dann sehe ich den tatsächlichen Auftragsstatus
    Und eine bereits angelegte Aufgabe bleibt sichtbar
    Und derselbe Auftrag wird bei Wiederaufnahme nicht erneut angelegt

  # VO-05
  Szenario: Teilstream und Limit verständlich behandeln
    Angenommen "Mira" beantwortet einen Sprach-Turn über OpenRouter
    Wenn der Antwortstream nach einem Teil wegen eines Anbieterlimits endet
    Dann sehe ich die bereits eingetroffenen Teile und den Fehlergrund
    Und ich kann den Turn bewusst erneut starten
    Und kein anderes Modell wird still verwendet
```
