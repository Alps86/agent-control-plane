# Stories zur Zusammenarbeit

Stand: 24. September 2026. Dies ist eine Planungsgrundlage, keine Implementierungsabnahme. Die primären Nummern stammen aus [story-index.json](../story-index.json); `ZA-01` bis `ZA-07` bleiben Fachkategorien. Die früheren Szenarien aus `spec/features/zusammenarbeit/` sind als [Markdown-Akzeptanzbeispiele](../akzeptanzbeispiele/bestand/zusammenarbeit/) gesichert. Sie ergänzen die bestehenden Paperclip-orientierten Grundlagen für Organisationen, Agenten, Aufgaben, Delegation, Läufe, Rechte und Verlauf. Discovery und GrillMe sind spätere Erweiterungen des gemeinsamen Arbeitswegs.

## Gemeinsame Fachregeln

- **Chat** ist eine dauerhafte Unterhaltung in einer Organisation. Eine Nachricht kann an einen einzelnen Agenten oder ein Team gerichtet sein. Teamchat bedeutet sichtbare Beiträge mehrerer berechtigter Agenten im selben Verlauf; er erzeugt keine neue Ausführungsart neben Eino und Codex CLI.
- **Gemeinsame Arbeit** hat eine überprüfbare Quelle: Chatnachricht, Aufgabe, Lauf, Delegation, Entscheidung und Arbeitsprodukt sind miteinander verlinkt. Eine Antwort wie „erledigt“ ersetzt weder eine angelegte Aufgabe noch ein prüfbares Ergebnis.
- **To-do** bezeichnet eine Aufgabe im bestehenden Aufgabenmodell mit Organisation, optionalem Team, Zuständigkeit, Status und Herkunft. Team- und Organisationsansichten sind Sichten auf diese Aufgaben, keine voneinander unabhängigen Schattenlisten. Sichtbarkeit und Bearbeitung richten sich nach Organisations-, Daten- und Delegationsgrenzen; eine Teamzuordnung verleiht keine zusätzlichen Rechte.
- **Organisationsablauf** ist je Organisation über erlaubte Aufgabenstatus-, Zuweisungs- und Delegationsübergänge konfigurierbar. Jeder tatsächliche Übergang wird gegen aktuellen Zustand und Rechte fachlich geprüft; Berichtswege allein erlauben keine Aktion. Dafür wird keine allgemeine Workflow-DSL eingeführt.
- **Aktion** bedeutet eine tatsächliche Zustandsänderung oder externe Ausführung, etwa Aufgabe anlegen, zuweisen, delegieren, Lauf starten oder Datei schreiben. Reine Chatantworten, Vorschläge und Zusammenfassungen sind keine Aktionen. Bei jeder Aktion werden Auftrag, ausführender Akteur, Ziel, Berechtigung, Status und Ergebnis erfasst. Fehlende Berechtigung oder unklare Zielzuordnung führen zu einer sichtbaren Rückfrage oder Ablehnung statt zu einer stillen Ersatzaktion.
- **Ausführungsgrenze** bedeutet ausschließlich benannte Fachfähigkeiten. Agenten erhalten keinen direkten Shell-, HTTP- oder Dateisystemzugriff. Eino orchestriert; die Go-App validiert und führt Side Effects über Ports/Adapter aus. Ein Chat, eine dauerhafte organisationsbezogene Session, ein Skill oder SOUL.md erteilt keine Rechte; jede Aktion wird neu geprüft.
- Sprache ist ein weiterer Ein- und Ausgabekanal für denselben Chat und dieselben Aufgaben. Das bestätigte Transkript und die daraus ausgelöste Aktion bleiben auch als Text nachvollziehbar. Ein Sprachbefehl umgeht keine Rechteprüfung und erhält keine Sondersemantik. OpenRouter.ai ist der festgelegte Anbieterweg für Sprache mit Eino; konkrete Modelle, Provider-Routen und Sprachschnittstellen bleiben bis zum technischen Nachweis offen.
- Die Anwendung hat in V1 keine menschliche Anmeldung. „Ich“ bezeichnet die lokale Bedienperson; Agentenidentität und Organisationskontext werden fachlich festgelegt. Review-, Reporting- und Monitoringzuständigkeiten sind keine menschlichen Anmelderollen. Budgetverwaltung bleibt ausgeschlossen.

