# Rechte artefakte betrieb

Quelle: `spec/features/kern-ergaenzungen/rechte-artefakte-betrieb.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
Funktionalität: Rechte und lokale Daten bleiben an den tatsächlichen Grenzen geschützt

  @PERM-02 @poc
  Szenario: Widerruf wirkt auf den nächsten Toolaufruf
    Angenommen "Mira" hat einen aktiven Eino-Lauf und darf "Websuche" verwenden
    Wenn ich "Websuche" für "Mira" entziehe
    Und der Lauf erneut "Websuche" anfordert
    Dann wird dieser Aufruf ohne Ausführung verweigert
    Und die Verweigerung steht ohne Zugangsdaten im Aktivitätsverlauf

  @PERM-04 @poc
  Szenario: Codex CLI kann freigegebenen Arbeitsbereich nicht verlassen
    Angenommen "Kai" darf nur im Arbeitsbereich des Projekts "Website" schreiben
    Wenn "Kai" einen Schreibvorgang außerhalb dieses Bereichs anfordert
    Dann bleibt die fremde Datei unverändert
    Und der Lauf zeigt eine konkrete Grenzverletzung

  @PERM-05
  Szenario: Connector-Freigabe behauptet keine Shell-Kontrolle
    Angenommen eine externe Aktion ist im Gateway für "Kai" gesperrt
    Und "Kai" verwendet Codex CLI mit Shell-Zugriff
    Wenn ich die Ausführungsgrenzen von "Kai" öffne
    Dann sehe ich getrennt die Gateway-Regel und die Shell-/Arbeitsbereichsgrenze
    Und ein nicht erzwingbarer Shell-Bypass wird nicht als verhindert bezeichnet

  @ART-01 @ART-02
  Szenario: Fehlendes und fremdes Artefakt sicher behandeln
    Angenommen ein Arbeitsprodukt von "Nordstern" hat Metadaten, aber seine Datei fehlt
    Wenn ich es im Artefaktregal öffne
    Dann sehe ich Ursprung und einen Fehler zum fehlenden Inhalt
    Wenn ich in "Südstern" seine direkte Adresse öffne
    Dann wird weder Inhalt noch ein fremder Vorschautext ausgeliefert

  @INT-02
  Szenario: Nachfrage hält externe Aktion vor dem Aufruf an
    Angenommen "Mira" darf eine Connector-Aktion nur nach Nachfrage ausführen
    Wenn "Mira" diese Aktion anfordert
    Dann sehe ich eine offene Anfrage mit Aktion und betroffener Verbindung
    Und die externe Aktion wurde noch nicht ausgeführt
    Wenn ich die Anfrage ablehne
    Dann bleibt die externe Seite unverändert

  @INT-03
  Szenario: Neu entdeckte Connector-Aktion beginnt gesperrt
    Angenommen "Mira" darf bisherige Aktionen einer Toolverbindung verwenden
    Wenn der Connector eine neue Aktion meldet
    Dann steht die neue Aktion in der Oberfläche zunächst auf "Aus"
    Und ein Aufruf durch "Mira" wird bis zur ausdrücklichen Freigabe verweigert

  @SKL-02
  Szenario: Skill gibt keine zusätzlichen Tools frei
    Angenommen "Mira" hat den Skill "Quellen prüfen" ohne Schreibwerkzeug
    Wenn der Skill einen Schreibaufruf anstößt
    Dann wird der Schreibaufruf verweigert
    Und der Skill bleibt als Anweisung statt als Toolfreigabe erkennbar

  @OPS-01 @poc
  Szenario: Unzugängliche SQLite-Datei erzeugt keinen leeren Ersatzstand
    Angenommen die vorhandene SQLite-Datenbank ist beim Start nicht lesbar
    Wenn ich die Anwendung starte
    Dann sehe ich einen Betriebsfehler zur Datenbank
    Und keine leere neue Organisation wird als mein bisheriger Stand angezeigt

  @OPS-05 @poc
  Szenario: Öffentliche Anwendung hat weder App-Anmeldung noch Budgetdaten
    Angenommen ich öffne Startseite, Agentenformular und öffentliche Datenschnittstelle
    Dann verlangt die Anwendung keine App-Registrierung oder App-Anmeldung
    Und sie zeigt und liefert keine Budgetfelder oder Budgetlimits
```
