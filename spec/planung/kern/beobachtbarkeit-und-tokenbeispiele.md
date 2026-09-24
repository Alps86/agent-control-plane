# Öffentliche Akzeptanzbeispiele: Lauf-Logs, Heartbeat und Token-Usage

Stand: 24. September 2026. Dies sind Planungsbeispiele für RUN-02 bis RUN-04, ACT-04/05 und MON-01, keine fertigen Feature-Dateien oder bestandenen Tests. Spätere ausführbare Fälle liegen unter `features/app/<fachbereich>/` und `features/ui/<fachbereich>/`.

## Gleiche Logstruktur für Eino und Codex CLI

**Gegeben:** Eine Aufgabe wird einmal mit Eino und einmal mit Codex CLI ausgeführt.

**Wenn:** Die Bedienperson beide Läufe und deren strukturierte Ereignisse über die öffentliche Anwendung öffnet.

**Dann:** Jedes Ereignis hat Level, Ereignisname, Zeit sowie die jeweils zugehörigen Organisations-, Projekt-, Task-, Agenten-, Session-, Run- und Attempt-Kennungen. Der tatsächlich verwendete Provider und das Modell sind beim Eino-Lauf sichtbar; beim CLI-Lauf werden die vom Adapter belegten Werte beziehungsweise ein ausdrücklich unbekannter Wert gezeigt. Schlüssel, Tokens, vollständige Prompts und Chain-of-Thought erscheinen weder in Ansicht noch Export oder Log.

## Heartbeat ohne erfundenen Fortschritt

**Gegeben:** Ein Agent arbeitet noch an einer Aufgabe; die letzte belegte Phase ist „Toolantwort verarbeitet“.

**Wenn:** Der nächste regelmäßige Lauf-Heartbeat erscheint, ohne dass der Adapter neue Fortschrittsdaten geliefert hat.

**Dann:** Zeit und letzte belegte Phase bleiben sichtbar; Prozent-, Token- und Modellfortschritt werden nicht aus der verstrichenen Zeit erfunden. Bleibt der Heartbeat aus, wird der Stand als veraltet und nicht als abgeschlossen angezeigt. Ein planender Scheduler-Wake wird nicht mit einem Lauf-Heartbeat verwechselt.

## Usage mit fehlender Reasoning-Zahl

**Gegeben:** Der Provider meldet Input- und Output-Token, aber keine Reasoning-Tokenzahl.

**Wenn:** Die Bedienperson Usage an Task, Run und Attempt öffnet.

**Dann:** Input und Output zeigen die gemeldeten Werte; Reasoning bleibt „unbekannt“ statt null. Rohbefund, Provider, Modell, Zeitpunkt, Zählsemantik und Vollständigkeitsstatus sind nachvollziehbar. Eine vom Provider als Teil der Output-Token gemeldete Reasoning-Zahl wird nicht ein zweites Mal zum Gesamtwert addiert. Es erscheinen keine Kosten- oder Budgetbehauptungen.

## Stream, Finale und Neustart

**Gegeben:** Ein Provider sendet Usage während des Streams und später im Finalereignis; der Server startet vor erneuter Zustellung neu.

**Wenn:** Die Finale und ein wiederholter Streamabschnitt dem gleichen Attempt erneut zugestellt werden.

**Dann:** Gespeicherte Rohmeldungen und normalisierte Usage bleiben demselben Task/Run/Attempt zugeordnet; dieselbe Meldung wird nicht doppelt gezählt. Eine unvollständige oder geschätzte Zahl ist entsprechend gekennzeichnet und wird nie als exakte Null dargestellt.

## Abbruch, Retry und Delegation

**Gegeben:** Ein Attempt wurde nach einer gemeldeten Teil-Usage abgebrochen; ein Retry erhält einen neuen Attempt, und eine delegierte Teilaufgabe hat einen eigenen Task.

**Wenn:** Die Bedienperson den Verlauf und die Usage aller Versuche öffnet.

**Dann:** Der gemeldete Verbrauch des abgebrochenen Attempts bleibt erhalten, sein nicht gemeldeter Rest bleibt unbekannt. Retry und delegierte Teilaufgabe behalten eigene Kennungen und Rohbefunde; weder erneute Zustellung noch Rückaggregation verdoppeln Tokenzahlen. Zusammenfassungen erklären, welche Providerzahlen enthalten sind und welche fehlen.
