# Produktanforderungen

Stand: 24. September 2026. Aktuelle Phase: vollständiger Quellen- und Feature-Abgleich, Stories, Akzeptanzkriterien und Abhängigkeiten. Die Produktimplementierung ist bis zum Abschluss dieser Planung gestoppt. Die spätere eigenständige Umsetzung bleibt das Ziel.

## Ziel

- Eigenständige Go-Anwendung mit eigenem Go-Server nach dem Funktionsvorbild [Paperclip](https://github.com/paperclipai/paperclip). Angestrebt ist dessen Funktionsspektrum zur Verwaltung und Koordination von Agentenorganisationen.
- Agenten können auch Nichtentwickler über eine verständliche UI anlegen und konfigurieren, möglichst mit Vorlagen.
- Eino ist die primäre Ausführungsart. Codex CLI ist nur für ein Agentenprofil wählbar, wenn der Adapter die unten genannten Grenzen nachweisbar einhält; andernfalls bleibt dieses Profil gesperrt.

## Modellzugang für Eino

- Eino-Agenten verwenden einen anbieterneutralen Modellzugang mit einem eigenen Eino-Chatmodell-Adapter je unterstütztem Anbieter. Als Standard ist der Modellzugang über das ChatGPT-/Codex-Abonnement gewünscht. Die Wahl eines Anbieters ändert weder Agentenablauf noch fachliche Rechte.
- Settings verwaltet die unterstützten Zugänge: Codex-Abo per Device-Code-Authentifizierung nach dem Bedienvorbild von Hermes Agent sowie einen zentralen OpenRouter-Key. Beim Gerätecode sind Start, Anmeldeseite, Code, Verbindungsstatus und Fehler sichtbar. Welche Authentifizierungsart und Fähigkeiten angeboten werden, stammt aus der Anbieter- und Verbindungskonfiguration.
- Anbieter, Authentifizierungsart, Verbindung, Modell und geprüfte Fähigkeiten sind einfach konfigurierbar. Ein abweichender Anbieter darf gezielt gewählt werden; bei Anmeldefehlern oder ausgeschöpften Anbieterlimits erfolgt kein stiller Wechsel auf kostenpflichtige API-Nutzung.
- Die Anbieteranmeldung ist keine Anmeldung an Agent Control Plane. Die Anwendung selbst bleibt ohne Login und Registrierung.
- Eino behält die Agentensteuerung und die Prüfung freigegebener Tools. Der Codex-CLI-Agent bleibt eine eigenständige Ausführungsart.
- **OpenRouter.ai** ist für jeden Eino-Agenten ausdrücklich als Modellanbieter wählbar. Der Betreiber kann einen zentralen OpenRouter-API-Schlüssel einfach in den Settings hinterlegen. Seine Nutzung wird je Organisation bewusst freigegeben und die Verbindung einem Agenten dieser Organisation ausdrücklich zugeordnet; das bloße Hinterlegen des Schlüssels gibt keinem Agenten Zugriff. Pro Agent ist ein verfügbares OpenRouter-Modell wählbar. Die Anbieterwahl ändert die Ausführungsart Eino nicht.
- Der direkte Abo-Zugang aus einem eigenen Eino-Chatmodell ist eine verbindliche Zielanforderung mit noch offenem Integrationsnachweis. Quellenstand, Stories und Prüfkriterien stehen in [modellzugang.md](modellzugang.md).
- Die Modellgrenze soll später um registrierte Anbieter erweitert werden können, ohne Agentenablauf, Fachregeln oder Rechteprüfungen pro Anbieter zu ändern. Gemini und Anthropic sind mögliche spätere Integrationen, keine V1-Lieferung und keine Zusage für OAuth- oder Abo-Zugang.

## Sprache und Zusammenarbeit

- Sprachbedienung soll Issues und Agenten anlegen sowie Aufgaben steuern können. Gespräche mit Agenten gibt es einzeln und als Teamchat; Chats können mit Issues verknüpft werden.
- Arbeit und Gesprächsverläufe sind sowohl im Team als auch entlang der Organisationsstruktur zugänglich und steuerbar, jeweils innerhalb der geltenden Rechte.
- Echtzeit-Voice-Streaming mit Agenten über den gewählten OpenRouter-/Eino-Weg ist ein Produktziel. Die tatsächlich nutzbare API-Fähigkeit, Latenz und Unterbrechbarkeit sind noch zu prüfen; die Planung darf diese Eigenschaften nicht als nachgewiesen ausgeben.
- Project Discovery und „GrillMe“ sind spätere geplante Bedienabläufe. Ihr genauer Funktionsschnitt und ihre Abnahme werden in der vollständigen Planung geklärt.

## Organisation und Arbeit

- Agenten sind in einer Organisation mit Zuständigkeiten und Rollen für Review, Reporting und Monitoring angeordnet.
- Jede Organisation konfiguriert ihre Delegations- und Arbeitsabläufe innerhalb der fachlichen Rechte. Es gibt keine globale feste Prozesskette und keine allgemeine Workflow-DSL; Eino und die Anwendung prüfen Zustände, Übergänge und Berechtigungen.
- Agenten erhalten konkrete Aufgaben, können geplant oder per Cron laufen, Arbeit delegieren und Fortschritt sowie Ergebnisse dauerhaft verwalten.
- Flexible Aufgabenbearbeitung bleibt erforderlich.

## Rechte und Ausführung

- Produktagenten erhalten niemals direkten Shell-, Terminal-, Prozess-, Interpreter-, HTTP- oder Dateisystemzugriff, auch nicht durch eine generische Freigabe. Sie verwenden nur benannte Fachfähigkeiten. Eino orchestriert; die Go-Anwendung validiert Zustände, Rechte und Parameter und führt Wirkungen über Ports und Adapter aus.
- Tool-, Daten- und Delegationsberechtigungen werden in der Anwendung an der tatsächlichen Ausführungsgrenze geprüft, auch bei delegierter Arbeit. Ein Kindauftrag oder Unteragent erhält nie mehr Rechte als Auftraggeber und ausführender Agent gemeinsam zulassen; Umwege über andere Tools bleiben gesperrt.
- Aufwendige Containerverwaltung pro einfachem Agenten soll vermieden werden. Die Fachfähigkeitsgrenze verlangt eine Prüfung von Terminal-, Prozess-, Interpreter-, HTTP- und Dateisystem-Umwegen; daraus folgt kein pauschales Sandbox- oder Isolationsversprechen.
- Codex CLI darf nur eingesetzt werden, wenn seine tatsächlichen Tool- und Nebenwege die Fachfähigkeitsgrenze beweisbar einhalten; sonst bleibt das Codex-CLI-Profil gesperrt.

- Agenten dürfen Markdown-Inhalt für Arbeitsprodukte liefern. Nur die Go-Anwendung schreibt ihn in einen begrenzten Organisations- und Sitzungs-Artefaktbereich. Ein Rechercheagent erzeugt oder persistiert ausschließlich Markdown-Arbeitsprodukte, keine ausführbaren Skripte oder Dateien; Markdown wird nie ausgeführt.
- Skills und optionale Verhaltensbeschreibungen wie `SOUL.md` erteilen keine Rechte. Ein Skill mit Script löst höchstens eine vorab registrierte und geprüfte Fachaktion per ID und validierten Parametern in der Go-Anwendung aus; Modell-Shelltext und generisches Exec sind ausgeschlossen. Websuche läuft nur über ein freigegebenes Eino-Tool und einen kontrollierten Go-Adapter, nie über einen allgemeinen HTTP-/URL-Zugriff des Agenten. Organisationsbezogene Sitzungen und Arbeitskontexte bleiben dauerhaft nachvollziehbar.

## Lauftransparenz und Nutzung

- Eino und ein gegebenenfalls zugelassenes Codex-CLI-Profil liefern einheitliche strukturierte Ereignisse mit Level, Ereignisart, Zeit und Bezügen zu Organisation, Projekt, Aufgabe, Agent, Sitzung, Lauf und Versuch sowie dem tatsächlich verwendeten Provider und Modell. Ein regelmäßiger Heartbeat zeigt die letzte belegte Phase und ihren Zeitpunkt; er erfindet weder Prozentfortschritt noch Modellaktivität.
- Die Go-Anwendung speichert gemeldete Token-Nutzung dauerhaft in SQLite je Aufgabe, Lauf, Versuch, Provider und Modell. Input-, Output- und Reasoning-Token werden nur erfasst, soweit der Adapter sie liefert. Fehlende, unvollständige oder geschätzte Werte bleiben als solche gekennzeichnet statt als Null zu erscheinen. Provider-Zählweise bleibt nachvollziehbar; Reasoning kann in Output enthalten sein und wird dann nicht doppelt addiert. Streaming, Retry, Wiederaufnahme, Delegation und Abbruch dürfen Verbrauch nicht doppelt zählen oder bereits gemeldete Nutzung verlieren. Logs und Nutzungsdaten enthalten keine Secrets, vollständigen Prompts oder Gedankengänge.

## Umfang und Kosten

- Budgetverwaltung, Budgetfelder, Budgetlimits sowie ein Abo- oder Kostenmodell sind für die erste Version ausdrücklich ausgeklammert. Token-Nutzung ist Diagnose und Nachvollziehbarkeit, keine Budgetabrechnung.
- Passende Modelle, begrenzte Aufträge und bedarfsgerechter Kontext sollen Kosten senken. Eine automatische Ersparnis allein durch Eino oder kürzere Sitzungen ist nicht belegt.
- SQLite ist als Datenbank gewählt. Weiteres Deployment bleibt offen. Es gibt **keinen Login und keine Login- oder Registrierungsoberfläche**. Weitere Cloud-Dienste sind nicht ausgewählt.
- Nach abgeschlossener Feature- und Story-Planung wird zuerst ein schnell nutzbarer POC mit einer Paperclip-ähnlich gestalteten HTML/CSS-Oberfläche integriert; danach wird der vollständige beauftragte Funktionsumfang weiter umgesetzt. Der POC reduziert das Produktziel nicht.
- Die statische Oberfläche und ihre Vorschau nutzen ausschließlich JSON-Fixtures für Ansichtsdaten. Die echte App liefert echte Daten über die öffentliche UI-Grenze; UI-Vorlagen enthalten keine fachlichen Datenmodelle und werden als lesbare Go-HTML-Templates gepflegt.
- Paperclip bleibt die vollständige fachliche Basis für den bestätigten Umfang; zusätzliche Sprach-, Chat- und Modellzugangsfunktionen erweitern diese Basis. Explizit ausgenommene Budget- und Login-Funktionen bleiben ausgenommen.
