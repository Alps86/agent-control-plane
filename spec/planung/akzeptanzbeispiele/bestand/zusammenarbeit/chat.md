# Chat

Quelle: `spec/features/zusammenarbeit/chat.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@ZA-01
Funktionalität: Mit einem Agenten oder einem Team zusammenarbeiten
  Als Betreiberin möchte ich Arbeit im Chat besprechen und den Verlauf wiederfinden.

  Szenario: Einen Agenten direkt beauftragen
    Angenommen "Mira" ist ein einsatzbereiter Agent in "Nordstern"
    Wenn ich im Einzelchat mit "Mira" die Nachricht "Prüfe die Startseite" sende
    Dann sehe ich meine Nachricht und "Mira" als Empfängerin im Chatverlauf
    Und ich sehe "Miras" Antwort beim selben Gespräch

  Szenario: Ein Team mit einem gemeinsamen Auftrag ansprechen
    Angenommen das Team "Web" in "Nordstern" besteht aus "Mira" und "Kai"
    Wenn ich im Teamchat "Web" die Nachricht "Bereitet die Freigabe vor" sende
    Dann sehe ich die Nachricht im gemeinsamen Verlauf von "Web"
    Und jede Antwort zeigt den antwortenden Agenten an
    Und die Teammitglieder können den bisherigen Gesprächsverlauf für diesen Auftrag sehen

  Szenario: Einen früheren Chat nach Neustart fortsetzen
    Angenommen ich habe mit "Mira" in "Nordstern" über "Startseite prüfen" gesprochen
    Wenn ich die Anwendung neu starte und den Einzelchat mit "Mira" öffne
    Dann sehe ich die bisherigen Nachrichten in ihrer Reihenfolge
    Und ich kann eine weitere Nachricht in diesem Gespräch senden
```
