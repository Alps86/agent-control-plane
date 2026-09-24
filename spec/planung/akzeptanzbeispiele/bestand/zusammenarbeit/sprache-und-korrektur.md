# Sprache und korrektur

Quelle: `spec/features/zusammenarbeit/sprache-und-korrektur.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@ZA-05
Funktionalität: Sprach-Turns und Korrekturen im Chat nachvollziehen
  Als Betreiberin möchte ich Sprachereignisse und ihre Fachaktionen im selben Gespräch sehen.

  # CHV-01, VO-02: Der Voice-Bereich liefert die Antwort; hier zählt ihre Projektion im Chat.
  Szenario: Sprachantwort im vorhandenen Agentenchat nutzen
    Angenommen ich habe einen berechtigten Sprach-Turn mit "Mira" im Agentenchat geführt
    Und der Voice-Bereich hat eine Textfassung und eine hörbare Antwort an diesen Chatkontext zurückgegeben
    Wenn ich den Agentenchat öffne
    Dann sehe ich "Miras" Antwort als Text bei ihrem Beitrag
    Und ich sehe den vom Voice-Bereich bestätigten Turn-Zustand bei diesem Beitrag

  # CHV-01: Der Voice-Bereich liefert Kontext- und Sitzungsreferenzen; ZA-01 definiert den Teamchat.
  Szenario: Sprach-Turn dem richtigen Teamgespräch zuordnen
    Angenommen ich habe einen Sprach-Turn im Teamchat "Web" mit "Mira" geführt
    Und der Voice-Bereich hat die Ereignisse mit der Referenz von "Web" zurückgegeben
    Wenn ich den Teamchat "Web" öffne
    Dann sehe ich Transkript und Antwort im Gespräch von "Web"
    Und die Antwort ist "Mira" zugeordnet

  # VO-04, CHV-01: Agentenanlage und Berechtigungen werden im Voice-Bereich abgenommen.
  Szenario: Per Sprache angelegten Agenten im Chat öffnen
    Angenommen ein gesendeter Sprachauftrag hat "Mira" in "Nordstern" angelegt
    Und das Ergebnis wurde dem zugehörigen Chatkontext zugeordnet
    Wenn ich das Gespräch öffne
    Dann sehe ich den erfolgreichen Auftrag mit Verweis auf "Mira"
    Und ich kann den Agenten aus dem Gespräch öffnen

  # VO-05, CHV-01, ZA-06: Abbruch und Wiederaufnahme werden an den jeweiligen Grenzen abgenommen.
  Szenario: Abbruchstatus und bereits ausgeführte Aktion im Chat zeigen
    Angenommen der Voice-Bereich hat einen Sprach-Turn als abgebrochen gemeldet
    Und vor dem Abbruch wurde die Aufgabe "Texte prüfen" angelegt
    Wenn ich das zugehörige Gespräch öffne
    Dann sehe ich den Abbruchstatus und die bereits angelegte Aufgabe
    Und die Aufgabe ist als ausgeführte Aktion vom abgebrochenen Turn unterscheidbar

  # CHV-01, ZA-03, ZA-06: Der Voice-Bereich bestätigt die Korrektur; hier zählt ihre Projektion.
  Szenario: Korrektur im Chat an einen früheren Auftrag anschließen
    Angenommen "Mira" hat im Gespräch den Auftrag "Prüfe die Startseite" bearbeitet
    Und der Voice-Bereich hat die Korrektur "Prüfe nur die Texte" als neuen Auftrag zum selben Gespräch bestätigt
    Wenn ich das Gespräch öffne
    Dann sehe ich die Korrektur mit Bezug zum ursprünglichen Auftrag im Gespräch
    Und bereits ausgeführte Aktionen bleiben mit ihrem tatsächlichen Status sichtbar
```
