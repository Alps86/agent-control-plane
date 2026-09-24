# Technische Festlegungen

## Struktur und Architektur

- Go/Golang mit hexagonaler Architektur und eigenem Go-Server; kein Node-Server.
- Die künftigen Wurzelverzeichnisse `app/` und `ui/` liegen nebeneinander; beide Produktbäume sind derzeit aus dem Arbeitsbaum entfernt.
- app/internal/ hat **genau vier unmittelbare Verzeichnisse**: domain/, adapter/, port/, app/. Keine weiteren Geschwister unter internal/.
- domain/ enthält Fachregeln, port/ die benötigten Schnittstellen, app/ Anwendungsfälle und adapter/ externe Grenzen und Implementierungen. Eino, Codex CLI, SQLite-Speicherung, Scheduler und Web sind Adapter.
- Modul- und Fachgrenzen bleiben erhalten. Keine UI-eigenen App-Datenmodelle und kein spontanes Shared-Type-Paket.
- Fachlich sinnvolle Subpackages werden unter diesen vier Verzeichnissen gebildet; die unmittelbare Ebene von app/internal/ bleibt unverändert.

## Go-Typen und Verhalten

- Objektorientiert mit Go-Structs und Methoden mit Receiver entwickeln. Verhalten wird grundsätzlich als Receiver-Methode umgesetzt.
- Freie Funktionen sind ausschließlich **exportierte Factories** zur Erzeugung von Structs beziehungsweise zur Konstruktion über Paketgrenzen hinweg. Keine freien Hilfsfunktionen oder weiteren zustandslosen Paketfunktionen als Ausnahme einführen.
- Alle Go-Typdeklarationen eines Packages gehören in dessen Datei types.go, auch in fachlichen Subpackages. Verhaltensmethoden können in passend benannten anderen Dateien liegen.
- **Enge technisch notwendige Ausnahme:** `func main()` in `app/cmd/` und `ui/cmd/preview/` sowie der von `go test`/Godog benötigte `func TestFeatures(*testing.T)` in `app/cucumber/` sind freie Einstiegspunkte, weil Go diese Signaturen verlangt. Sie verdrahten nur Receiver-basierte Objekte. Alle weiteren freien Funktionen bleiben ausgeschlossen; insbesondere gibt es keine freien Helfer in Produkt- oder Step-Paketen.
- Dependency Injection sowie Facade-, Factory- und besonders Observer-Pattern dort einsetzen, wo sie fachlich passen und die Teile entkoppeln. Clean Code und YAGNI gelten dabei; keine künstliche Komplexität aus einer pauschalen Patternpflicht ableiten.

## Code-Stil

- Clean Code und YAGNI.
- Keine else-Zweige, keine switch-Anweisungen.
- Kurze Funktionen mit maximal 20 Zeilen.
- Nach einer schließenden geschweiften Klammer eines Blocks, etwa if oder for, folgt eine Leerzeile.

## Fachliche Tests

- Von Anfang an Go-Cucumber mit **Godog** und Specification by Example, ausschließlich als **Blackbox-Tests über öffentliche Anwendungsgrenzen**.
- Keine Whitebox- oder Unit-Tests; keine Tests neben Produktquellcode.
- Die spätere Struktur von `app/` umfasst `cmd/`, `internal/` und `cucumber/`. Godog-Runner und Step-Implementierungen gehören nach `app/cucumber/`. Ein Spezifikationssubagent schreibt für jede Story zuerst die ausführbare `.feature`-Datei unter `features/app/<fachbereich>/` beziehungsweise für sichtbare Bedienfälle unter `features/ui/<fachbereich>/`. Erst danach folgen Implementierung, Steps, Godog-Prüfung und separates Review. Die früheren Dateien unter `spec/features/` wurden als [Markdown-Planungsbeispiele](planung/akzeptanzbeispiele/README.md) gesichert; sie sind keine bestandenen Tests.
- Lesbare fachliche Szenarien sind zugleich ausführbare Spezifikation und Dokumentation.
- Die Testabdeckung der gesamten Anwendung soll durch sinnvolle Cucumber-Szenarien möglichst in Richtung 100 Prozent entwickelt werden. Es gibt unvollständigen, unvalidierten Implementierungsvorlauf und keine aktuelle Abdeckungszahl; implementierungsspiegelnde Fülltests sind nicht gemeint.
- Sobald ein nutzbarer Stand verfügbar ist, werden zusätzlich echte Oberflächen und relevante Bedienpfade im Browser geprüft, soweit ein Browser verfügbar ist. Reviewer prüfen konkrete Risiken; Fixes werden gezielt nachgeprüft.