## Stories und Abnahme

### Story-91 · ZA-01 – Einzel- und Teamchat

**Nutzen:** Als Bedienperson möchte ich mit einem Agenten oder einem Team in einer Organisation sprechen und den Verlauf später fortsetzen, damit Fragen, Aufträge und Antworten in einem gemeinsamen Arbeitskontext bleiben.

**Akzeptanzkriterien:**

1. Ich kann aus einer Organisation einen Einzelchat mit einem einsatzbereiten Agenten und einen Teamchat mit einem dort definierten Team beginnen. Adressaten und Organisationskontext sind vor dem Senden sichtbar.
2. Jede gesendete Nachricht erscheint einmal mit Absender, Zeitpunkt und Zustell- beziehungsweise Bearbeitungszustand. Beiträge mehrerer Teamagenten sind eindeutig zugeordnet; der Verlauf behält ihre Reihenfolge und bleibt nach Neustart über eine organisationsgebundene Session zugänglich. Die Session kann nur im berechtigten Organisationskontext fortgesetzt werden und verleiht keine neuen Rechte.
3. Ein Agent kann auf eine Nachricht antworten oder einen nachvollziehbaren Lauf daraus starten. Ein Lauf zeigt Status, Ergebnis oder Fehler und verweist auf die auslösende Nachricht; ein fehlgeschlagener Lauf löscht die Nachricht nicht.
4. Fortsetzung verwendet nur zulässigen Gesprächskontext. Ist eine Adressierung ungültig, ein Agent nicht einsatzbereit oder eine Verbindung nicht verfügbar, sieht die Bedienperson den Grund und kann den Verlauf weiterhin lesen.

**Abhängigkeiten:** ZA-02 für die Teamidentität und deren Organisationsgrenze; Organisation, Agentenkonfiguration und Persistenz ([organisationen/organisation-anlegen.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/organisationen/organisation-anlegen.md), [agenten/agenten-anlegen.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/agenten/agenten-anlegen.md), [betrieb/betrieb.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/betrieb/betrieb.md)). Echte Eino-Antworten benötigen RUN-01 und MS-06 sowie die für den gewählten Zugang geltenden MS-01/02/05 beziehungsweise MS-03/04/05; Codex-CLI-Antworten benötigen RUN-01. Ein vorhandener, später unzulässiger Chat bleibt lesbar und sperrt neue Zustellung. Primäre Voraussetzungen: Story-90 (ZA-02), Story-27 (RUN-01), Story-08 (MS-01), Story-09 (MS-02), Story-23 (MS-03), Story-24 (MS-04), Story-25 (MS-05), Story-26 (MS-06).

### Story-90 · ZA-02 – Team- und Organisationsarbeit mit gemeinsamen To-dos

**Nutzen:** Als Bedienperson möchte ich Teamzuständigkeiten und die offenen Arbeiten einer Organisation gemeinsam sehen, damit mehrere Agenten auf denselben Stand hinarbeiten und keine Aufgaben zwischen Chat und Aufgabenliste verschwinden.

**Akzeptanzkriterien:**

1. Teams gehören zu genau einer Organisation. Mitglieder, koordinierende Zuständigkeit und Berichtswege sind im Organigramm sichtbar. Ein Team kann ein gemeinsames To-do übernehmen; einzelne Arbeitsschritte werden konkret zuständigen Agenten zugeteilt. Eine Änderung der Teamzuordnung wird nachvollziehbar und ändert nicht rückwirkend die Herkunft früherer Beiträge.
2. Die Teamansicht zeigt die ihr zugeordneten offenen, laufenden, blockierten und abgeschlossenen Aufgaben mit Verantwortlichen. Die Organisationsansicht fasst ihre Teams und organisationsweite Aufgaben ohne Dubletten zusammen. Statusänderungen erscheinen in beiden passenden Ansichten und nach Neustart.
3. Jede Aufgabe kann aus der Ansicht geöffnet und im bestehenden Aufgabenworkflow bearbeitet werden. Teamzuordnung, Zuweisung, Priorität und Blockade sind erkennbar; unzugewiesene Arbeit bleibt ausdrücklich unzugewiesen.
4. Aufgaben und Ergebnisse einer anderen Organisation sind weder im Teamkontext noch in To-do-Ansichten sichtbar. Teammitgliedschaft erweitert weder Daten-, Tool- noch Delegationsrechte. Übergaben benötigen sowohl einen möglichen Weg im Organigramm als auch einen nach aktueller Organisationsregel zulässigen Übergang und die Rechte beider Agenten; Auftraggeber, koordinierendes Team und ausführender Agent bleiben getrennt nachvollziehbar. Die Chatkontextgrenze wird anschließend in ZA-01 geprüft.

