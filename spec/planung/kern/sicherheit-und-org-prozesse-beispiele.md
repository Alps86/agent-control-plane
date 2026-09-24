# Öffentliche Akzeptanzbeispiele: Organisationsabläufe und Agentengrenzen

Stand: 24. September 2026. Diese Beispiele sind Planungsgrundlage für ORG-02, AGT-01/02, TASK-04, DEL-01/02, RUN-01/05, ART-01, SKL-01/02, INT-02/03 und PERM-02 bis PERM-05. Sie sind keine fertigen Feature-Dateien oder bestandenen Tests. Künftige ausführbare Fälle liegen unter `features/app/<fachbereich>/` und `features/ui/<fachbereich>/`.

## Organisationsspezifische Arbeitsregel

**Gegeben:** Organisation Nord erlaubt die Delegation eines laufenden Issues an einen freigegebenen Berichtsempfänger; Organisation Süd verlangt dafür zuerst einen prüfbaren Freigabeschritt.

**Wenn:** Die Bedienperson dieselbe Übergabe in beiden Organisationen über die öffentliche Anwendung auslöst.

**Dann:** Nord legt genau die erlaubte Teilaufgabe an. Süd zeigt den noch fehlenden Freigabeschritt und legt keine Teilaufgabe an. Eine Änderung der Regel in Süd ändert den Verlauf alter Entscheidungen nicht. Ein ungültiger Statusübergang oder ein Empfänger ohne Rechte wird auch bei direktem Aufruf abgewiesen. Die Regeln sind konkrete Konfigurationswerte, keine frei ausführbare Workflow-Sprache.

## Kein direkter Agentenzugriff

**Gegeben:** Ein Eino- und ein Codex-CLI-Agent sind ohne direkte Shell-, Terminal-, Prozess-, Interpreter-, HTTP- und Dateisystemfähigkeit konfiguriert.

**Wenn:** Einer der Agenten einen direkten Befehl, eine URL, einen Dateipfad oder Interpretercode als auszuführende Aktion vorschlägt.

**Dann:** Es erfolgt kein direkter Aufruf und keine Seitenwirkung. Die Anwendung zeigt die fehlende benannte Fachfähigkeit oder verweigert den Aufruf. Kann ein Adapter diese Grenze nicht nachweisbar erzwingen, ist dieser Agent für das Profil nicht startbereit. Eine sichtbare UI-Sperre allein gilt nicht als Nachweis.

## Benannte Websuche und Skill-Aktion

**Gegeben:** Einem Eino-Agenten ist die benannte Websuche für einen begrenzten Rechercheauftrag erlaubt; ein Skill referenziert eine vorab registrierte Fachaktion mit ID und validierten Parametern.

**Wenn:** Der Agent die Websuche oder die Fachaktion nutzt.

**Dann:** Der kontrollierte Go-Adapter beziehungsweise Go-Anwendungsfall prüft Organisation, Agent, Aktion und Parameter unmittelbar vor dem Aufruf und zeigt Herkunft und Ergebnis. Ein beliebiger HTTP-Aufruf, Shelltext, Interpretercode oder generisches Exec wird abgewiesen. Ein Skilltext oder SOUL.md allein erteilt keine Freigabe.

## Markdown-Arbeitsprodukt

**Gegeben:** Ein Rechercheagent liefert Markdown-Inhalt für ein Arbeitsprodukt in einer berechtigten Organisationssession.

**Wenn:** Die Anwendung das Ergebnis annimmt.

**Dann:** Die Go-App bestimmt Ziel und Typ, validiert den Inhalt und speichert ihn im begrenzten Organisations-/Session-Artefaktbereich. Der Agent schreibt nicht selbst in das Dateisystem. Ausführbare Skripte oder Dateien werden weder erzeugt noch persistiert; Markdown wird nicht ausgeführt. Ein fremder Organisationspfad wird abgewiesen.

## Dauerhafte Session und Delegation

**Gegeben:** Ein Agentengespräch und sein Arbeitskontext bestehen nach Neustart fort; eine Teilaufgabe soll an einen Kindagenten übergehen.

**Wenn:** Der Auftrag aus der berechtigten Organisation fortgesetzt oder delegiert wird.

**Dann:** Die Session bleibt sichtbar und organisationsgebunden. Der Kindagent erhält nur die Schnittmenge der Rechte von Organisation, Auftraggeber und eigenem Profil. Jede neue Daten-, Tool- und Artefaktaktion wird erneut geprüft; Session, Teamzuordnung, Skill und SOUL.md erweitern keine Rechte.