## Beobachtbarkeit und Token-Nutzung

- Eino und ein zugelassenes Codex-CLI-Profil schreiben denselben strukturierten Ereignisvertrag: Level, Ereignis, Zeit, Organisation, Projekt, Aufgabe, Agent, Sitzung, Lauf, Versuch, tatsächlicher Provider und tatsächliches Modell. Ein regelmäßiger Heartbeat berichtet nur die letzte belegte Phase mit Zeitstempel; Prozentwerte oder Modellfortschritt ohne Beleg werden nicht erzeugt.
- SQLite hält Token-Nutzung je Aufgabe, Lauf, Versuch, Provider und Modell persistent. Input-, Output- und Reasoning-Zahlen werden mit der rohen Zählsemantik des Providers übernommen, soweit geliefert. Unbekannt, unvollständig und geschätzt sind eigene Zustände, nicht der Zahlenwert null. Wenn Reasoning bereits Teil von Output ist, wird es nicht nochmals addiert. Ereignis- und Nutzungs-IDs verhindern Doppelzählung bei Streaming, Retry, Wiederaufnahme und Delegation; gemeldete Nutzung bleibt auch nach Abbruch erhalten.
- Logs, Nutzungssätze und Heartbeats enthalten keine Secrets, vollständigen Prompts oder Chain-of-Thought-Inhalte. Die Nutzungsdaten begründen kein Budget-, Abonnement- oder Kostenmodell.

## Oberfläche

- HTMX im Frontend. `ui/web/` enthält Vite, JavaScript, CSS, HTML-Templates und erzeugte Assets; Go-Implementierung liegt räumlich getrennt unter `ui/internal/`.
- `ui/cmd/preview/` ist nur der Einstiegspunkt des kleinen Go-Vorschauserver; Vorschauablauf und Template-Rendering liegen getrennt unter `ui/internal/preview/` und `ui/internal/render/`. Die einzigen Vorschau-Daten liegen als JSON-Fixtures unter `ui/web/fixtures/`.
- `ui/bridge/` ist die **einzige öffentliche Go-Integrationsgrenze** für `app/`. Nach dem Vite-Build werden `ui/web/templates/` und `ui/web/fixtures/` ausdrücklich nach `ui/web/dist/` kopiert; danach wird das vollständige `ui/web/dist/` nach `ui/bridge/dist/` kopiert. Dort bettet `go:embed dist` es ohne verbotenen `..`-Pfad ein. `ui/internal/render/` verarbeitet die übergebene `fs.FS` und rendert Go-Templates. `app/` importiert ausschließlich `agentcontrolplane/ui/bridge`, niemals `ui/internal/`.
- Die statische UI und der Vorschauserver lesen ihre Ansichtsdaten **ausschließlich aus JSON-Fixtures** unter `ui/web/fixtures/`. Es gibt keine in Go hardcodierten Beispieldaten und keine fachlichen App-Modelle im UI-Modul. Die echte App übergibt echte, bereits gefilterte Ansichtsdaten über die öffentliche UI-Fassade.
- Go-HTML-Templates sind sauber strukturiert, vernünftig eingerückt und lesbar. Frühere minifizierte Einzeiler aus dem entfernten UI-Vorlauf sind kein Stilvorbild.
- Das bestehende Trading-Projekt dient als **nur lesend** betrachtetes Strukturvorbild; Belege stehen in spec/trading-referenz.md.

Aktuell wird zuerst die vollständige Feature- und Story-Planung mit unabhängigem Planungsreview abgeschlossen. Produktimplementierung und Implementierungsreviews beginnen erst danach. Die früheren Produktbäume `app/` und `ui/` sind gesichert und aus dem Arbeitsbaum entfernt; ihr unvalidierter Vorlauf legt die Architektur nicht zusätzlich fest. Git wurde nach dem Zurücksetzen durch den Nutzer verwaltet; diese Planungsarbeit hat keine Git-Aktionen ausgeführt. Versionierung und Entwicklungsstart werden separat geklärt.

## Modellanbieter und Anmeldung