**Abhängigkeiten:** Organisation, Organigramm und Delegationswege, Projekt und Aufgabenworkflow ([organisationen/organisation-anlegen.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/organisationen/organisation-anlegen.md), [organisationen/delegation.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/organisationen/delegation.md), [agenten/agenten-anlegen.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/agenten/agenten-anlegen.md), [projekte/projekte.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/projekte/projekte.md), [aufgaben/aufgaben.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/aufgaben/aufgaben.md)), Berechtigungen ([berechtigungen/rechte.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/berechtigungen/rechte.md)). ZA-02 legt Teamidentität, Mitgliedschaft und To-do-Sichten unabhängig vom Chat an; ZA-01 verwendet diese Teamidentität später als Chatadressat. Keine andere ZA-Story als Voraussetzung.

### Story-93 · ZA-03 – Nachricht, Aufgabe und Arbeitsprodukt verbinden

**Nutzen:** Als Bedienperson möchte ich einen Auftrag im Chat als echte Aufgabe verfolgen und vom Ergebnis zur Ausgangsnachricht zurückkehren, damit besprochene Arbeit weder unsichtbar noch doppelt erfasst wird.

**Akzeptanzkriterien:**

1. Ein hinreichend bestimmter Chatauftrag kann eine Aufgabe im ausgewählten Projekt erzeugen oder eine vorhandene Aufgabe eindeutig referenzieren. Ist Projekt, Ziel oder Empfänger unklar, fordert das System eine Klärung an und legt keine Aufgabe auf Verdacht an.
2. Die Ausgangsnachricht zeigt, welche Aufgabe oder welcher Lauf daraus entstand. Die Aufgabe nennt die Nachricht als Herkunft; Status, Zuständigkeit, Kommentare und Blockaden stammen aus dem bestehenden Aufgabenmodell.
3. Ein erzeugtes Arbeitsprodukt wird mit Aufgabe und Lauf verknüpft. Vom Chat sind Aufgabe, Lauf und Arbeitsprodukt erreichbar; von der Aufgabe führt ein Weg zur relevanten Chatstelle zurück. Eine bloße Behauptung des Agenten wird nicht als Arbeitsprodukt ausgegeben.
4. Der sichtbare Verlauf trennt Antwort, Vorschlag, angeforderte Aktion, laufende Aktion, Erfolg, Fehler und Ablehnung. Nur eine erfolgreich bestätigte Zustandsänderung wird als erledigte Aktion dargestellt. Eine wiederholte Zustellung erzeugt keine zweite Aufgabe oder Datei.

**Abhängigkeiten:** ZA-01, ZA-02 für Teamzuordnung, bestehende Aufgaben, Läufe und Arbeitsprodukte ([aufgaben/aufgaben.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/aufgaben/aufgaben.md), [ausfuehrung/lauf.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/ausfuehrung/lauf.md), [ausfuehrung/aktivitaet.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/ausfuehrung/aktivitaet.md)), Rechte ([berechtigungen/rechte.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/berechtigungen/rechte.md)). Die verbindliche Ausführung mutierender Chataktionen setzt ZA-06 voraus. Primäre Voraussetzungen: Story-91 (ZA-01), Story-90 (ZA-02), Story-92 (ZA-06).

### Story-94 · ZA-04 – Delegation und Reporting im Chat

**Nutzen:** Als beauftragender Agent oder als Bedienperson möchte ich Übergaben, Fortschritt und Rückmeldungen im gemeinsamen Verlauf sehen, damit Teamarbeit ohne versteckte Nebenaufträge steuerbar bleibt.

**Akzeptanzkriterien:**

