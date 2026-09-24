# Story-09 / MS-02 – Codex-Abo per Gerätecode verbinden

Stand: 25. September 2026. Der [kanonische Storykatalog](../planung/modelle-sprache/stories.md) verlangt Start, Anmeldeseite und Code, Status und Abbruch des Gerätecode-Versuchs, verständliche Zustände bei deaktivierter Freigabe, Ablauf und Verweigerung sowie „verbunden“ nach Erfolg und erneute Anmeldung nach endgültigem Sitzungsfehler. Die öffentliche Settings-Grenze und geschützte Secret-Speicherung sind Voraussetzungen; der Eino-Modelltransport gehört zu Story-08 / MS-01.

## Öffentliche Abnahme

- `cd app && GOCACHE=/tmp/acp-go-cache go test ./cucumber/modellverbindung -count=1 -v` bestand mit **40/40 Godog-Szenarien und 206/206 Schritten**. Ein kontrollierter lokaler OAuth-Testanbieter prüfte Start, Polling, Abbruch, Erfolgs- und Fehlerzustände, Erneuerung und den serverseitigen `AccessResolver` mit Token und Konto-ID derselben Sitzung. Dazu gehören ein überhöhtes Anbieter-Abfrageintervall und lokaler Abbruch trotz vorübergehendem Anbieterfehler. Die Tests verwendeten keine echte Anbieteranmeldung und keinen echten Bearer-Aufruf.
- Nach simuliertem erfolgreichem Gerätecode-Ablauf wurde das Tokenbündel verschlüsselt gespeichert. Ein neu aufgebauter Settings-Handler und Dienst öffneten denselben geschützten Speicher; die öffentliche Statusroute meldete wieder „connected“, ohne einen neuen Gerätecode anzufordern. Der serverseitige Port lieferte dasselbe Token-Konto-Paar. Fehlender Masterschlüssel und zu offene Dateiberechtigungen wurden abgelehnt.
- Die öffentliche Browserabnahme `go test -tags browser ./cucumber/modellverbindung/browser -count=1 -v` bestand mit **9/9 Szenarien und 45/45 Schritten**. Nach Review der Antwortreihenfolge prüfte echter Headless Chrome, dass eine verspätete „pending“-Antwort nach „connected“ oder „cancelled“ den Gerätecode nicht wieder einblendet.

## Menschlicher Gerätecode-Nachweis

Ein Mensch schloss die Anmeldung über die ausgelieferte Oberfläche beim Anbieter ab. Die öffentliche Statusroute des gebauten Servers antwortete danach mit **HTTP 200, `connected`**. Nach Stop und Neustart des gebauten Servers mit denselben isolierten, verschlüsselten Credential-Pfaden antwortete dieselbe öffentliche Statusroute erneut mit **HTTP 200, `connected`**. Dabei wurden weder Token noch Konto-ID ausgelesen.

Die automatische Freigabeprüfung untersagte die Weitergabe des Gerätecodes an Agenten. Der Mensch führte die Eingabe selbst aus; dieses Dokument enthält weder Gerätecode noch Anmeldelink, Token oder Konto-ID.

## Grenze des Nachweises

Der reale Ablauf belegt Gerätecode-Bedienung, verbundenen Status und Status nach Neustart. Eine erzwungene Live-Token-Erneuerung wurde nicht durchgeführt. Längerfristige Wiederverwendung und koordinierte Erneuerung sind Gegenstand von Story-29 / MS-07. Ein Modellaufruf oder Eino-Toollauf über dieses Abo wurde hier nicht nachgewiesen; dafür gilt das separate technische Gate Story-08 / MS-01. Aus den Statusantworten folgt keine darüber hinausgehende Aussage zur Anbieterkompatibilität.
