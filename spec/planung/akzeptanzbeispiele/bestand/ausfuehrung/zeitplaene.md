# Zeitplaene

Quelle: `spec/features/ausfuehrung/zeitplaene.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Wiederkehrende Arbeit planen

  Szenario: Cron-Zeitplan mit Zeitzone anlegen
    Angenommen "Mira" ist ein einsatzbereiter Agent
    Wenn ich für "Mira" einen täglichen Lauf um 09:00 Uhr in "Europe/Berlin" plane
    Dann sehe ich den nächsten geplanten Zeitpunkt in dieser Zeitzone
    Und der Plan ist als aktiv gekennzeichnet

  Szenario: Geplante Ausführung verfolgen
    Angenommen für "Mira" ist ein fälliger Zeitplan aktiv
    Wenn der geplante Zeitpunkt erreicht ist
    Dann sehe ich einen Lauf mit dem Auslöser "Zeitplan"
    Und ich sehe dessen Ergebnis oder Fehlergrund im Verlauf

  Szenario: Plan pausieren und manuell auslösen
    Angenommen ein Zeitplan für "Mira" ist aktiv
    Wenn ich den Zeitplan pausiere
    Dann entsteht zum nächsten geplanten Zeitpunkt kein geplanter Lauf
    Wenn ich "Mira" manuell starte
    Dann sehe ich einen Lauf mit dem Auslöser "Manuell"

  Szenario: Überschneidung beherrschen
    Angenommen ein Lauf von "Mira" aus einem Zeitplan ist noch aktiv
    Wenn derselbe Zeitplan erneut fällig wird
    Dann sehe ich keine parallele Doppelbearbeitung derselben geplanten Aufgabe
    Und die Anwendung zeigt, wie der ausgelassene oder nachgeholte Termin behandelt wurde

  Szenario: Fälligen Termin nach Neustart behandeln
    Angenommen ein aktiver Zeitplan für "Mira" war während eines Neustarts fällig
    Wenn ich die Anwendung wieder starte
    Dann sehe ich, ob der Termin nachgeholt oder ausgelassen wurde
    Und der nächste geplante Zeitpunkt wird angezeigt

  Szenario: Fehler eines geplanten Laufs sichtbar machen
    Angenommen ein geplanter Lauf von "Mira" ist fehlgeschlagen
    Wenn ich den Zeitplan öffne
    Dann sehe ich den fehlgeschlagenen Lauf mit Fehlergrund
    Und ich kann einen neuen Lauf manuell auslösen
```