1. Eine Delegation aus dem Chat benennt Ausgangsaufgabe, Auftraggeber, Empfänger, erwartetes Ergebnis und gegebenenfalls Teilaufgabe. Sie nutzt dieselbe aktuelle Organisationsregel, fachliche Zustandsprüfung und Rechteprüfung wie die Aufgabenoberfläche; die Übergabe erscheint in Chat und Aufgabenverlauf.
2. Empfänger können Fortschritt, Blockade, Ergebnis und Fehler an der Aufgabe melden. Auftraggeber und zuständige Reviewer sehen die Rückmeldung an der richtigen Aufgabe; ein Teambericht fasst tatsächliche Aufgaben- und Laufzustände zusammen und verlinkt die Belege.
3. Ein unzulässiger Delegationsweg oder Zugriff über Organisations- und Datengrenzen wird vor der Übergabe verweigert. Die Aufgabe bleibt beim bisherigen Verantwortlichen, und die Ablehnung ist ohne fremde Daten im Verlauf erkennbar.
4. Ein Chatbericht behauptet keine Freigabe, solange die fachliche Prüfung offen oder zurückgegeben ist. Reviewentscheidung und Begründung bleiben beim bestehenden Reviewablauf und sind aus dem Chat erreichbar.

**Abhängigkeiten:** ZA-01 bis ZA-03, Delegation ([organisationen/delegation.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/organisationen/delegation.md)), Rechte ([berechtigungen/rechte.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/berechtigungen/rechte.md)), Review und Monitoring ([review/review-und-monitoring.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/review/review-und-monitoring.md)). Mutierende Übergaben setzen ZA-06 voraus. Primäre Voraussetzungen: Story-91 (ZA-01), Story-90 (ZA-02), Story-93 (ZA-03), Story-92 (ZA-06).

### Story-104 · ZA-05 – Voice im Chat bedienen und nachvollziehen

**Nutzen:** Als Bedienperson möchte ich einen verfügbaren Voice-Turn aus dem Einzel- oder Teamchat starten und dessen bestätigten Verlauf dort wiederfinden, damit Sprache und Text denselben Arbeitsstand zeigen.

**Akzeptanzkriterien:**

1. Im Einzel- oder Teamchat kann ich den dort verfügbaren Voice-Turn für einen berechtigten Eino-Agenten starten. Chatkontext, Zielagent und tatsächlicher Turn-Modus sind vor dem Start erkennbar. Ist der Voice-Pfad nicht verfügbar, bleibt der Textchat bedienbar und zeigt den Grund.
2. Das vom Voice-Pfad gelieferte gesendete Eingabetranskript, die empfangene Textfassung der Antwort und der bestätigte Turn-Zustand erscheinen im ursprünglichen Chat mit Sprecher, Zeit und Zuordnung zur Nachricht. Die Chatprojektion erzeugt keinen zweiten Auftrags- oder Nachrichtenzustand.
3. Eine vom Voice-Pfad bestätigte Fachaktion, etwa Issue- oder Agentenanlage, wird mit ihrem tatsächlichen Ergebnis und den vorhandenen Verweisen auf Aufgabe, Agent, Lauf und Arbeitsprodukt im Chat angezeigt. Von dort sind die Fachobjekte erreichbar; ein Vorschlag oder unbestätigter Voice-Status erscheint nicht als ausgeführte Aktion.
4. Bediene ich im Chat die vom Voice-Pfad angebotene Unterbrechung oder Wiederaufnahme, zeigt die Chatansicht den zurückgemeldeten Zustand. Bereits angenommene oder ausgeführte Aktionen bleiben sichtbar; eine Korrektur erscheint als neuer, zugeordneter Auftrag, wenn der Voice-Pfad sie als solchen meldet. Der Chat selbst behauptet weder einen erfolgreichen Abbruch noch eine Rücknahme ohne Bestätigung.
5. Ein späteres Angebot echter bidirektionaler Echtzeit wird im Chat erst nach bestandenem `VO-07` als solches bezeichnet. Der zuvor nachgewiesene Sprach-Turn mit gestreamter Antwort bleibt als eigener Modus erkennbar.