- Eino-Ausführung und Modellzugang sind getrennte Verantwortungen. Für jeden unterstützten Anbieter gibt es einen eigenen Eino-Chatmodell-Adapter hinter einem kleinen anbieterneutralen Modell-Port. Eine mapbasierte Factory/Registry wählt beim Laufstart anhand der konfigurierten Anbieterkennung den registrierten Adapter; unbekannte oder unvollständig konfigurierte Anbieter liefern einen verständlichen Fehler ohne Ersatzwahl. `domain/` und `app/` kennen Anbieterkennung, Verbindungsreferenz, Modell und Fähigkeiten, aber keine anbieterspezifischen Authentifizierungs- oder Transporttypen. Codex-Abo bleibt Standard; OpenRouter ist bewusste Alternative.
- OpenRouter.ai ist ein ausdrücklich wählbarer Modellanbieter pro Eino-Agent. Ein zentraler OpenRouter-API-Schlüssel wird mit `MS-03` in Settings erfasst und serverseitig geschützt verwahrt. Seine Nutzung wird mit `MS-04` je Organisation und Agent ausdrücklich freigegeben; die daraus angebotene Verbindungsreferenz gehört nur zu dieser Organisation. Key-Bestand, neue Organisationen, neue Agenten und Rotation erweitern Rechte nicht. Vor jedem neuen und noch nicht ausgeführten Aufruf werden Freigabe und Agentenzuordnung geprüft. Die Modellauswahl gehört zur Agentenkonfiguration; ein Anbieterfehler löst keinen stillen Wechsel aus.
- Settings verwaltet den Codex-Gerätecode-Ablauf und den zentralen OpenRouter-Key. Authentifizierungsart, Verbindung, Modell und nachgewiesene Fähigkeiten kommen aus der jeweiligen Anbieterkonfiguration; Gerätecode, Key-Nutzung, Sitzungsablauf, Erneuerung und Abmeldung bleiben beim zuständigen Anbieteradapter beziehungsweise Credential-Port. Eino erhält keine Anmeldegeheimnisse im Prompt oder als Toolparameter.
- Konfiguration und Agenten speichern nur Verbindungsreferenzen; Zugangsdaten bleiben ausschließlich auf dem Server hinter einem dafür vorgesehenen Port. Die lokale geschützte Speicherung mit getrenntem Masterschlüssel und restriktiven Dateirechten ist im Architekturplan festgelegt und wird mit `MS-07` und `OPS-02` umgesetzt und geprüft.
- Eino bleibt für Agentenschritte und die Toolausführung zuständig; die Anwendung validiert fachliche Zustände, Übergänge und Rechte vor jeder Wirkung. Organisationen können ihre Delegations- und Arbeitsabläufe konfigurieren, ohne globale feste Prozesskette oder allgemeine Workflow-DSL. Ein vollständiger Codex-Agentenlauf ist kein Nachweis für die gewünschte Eino-Modellanbindung.
- Der Codex App Server ist eine offiziell dokumentierte Integrationsmöglichkeit für den Codex-Adapter. Seine Anmeldung belegt allein keine allgemeine Modell-API für den Eino-Adapter.
- Der noch nötige Integrationsnachweis und die Quellen stehen in [modellzugang.md](modellzugang.md). Keine Produktimplementierung oder echte Anmeldung wurde im Rahmen dieser Planungsrecherche gestartet.
- Ein registrierter Testanbieter wird später über externe Konfiguration in denselben Auswahl-, Lauf-, Rechte- und Fehlerweg eingebunden. Diese Erweiterbarkeit benötigt keinen generischen Plugin-Lader und keine vorgezogenen Gemini-/Anthropic-Adapter; für sie wird weder V1-Unterstützung noch eine OAuth-/Abo-Anmeldung behauptet.
- Für Echtzeit-Voice-Streaming über OpenRouter/Eino bleibt die konkrete API-Fähigkeit offen. Streaming-Ereignisse, Audio-Transport, Unterbrechung und Rückbindung an Chat und Issue erhalten erst nach belastbarem Nachweis einen verbindlichen Adaptervertrag.

## Rechte an Ausführungsgrenzen

- Produktagenten erhalten niemals direkten Shell-, Terminal-, Prozess-, Interpreter-, HTTP- oder Dateisystemzugriff. Es gibt keine generische Ausnahme oder Modell-Exec-Schnittstelle. Eino orchestriert benannte Fachfähigkeiten; die Go-Anwendung validiert ID, Parameter, Zustand und Rechte und führt Side Effects über Ports und Adapter aus. Wrapper, Toolketten und Kinddelegation dürfen diese Grenze nicht umgehen.
- Die Anwendung schneidet bei Delegation Rechte auf die Schnittmenge der zulässigen Rechte zu; Kindaufträge und Unteragenten erhalten keine zusätzlichen Tool-, Daten- oder Delegationsrechte. Grenzen werden an jedem tatsächlichen Aufruf erneut geprüft. Codex CLI bleibt für ein Profil gesperrt, falls seine tatsächlichen Tool- und Nebenwege diese Grenze nicht beweisbar einhalten. Daraus folgt kein pauschales Sandbox- oder Isolationsversprechen.

