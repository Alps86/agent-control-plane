# Voice chat schnittstelle

Quelle: `spec/features/modelle-sprache/voice-chat-schnittstelle.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@CHV-01
Funktionalität: Sprach-Turn im bestehenden Chat verwenden
  Als Betreiberin möchte ich einen Sprach-Turn im Team- oder Agentenchat führen
  und seine sichtbaren Ergebnisse im Gespräch wiederfinden.

  Szenariogrundriss: Gesprochenen Beitrag nach Neustart im Chat wiederfinden
    Angenommen ich habe einen berechtigten "<Chatart>" mit "Mira" geöffnet
    Wenn ich "Prüfe die Startseite" einspreche
    Dann sehe ich das erkannte Transkript und kann es vor dem Senden korrigieren
    Wenn ich "Prüfe nur die Texte" als korrigierten Beitrag sende
    Und ich die Anwendung neu starte und denselben Chat öffne
    Dann sehe ich "Prüfe nur die Texte" genau einmal im bisherigen Gesprächsverlauf

    Beispiele:
      | Chatart    |
      | Teamchat   |
      | Agentenchat |

  Szenario: Angelegte Aufgabe vom gesendeten Sprachauftrag aus öffnen
    Angenommen ich habe einen berechtigten Teamchat mit "Mira" geöffnet
    Und "Mira" darf im Projekt "Website" Aufgaben anlegen
    Wenn ich "Lege im Projekt Website die Aufgabe Startseite prüfen an" einspreche
    Dann sehe ich das korrigierbare Transkript vor dem Senden
    Und die Aufgabe "Startseite prüfen" besteht noch nicht
    Wenn ich den eindeutigen Auftrag sende
    Dann wird die Aufgabe ohne weitere pauschale Bestätigung angelegt
    Und ich sehe die Auftragskennung beim gesendeten Beitrag
    Und ich kann sie über einen Link beim gesendeten Auftrag im Chat öffnen
    Und von der Aufgabe führt ein Rücklink genau zur auslösenden Chatstelle
    Wenn ich nach einem Verbindungsabbruch denselben Auftrag mit derselben Auftragskennung erneut zustelle
    Und ich den Chat erneut öffne
    Dann sehe ich den Auftrag und seinen Aufgabenlink genau einmal
    Und der Aufgabenlink führt weiterhin zur ursprünglich angelegten Aufgabe
    Und im Projekt "Website" besteht genau eine Aufgabe "Startseite prüfen"

  Szenario: Fehlendes Recht beim Start eines Sprach-Turns anzeigen
    Angenommen ich darf "Mira" im Teamchat nicht ansprechen
    Wenn ich dort einen Sprach-Turn mit "Mira" starte
    Dann sehe ich die Berechtigungsverweigerung
    Und es beginnt keine Aufnahme
    Und keine Aufgabe wird angelegt

  Szenario: Nicht gesendeten Sprachauftrag abbrechen
    Angenommen ich habe im Agentenchat mit "Mira" ein Auftragstranskript vor dem Senden geöffnet
    Wenn ich den Sprach-Turn abbreche
    Dann endet die Aufnahme oder laufende Ausgabe soweit unterstützt
    Und der Auftrag erscheint nicht als gesendete Nachricht im Gespräch
    Und keine Aufgabe wird angelegt
```
