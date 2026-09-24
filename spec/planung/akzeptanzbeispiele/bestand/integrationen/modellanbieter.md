# Modellanbieter

Quelle: `spec/features/integrationen/modellanbieter.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Modellanbieter für Eino einfach verbinden und auswählen

  @poc
  Szenario: Codex-Abo als Standard für neue Eino-Agenten anbieten
    Angenommen ich lege einen Eino-Agenten an
    Wenn ich seinen Modellzugang öffne
    Dann ist "Codex-Abo" als Modellanbieter vorausgewählt
    Und ich kann eine Verbindung und ein verfügbares Modell wählen
    Und die Ausführungsart des Agenten bleibt "Eino"

  @poc
  Szenario: Codex-Abo mit einem Gerätecode verbinden
    Angenommen ich habe noch keine verbundene Codex-Abo-Verbindung
    Wenn ich in den Modelleinstellungen "Codex-Abo verbinden" wähle
    Dann sehe ich die Anmeldeseite des Anbieters und einen Gerätecode
    Wenn ich die Anmeldung beim Anbieter erfolgreich abschließe
    Dann zeigt die Anwendung die Verbindung als verbunden an
    Und ich kann ein über diese Verbindung verfügbares Modell auswählen

  Szenariogrundriss: Unvollständige Anmeldung verständlich behandeln
    Angenommen ich habe eine Gerätecode-Anmeldung begonnen
    Wenn die Anmeldung "<Ergebnis>"
    Dann sehe ich "<Hinweis>"
    Und die neue Verbindung wird nicht als einsatzbereit angezeigt

    Beispiele:
      | Ergebnis                       | Hinweis                                 |
      | vom Nutzer abgebrochen wird     | Anmeldung abgebrochen                   |
      | wegen abgelaufenem Code endet   | Gerätecode abgelaufen                   |
      | vom Anbieter verweigert wird    | Anmeldung vom Anbieter nicht zugelassen |

  Szenario: Deaktivierte Gerätecode-Anmeldung erklären
    Angenommen der Anbieter erlaubt für mein Konto keine Gerätecode-Anmeldung
    Wenn ich "Codex-Abo verbinden" wähle
    Dann sehe ich einen verständlichen Hinweis zur fehlenden Anbieterfreigabe
    Und die Verbindung wird nicht als einsatzbereit angezeigt

  Szenario: Anderen Modellanbieter ausdrücklich auswählen
    Angenommen "Mira" ist ein Eino-Agent mit dem Modellanbieter "Codex-Abo"
    Und eine für "Mira" freigegebene alternative Modellverbindung ist eingerichtet
    Wenn ich diese Verbindung und ihr verfügbares Modell für "Mira" speichere
    Dann sehe ich den gewählten Modellanbieter und das gewählte Modell bei "Mira"
    Und die Ausführungsart von "Mira" bleibt "Eino"

  @poc
  Szenario: Mit dem verbundenen Abo unter Eino arbeiten
    Angenommen "Mira" ist ein einsatzbereiter Eino-Agent mit verbundenem Codex-Abo
    Und "Mira" darf den Dienststatus des Projekts "Website" lesen
    Wenn ich "Mira" mit einer Zusammenfassung dieses Dienststatus beauftrage
    Dann sehe ich das Ergebnis im Lauf von "Mira"
    Und der Lauf weist Eino als Ausführungsart und Codex-Abo als Modellanbieter aus
    Und die protokollierten Datenzugriffe bleiben auf den freigegebenen Bereich begrenzt

  Szenario: Verbindung nach Neustart weiter nutzen
    Angenommen "Mira" verwendet eine gültige gespeicherte Codex-Abo-Verbindung
    Wenn ich die Anwendung neu starte
    Dann bleibt diese Verbindung "Mira" zugeordnet
    Und "Mira" kann bei weiterhin gültiger Anbieterfreigabe einen neuen Lauf beginnen

  Szenario: Erneute Anmeldung nach endgültigem Sitzungsfehler verlangen
    Angenommen die Sitzung der Codex-Abo-Verbindung wurde vom Anbieter widerrufen
    Wenn ein geplanter Lauf diese Verbindung verwenden soll
    Dann sehe ich "Erneute Anmeldung erforderlich" beim betroffenen Lauf
    Und die Anwendung wechselt nicht zu einer anderen Modellverbindung

  Szenario: Anbieterlimit ohne automatische API-Nutzung behandeln
    Angenommen "Mira" verwendet die Codex-Abo-Verbindung
    Wenn der Anbieter den Modellaufruf wegen eines erreichten Nutzungslimits ablehnt
    Dann sehe ich den Grund beim betroffenen Lauf
    Und es erfolgt kein Modellaufruf über einen separat abgerechneten API-Zugang

  Szenario: Nicht verfügbares Modell sichtbar ablehnen
    Angenommen das für "Mira" ausgewählte Modell ist über ihre Verbindung nicht mehr verfügbar
    Wenn ich einen Lauf mit "Mira" starte
    Dann sehe ich einen verständlichen Hinweis zur Modellauswahl
    Und die Anwendung verwendet nicht stillschweigend ein anderes Modell

  Szenario: Anbieterzugang bleibt außerhalb von Agentenkontext und Laufansicht
    Angenommen "Mira" verwendet eine verbundene Codex-Abo-Verbindung
    Wenn ich ihre Konfiguration und ihren abgeschlossenen Lauf über die öffentliche Anwendung öffne
    Dann sehe ich den Verbindungsstatus und den verwendeten Modellanbieter
    Und keine Zugangstokens werden angezeigt
    Und die an den Modellanbieter gesendeten Agentennachrichten enthalten keine Zugangstokens

  Szenario: Verbindung trennen
    Angenommen "Mira" verwendet eine verbundene Codex-Abo-Verbindung
    Wenn ich diese Verbindung in der Anwendung trenne
    Dann wird sie als getrennt angezeigt
    Und neue Läufe mit dieser Verbindung verlangen eine erneute Anmeldung
```