- Modellgeneriertes Markdown wird als Inhalt behandelt und niemals ausgeführt. Ausschließlich die Go-Anwendung schreibt Arbeitsprodukte in einen begrenzten Organisations- und Sitzungs-Artefaktbereich; Rechercheagenten dürfen nur Markdown-Arbeitsprodukte, keine ausführbaren Skripte oder Dateien erzeugen oder persistieren. Dauerhafte organisationsbezogene Sitzungen und Arbeitskontexte sind an Rechte und Herkunft gebunden.
- Skills und optionale `SOUL.md`-Verhaltensbeschreibungen erteilen keine Rechte. Script-Skills dürfen nur vorab registrierte und geprüfte Aktionen über eine Aktions-ID mit validierten Parametern anfordern; die Go-Anwendung entscheidet und führt sie aus. Websuche ist ausschließlich ein freigegebenes Eino-Tool mit kontrolliertem Go-Adapter, kein allgemeiner HTTP- oder URL-Zugriff.

## Planungsstand der Schnittstellen

- Die geplante Paket- und Umsetzungsfolge steht unter [planung/architektur/](planung/architektur/). Sie ist ein Vorschlag für die spätere Implementierung, keine Validierung des bestehenden Vorlaufs.
- Zwischen Go-Webadapter und `ui/` dient bis zur endgültigen Festlegung eine explizite Template-Map je Ansicht als Datenvertrag. Fachmodelle bleiben in `domain/`; der Webadapter projiziert nur freigegebene Felder auf benannte Template-Werte. Die Vorschau lädt denselben Vertrag ausschließlich aus JSON-Fixtures.
- Die oben festgelegte Ausnahme für `main` und Godog folgt aus den vorgegebenen Go-Einstiegssignaturen; sie erweitert die Receiver-Regel nicht für Hilfsfunktionen.

## Arbeitsweise für die spätere Umsetzung

- Für die Projektarbeit an Agent Control Plane einschließlich neu angelegter Chats und sämtlicher Subagenten gilt **GPT-6 Sol mit Reasoning Medium** (`gpt-6-sol`, `medium`). Diese Vorgabe betrifft nicht laufende Trading-Aufträge.
- Der Hauptagent koordiniert die Arbeit; die Umsetzung erfolgt mit Subagenten. Unabhängige Teilbereiche dürfen bei sinnvoller Parallelisierung eigene Chats mit eigenen Subagenten bekommen. Der Nutzer hat diese Delegation und zusätzliche Chats ausdrücklich erlaubt.
- Alle neuen Chats und Dateiarbeiten für Agent Control Plane gehören ausschließlich in das gespeicherte Projekt Agent Control Plane; keine globalen oder projektlosen Aufgaben.
- Regelmäßige Code Reviews erfolgen durch einen separaten Subagenten. Konkrete Befunde werden behoben.
- Der Hauptagent koordiniert mehrere echte Chats ausschließlich im gespeicherten Projekt. Sinnvolle hohe Parallelität setzt klare Dateibesitzer und integrierbare Schnittstellen voraus; Routineprobleme lösen die jeweiligen Agenten eigenständig.
- Die Hauptkoordination pflegt `PLANUNG.md` mit Story, Status, zuständigem Chat/Subagenten, Review und Blockade. Git wurde nach dem Zurücksetzen durch den Nutzer verwaltet; diese Planungsarbeit führt keine Git-Aktionen aus. Nach gesondert geklärter Versionierung und Entwicklungsfreigabe kann jede Entwicklungsstory in einem eigenen Feature-Branch und isolierten Worktree bearbeitet werden. Ein fertiger Story-Branch wird separat gegen `main` geprüft und nach behobenen Befunden zeitnah integriert; Routinekonflikte lösen die Besitzer, echte Blockaden gehen an die Hauptkoordination.
- Nach relevanten Änderungen oder konkreten verbleibenden Risiken gezielt nachprüfen. Keine endlosen Review-Schleifen ohne neuen Anlass.
- Alle ausführenden und Review-Agenten dieses Projekts verwenden GPT-6 Sol / Medium und halten die Architektur-, Stil- und Godog-Vorgaben ein.
- Diese Arbeitsweise gilt ab dem freigegebenen Implementierungsstart. Zusätzliche Chats und Subagenten werden für konkret abgegrenzte Arbeit eingesetzt.
- Die Umsetzung erfolgt eigenständig in integrierbaren Schritten über mehrere Stunden: zuerst ein lokal lauffähiger POC, danach der vollständige beauftragte Umfang. Keine automatische Veröffentlichung, kein GitHub-Push und kein externes Deployment.
