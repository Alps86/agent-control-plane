# Rechte

Quelle: `spec/features/berechtigungen/rechte.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Freigaben für Tools, Daten und Delegation verbindlich durchsetzen

  Szenario: Nicht freigegebenes Tool verweigern
    Angenommen "Mira" darf nur das Tool "Websuche" verwenden
    Wenn "Mira" über die öffentliche Ausführungsschnittstelle "Dateien schreiben" anfordert
    Dann wird der Toolaufruf verweigert
    Und ich sehe die Verweigerung im Aktivitätsverlauf

  @poc
  Szenario: Datenzugriff auf freigegebenen Bereich begrenzen
    Angenommen "Mira" darf Daten des Projekts "Website" lesen
    Wenn "Mira" über die öffentliche Datenschnittstelle Daten eines anderen Projekts anfordert
    Dann wird der Zugriff verweigert
    Und es werden keine fremden Daten zurückgegeben

  Szenario: Delegation erhält keine zusätzlichen Rechte
    Angenommen "Kai" darf nur Daten des Projekts "Website" lesen
    Und "Kai" darf an "Mira" delegieren
    Wenn "Kai" "Mira" eine Aufgabe mit Zugriff auf ein fremdes Projekt übergibt
    Dann wird dieser Zugriff auch bei "Mira" verweigert

  @poc
  Szenario: Eino-Agent hat ohne Freigabe keine Shell
    Angenommen "Mira" ist ein neu angelegter Eino-Agent
    Wenn "Mira" über die öffentliche Ausführungsschnittstelle eine Shell-Aktion anfordert
    Dann wird die Aktion verweigert
```
