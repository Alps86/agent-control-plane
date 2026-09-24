# Ausfuehrung und routinen

Quelle: `spec/features/kern-ergaenzungen/ausfuehrung-und-routinen.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Läufe und geplante Arbeit besitzen eindeutige Herkunft und Fehlerzustände

  @RUN-01 @poc
  Szenariogrundriss: Beide Ausführungsarten erzeugen echte Ergebnisse
    Angenommen ein einsatzbereiter "<Adapter>"-Agent hat die Aufgabe "Status zusammenfassen"
    Wenn ich die Aufgabe aus der Oberfläche starte
    Dann entsteht ein tatsächlich ausgeführter Lauf mit "<Adapter>" als Ausführungsart
    Und sein erfolgreiches Ergebnis ist an der Aufgabe lesbar

    Beispiele:
      | Adapter   |
      | Eino      |
      | Codex CLI |

  @RUN-03
  Szenario: Abbruch und erneuter Versuch behalten beide Läufe
    Angenommen ein Lauf von "Mira" ist aktiv
    Wenn ich den Lauf abbreche
    Dann endet er mit dem Status "Abgebrochen"
    Wenn ich die Aufgabe erneut starte
    Dann erhält der neue Lauf eine andere Kennung
    Und der abgebrochene Lauf bleibt im Verlauf sichtbar

  @RUN-04
  Szenario: Nach Neustart keine stille Doppelaufnahme
    Angenommen ein Agentenprozess zu "Startseite prüfen" geht während des Laufs verloren
    Wenn die Anwendung neu startet
    Dann ist der alte Lauf als verlorener Prozess gekennzeichnet
    Und ein neuer Lauf entsteht nur durch eine sichtbare Fortsetzung oder einen neuen Start

  @RUN-02
  Szenario: Aufruf und Transkript verraten kein Geheimnis
    Angenommen "Mira" nutzt eine gespeicherte Modellverbindung
    Wenn ich ihren Lauf mit Aufruf, Ausgabe und Transkript öffne
    Dann sehe ich Auslöser, Zeiten, Agenten und Aufgabenbezug
    Und kein Verbindungstoken oder geheimes Umgebungsdatum wird angezeigt

  @ACT-01 @poc
  Szenario: Aktivitätsverlauf übersteht einen Neustart
    Angenommen "Status zusammenfassen" wurde angelegt und "Mira" zugewiesen
    Wenn ich die Anwendung neu starte und die Aktivität von "Nordstern" öffne
    Dann sehe ich Anlage und Zuweisung mit Zeitpunkt, Quelle und betroffenem Objekt
    Und ich kann von beiden Einträgen zur Aufgabe wechseln

  @SCH-02 @SCH-03
  Szenario: Cron-Routine erzeugt eine verknüpfte Aufgabe
    Angenommen ich plane die Routine "Tagesbericht" für "Mira" und das Projekt "Website" um 09:00 Uhr in "Europe/Berlin"
    Wenn der nächste geplante Termin fällig ist
    Dann sehe ich genau eine durch "Tagesbericht" erzeugte oder geöffnete Ausführungsaufgabe
    Und ihr Lauf zeigt Routine, Termin und auslösende Revision

  @SCH-02
  Szenariogrundriss: Ungültigen Zeitplan vor dem Speichern erklären
    Angenommen ich bearbeite die Routine "Tagesbericht"
    Wenn ich "<Eingabe>" eingebe
    Dann sehe ich einen Feldhinweis und keinen nächsten Ausführungstermin
    Und die bisherige gültige Routine bleibt unverändert

    Beispiele:
      | Eingabe                    |
      | einen ungültigen Cron-Ausdruck |
      | eine unbekannte Zeitzone  |

  @SCH-04
  Szenario: Nach Neustart gilt die gewählte Nachholregel
    Angenommen "Tagesbericht" war während eines Neustarts zweimal fällig
    Und die Regel erlaubt höchstens einen nachgeholten Termin
    Wenn die Anwendung wieder startet
    Dann entsteht höchstens eine nachgeholte Ausführungsaufgabe
    Und beide fälligen Termine haben einen sichtbaren behandelten Status

  @SCH-02
  Szenario: Zeitumstellung erzeugt keinen doppelten Routineauftrag
    Angenommen "Tagesbericht" ist für 02:30 Uhr in "Europe/Berlin" geplant
    Wenn die lokale Stunde im Herbst zweimal vorkommt
    Dann entsteht nur zum ersten Vorkommen eine Ausführungsaufgabe
    Und die Vorschau erklärt die Behandlung der doppelten Stunde

  @SCH-02
  Szenario: Ausfallende Stunde erzeugt keinen Routineauftrag
    Angenommen "Tagesbericht" ist für 02:30 Uhr in "Europe/Berlin" geplant
    Wenn 02:30 Uhr im Frühjahr ausfällt
    Dann wird der ausgefallene Termin ohne Ausführungsaufgabe protokolliert

  @SCH-05
  Szenario: Pausierte Routine kann bewusst manuell laufen
    Angenommen "Tagesbericht" ist pausiert
    Wenn ein geplanter Termin fällig wird
    Dann entsteht keine automatische Ausführungsaufgabe
    Wenn ich "Jetzt ausführen" wähle
    Dann trägt die neue Aufgabe "Manuell" als Auslöser
```
