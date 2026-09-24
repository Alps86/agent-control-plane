# Modellzugang und Codex-Abonnement

Stand: 24. September 2026. Planungsstand; kein getesteter Integrationsnachweis.

## Festgelegtes Produktverhalten

Eino-Agenten sollen über einen eigenen Chatmodell-Adapter standardmäßig den Modellzugang des ChatGPT-/Codex-Abonnements nutzen. Die Bedienung orientiert sich am Gerätecode-Login von Hermes Agent. Ausführungsart, Modellanbieter, Verbindung und Modell sind getrennt konfigurierbar. Für einen Eino-Agenten bleibt die Ausführungsart Eino, auch wenn sein Modellzugang über Codex erfolgt.

Der Einrichtungsweg für das Codex-Abo ist die eigene Bedienstory **MS-02**: Unter Modellanbieter „Codex-Abo verbinden“ wählen, den Verbindungsvorgang starten, Anmeldeseite und Gerätecode sehen, die Anmeldung beim Anbieter bestätigen, Status prüfen und anschließend eine verfügbare Verbindung samt Modell in **MS-05** auswählen. Die UI zeigt auch deaktivierte Gerätecode-Freigabe, Ablauf, Abbruch, verweigerte Anmeldung und nötige erneute Anmeldung verständlich an. Ein erfolgreiches Verbinden beweist noch keinen Eino-Modelltransport; dafür gilt das getrennte Gate **MS-01**. Das externe Konto wird mit dem Anbieter verbunden; Agent Control Plane erhält dadurch keine eigene Benutzeranmeldung.

Eine gültige Verbindung ist wiederverwendbar für die ihr zugeordneten Agenten. Für OpenRouter gibt es genau **einen zentralen Key pro Installation** (`MS-03`); seine Einrichtung erteilt noch kein Nutzungsrecht. `MS-04` verlangt ausdrückliche Freigaben sowohl für die Organisation als auch für den ausführenden Agenten, geprüft bei Auswahl und unmittelbar vor dem Aufruf, auch bei delegierter Arbeit. Tokens erscheinen nicht in Agentenprompts, Kommentaren, Laufprotokollen oder Konfigurationsantworten. Eine gewöhnliche Konfigurationsansicht zeigt Verbindungsstatus und Referenzen, keine Anmeldegeheimnisse. Sitzungen werden serverseitig erneuert, soweit der Anbieter dies erlaubt (`MS-07`). Ein endgültig ungültiger Zugang erfordert eine erneute Anmeldung; parallele Läufe dürfen keine konkurrierende Erneuerung derselben Sitzung auslösen.

Ein anderer eingerichteter Anbieter oder ein anderes verfügbares Modell kann ausdrücklich ausgewählt werden. Fehler oder erreichte Anbieterlimits verursachen keinen stillen Wechsel zu einer separat abgerechneten API. Anbieterlimits müssen als Ausführungszustand behandelt werden; daraus entsteht keine eigene Budgetverwaltung.

Settings verwaltet Codex-Abo mit Gerätecode und den zentralen OpenRouter-Key als konkrete Zugangsarten. Die angebotene Authentifizierungsart, Verbindungsreferenz, Modellliste und geprüften Fähigkeiten ergeben sich aus der jeweiligen Anbieterkonfiguration. Beim Start eines Eino-Laufs löst eine kleine mapbasierte Registry die gewählte Anbieterkennung zu genau einem registrierten Chatmodell-Adapter auf. Agentenablauf, Anwendungsfälle und Rechteprüfung verwenden dafür anbieterneutrale Werte; anbieterspezifische Anmeldung, Transportdetails und Fehlerübersetzung liegen in den Adaptern. Unbekannte Anbieter oder fehlende Konfiguration blockieren den Lauf mit verständlichem Grund und lösen keinen Ersatzaufruf aus. Gemini und Anthropic bleiben mögliche spätere Erweiterungen ohne V1-, OAuth- oder Abo-Zusage. Eine öffentliche Blackbox-Abnahme mit einem registrierten Testanbieter ist als MS-09 geplant; sie verlangt keinen vorgezogenen allgemeinen Plugin-Lader.

## Geprüfte Quellen und Grenzen ihrer Aussage