**Abhängigkeiten:** ZA-01 und ZA-02 für Chatkontext und Sichtbarkeit, ZA-03 und ZA-06 für Aktionsverknüpfung und Wiederholung. `CHV-01` liefert den Voice-Chat-Vertrag; `VO-01` und `VO-02` liefern den geprüften OpenRouter-/Eino-Sprach-Turn samt Transkript und Antwort, `VO-03`/`VO-04` die sprachliche Issue-/Agentenanlage und `VO-05` Abbruch und Wiederaufnahme. `VO-07` ist nur für die Echtzeitbezeichnung und -bedienung erforderlich. Providertransport, Audiofähigkeit, Anlage und Abbruch werden kanonisch in `spec/planung/modelle-sprache/stories.md` abgenommen; ZA-05 prüft deren Projektion und Bedienung im gemeinsamen Chat. Primäre Voraussetzungen: Story-91 (ZA-01), Story-90 (ZA-02), Story-93 (ZA-03), Story-92 (ZA-06), Story-103 (CHV-01), Story-95 (VO-01), Story-96 (VO-02), Story-97 (VO-03), Story-98 (VO-04), Story-99 (VO-05).

### Story-92 · ZA-06 – Eindeutige Aktionssemantik bei Wiederholung und Verbindungsabbruch

**Nutzen:** Als Bedienperson möchte ich nach einem Sende- oder Verbindungsfehler zuverlässig erkennen, ob ein Auftrag ausgeführt wurde, damit Wiederholungen keine doppelten Aufgaben, Läufe oder externen Aktionen auslösen.

**Akzeptanzkriterien:**

1. Für jede mutierende Chat- oder Sprachaktion ist eine stabile Auftragskennung mit Ausgangsnachricht und beabsichtigtem Ziel sichtbar beziehungsweise abrufbar. Die öffentliche Anwendungsgrenze prüft bereits beim Annehmen, ob ein wiederholtes Senden, ein Retry nach Timeout, Seitenneuladen oder Wiederverbinden denselben Auftrag meint; derselbe Auftrag liefert seinen gespeicherten Stand und legt weder eine zweite Aufgabe noch einen zweiten Lauf an.
2. Die Oberfläche unterscheidet „nicht angenommen“, „angenommen“, „in Arbeit“, „erfolgreich“, „fehlgeschlagen“ und „Ausgang noch unklar“. Bei unklarem Ausgang wird der tatsächliche lokale und, soweit möglich, externe Stand abgefragt. Bis zur Klärung wird keine neue Ausführung mit neuer Kennung ausgelöst. Bei einer externen Gegenseite ohne überprüfbaren Status bleibt die mögliche Wirkung ausdrücklich unklar und verlangt eine bewusste Entscheidung zum weiteren Vorgehen.
3. Rechte werden sowohl bei Annahme als auch unmittelbar vor tatsächlicher Ausführung geprüft. Wird eine Freigabe dazwischen entzogen, verhindert das die noch ausstehende Aktion, auch wenn eine frühere Chatantwort sie angekündigt hatte. Verweigerung und Fehler erscheinen mit verständlichem Grund und ohne Geheimnisse oder fremde Daten.
4. Ein bewusst neuer Versuch nach fachlichem Fehler ist als neuer Auftrag erkennbar und verweist auf den vorherigen. Externe Aktionen nutzen eine eindeutige Gegenstellenkennung oder einen nachweisbaren Abgleich, sofern verfügbar. Ohne diese Sicherung wird keine genau einmalige externe Wirkung behauptet; automatischer Retry kann die unbekannte Wirkung nicht verdoppeln.

**Abhängigkeiten:** ZA-01 und dauerhafter Nachrichten-/Aktionsverlauf ([betrieb/betrieb.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/betrieb/betrieb.md), [ausfuehrung/aktivitaet.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/ausfuehrung/aktivitaet.md)), Berechtigungen ([berechtigungen/rechte.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/berechtigungen/rechte.md)). ZA-06 ist eine Freigabebedingung für mutierende Teile von ZA-03, ZA-04 und ZA-05. Primäre Voraussetzungen: Story-91 (ZA-01).

### Story-100 · ZA-07 – Spätere Discovery und GrillMe

**Nutzen:** Als Bedienperson möchte ich nach dem stabilen Kernweg ein Vorhaben gemeinsam erkunden und Annahmen kritisch befragen lassen, damit aus unscharfen Ideen prüfbare Ziele und Aufgaben werden.

**Akzeptanzkriterien:**

1. Discovery ist ein ausdrücklich gestarteter, beendbarer Modus innerhalb eines vorhandenen Chats. Der Agent sammelt Ziel, offene Fragen, Annahmen, Randbedingungen und mögliche nächste Schritte mit Quellenbezug zum Gespräch; er legt ohne erkennbaren Auftrag keine Aufgaben an.
2. GrillMe ist ein ausdrücklich gestarteter, beendbarer Fragemodus. Der Agent stellt kurze, konkrete Rückfragen, wartet auf Antworten und macht Widersprüche oder fehlende Entscheidungen sichtbar, ohne Antworten zu erfinden oder den Modus still zu wechseln.
3. Ergebnisse beider Modi bleiben als bearbeitbare Zusammenfassung im Chat. Erst ein konkreter Aktionsauftrag führt gemäß ZA-03 und ZA-06 zu Zielen, Aufgaben oder Delegationen. Herkunft und eventuelle Freigabe sind rückverfolgbar.
4. Die Modi sind optional. Einzelchat, Teamchat, To-dos und reguläre Ausführung funktionieren vollständig ohne sie; Organisations- und Toolrechte gelten unverändert.

**Abhängigkeiten:** Abgenommener Text- und Aktionskern ZA-01 bis ZA-04 sowie ZA-06, Ziele ([ziele/ziele.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/ziele/ziele.md)) und Aufgaben ([aufgaben/aufgaben.feature (frühere Quelle; gesichert)](../akzeptanzbeispiele/bestand/aufgaben/aufgaben.md)). ZA-05 ist keine Voraussetzung für die textbasierten Interviewmodi. Die genaue Gesprächsführung und Qualitätsbeispiele werden vor Implementierung als zusätzliche fachliche Beispiele konkretisiert. Primäre Voraussetzungen: Story-91 (ZA-01), Story-90 (ZA-02), Story-93 (ZA-03), Story-94 (ZA-04), Story-92 (ZA-06).

## Reihenfolge und Abnahmegrenzen

1. **Grundlage:** Bestehende Organisations-, Agenten-, Aufgaben-, Lauf-, Rechte- und Persistenzstories sowie der praktische Nachweis `MS-01` für den eigenen Eino-Chatmodell-Adapter. Für den gewählten Eino-Zugang gelten zusätzlich `MS-02/05` beim Abo oder `MS-03/04/05` bei OpenRouter. Der bereits geplante Paperclip-Basisumfang bleibt vollständig; diese Zusammenarbeit ergänzt ihn.
2. **Gemeinsamer Textweg:** ZA-02 liefert zuerst Teams und Aufgabenansichten ohne Chatvoraussetzung. ZA-01 nutzt danach die festgelegte Teamidentität und liefert dauerhafte Einzel- und Teamgespräche. Die Vorbereitung beider Stories kann parallel erfolgen; ihre Abnahme folgt diesem Gate.
3. **Aktionssicherheit:** ZA-06 vor der Freigabe mutierender Chatbefehle. Anschließend ZA-03 und ZA-04 mit denselben Aufgaben-, Delegations- und Reviewregeln wie außerhalb des Chats.
4. **Voice-Integration im Chat:** Nach `VO-01`/`VO-02` und dem Vertrag `CHV-01` bildet ZA-05 einen vorhandenen Sprach-Turn im gemeinsamen Verlauf ab. Fachaktionen und Abbruch stammen aus `VO-03` bis `VO-05`; deren Chatbezug wird hier sichtbar geprüft. Die Echtzeitbezeichnung setzt gesondert `VO-07` voraus.
5. **Spätere Gesprächsmodi:** ZA-07 nach Abnahme des Kerns und mit eigenen Qualitätsbeispielen für Discovery und GrillMe.

Jede Story wird über öffentliche Anwendungsgrenzen mit Godog geprüft; sichtbare Bedienwege werden zusätzlich im Browser geprüft, sobald ein nutzbarer Stand verfügbar ist. `spec/features/zusammenarbeit/` enthält die ausführbaren fachlichen Beispiele. Ein UI-Prototyp oder ein Modelltext allein erfüllt keine Story, deren Kriterium eine dauerhafte Aktion oder Rechteprüfung verlangt.