| Quelle | Belegter Sachverhalt | Noch nicht dadurch belegt |
| --- | --- | --- |
| [OpenAI: Authentication](https://learn.chatgpt.com/docs/auth) | Codex unterstützt ChatGPT-Anmeldung für Abo-Zugang und API-Schlüssel für separat abgerechnete API-Nutzung. Gerätecode-Anmeldung ist dokumentiert, als Beta bezeichnet und muss gegebenenfalls in Konto- oder Workspace-Einstellungen aktiviert sein. | Allgemeine Freigabe eines eigenständigen Eino-Clients für direkten Zugriff auf den Codex-Modelltransport. |
| [OpenAI: Codex App Server](https://learn.chatgpt.com/docs/app-server#auth-endpoints) | Einbettung von Codex mit Anmeldeablauf, Gerätecode, Status, Abmeldung und automatisch verwalteten Tokens ist dokumentiert. Extern verwaltete ChatGPT-Tokens sind als experimenteller Modus beschrieben. | Dass ein App-Server-Agentenlauf die gewünschte reine Chatmodell-Schnittstelle unter Eino ersetzt. |
| [Hermes: AI Providers](https://github.com/NousResearch/hermes-agent/blob/main/website/docs/integrations/providers.md) | Hermes dokumentiert einen Codex-Anbieter mit Gerätecode-Anmeldung, eigenem Zugangsspeicher und Erneuerung ohne erforderliche Codex-CLI-Installation. | Hermes dokumentiert die unterstützten ChatGPT-Abostufen und die genaue Anrechnung auf deren Kontingente ausdrücklich nicht vollständig. Das Verhalten unseres Adapters ist damit ebenfalls noch nicht geprüft. |

Eine erfolgreiche Gerätecode-Anmeldung und eine erfolgreiche Modellanfrage sind getrennte Nachweise. Die Recherche bestätigt ein vorhandenes Vorbild; sie bestätigt noch keine funktionierende Eino-Integration oder allgemeine offizielle Unterstützung ihres direkten Transports.

## Kanonische Stories und frühere IDs

Die kanonischen Story-IDs und Abnahmen stehen ausschließlich in [Modellwahl und Sprache](planung/modelle-sprache/stories.md): **MS-01 bis MS-09**. `MOD-01` bis `MOD-05` sind ausschließlich Legacy-Aliasse aus dem ersten Planungsentwurf; sie erzeugen weder eigene Branches noch doppelte Abnahmen.

| Früherer Verweis | Kanonische Story | Inhalt und tatsächliche Voraussetzung |
| --- | --- | --- |
| MOD-01 | **MS-01** | Echter Abo-Modelltransport im eigenen Eino-Chatmodell-Adapter; frühes technisches Gate nach der Planung. |
| MOD-02 | **MS-02** | Gerätecode-Bedienung mit Status, Ablauf, Abbruch und erneuter Anmeldung. Die spätere dauerhafte Wiederverwendung gehört zu MS-07. |
| MOD-03 | **MS-05** | Anbieter, freigegebene Verbindung und konkretes Modell je Eino-Agent wählen. |
| MOD-04 | **MS-07** | Sichere Persistenz, Erneuerung und Trennen; redigierte Anzeige und Rechte werden zusätzlich in MS-08 geprüft. |
| MOD-05 | **MS-06** | Gewählte Route und Fehler ohne stillen Wechsel; die sichtbare Laufdiagnose und der Geheimnisschutz werden zusätzlich in MS-08 geprüft. |

**MS-03** (OpenRouter-Key), **MS-04** (Freigaben), **MS-08** (Diagnose und Geheimnisschutz) und **MS-09** (registrierter Testanbieter über Konfiguration) sind zusätzliche eigenständige Stories. Fachliche Szenarien: [Modellanbieter verbinden](planung/akzeptanzbeispiele/bestand/integrationen/modellanbieter.md), [Modellwahl](planung/akzeptanzbeispiele/bestand/modelle-sprache/modellwahl.md) und [MS-09-Akzeptanzbeispiele](planung/modelle-sprache/anbieter-erweiterbarkeit-beispiele.md). Sie sind geplante Blackbox-Abnahmen. Eine echte Abo-Anbindung wird erst in MS-01 nach dem Planungsabschluss geprüft; bei Scheitern bleibt der Eino-POC sichtbar blockiert, ohne automatischen Wechsel auf eine separat abgerechnete API.

Die anfänglichen Eino-Tools, Codex-CLI-Arbeitsgrenzen und Secret-Speicherung sind in [Planungsentscheidungen](planung/entscheidungen.md), [Kernkatalog](planung/kern/storykatalog.md) (`PERM-01/02/04/05`, `OPS-02`) und [Modellkatalog](planung/modelle-sprache/stories.md) (`MS-04`, `MS-07/08`) verankert.
