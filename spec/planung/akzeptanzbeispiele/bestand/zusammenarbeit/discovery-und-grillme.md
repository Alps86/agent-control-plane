# Discovery und grillme

Quelle: `spec/features/zusammenarbeit/discovery-und-grillme.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@ZA-07
Funktionalität: Später Anforderungen erkunden und kritisch prüfen
  Als Betreiberin möchte ich bei Bedarf aus einer Idee einen klaren Auftrag entwickeln.

  Szenario: Discovery sammelt fehlende Angaben vor dem Arbeitsauftrag
    Angenommen ich beginne im Chat eine Discovery für "Wissensportal verbessern"
    Wenn ich das Ziel "Neue Besucher finden Antworten schneller" beschreibe
    Dann fragt der Agent nach fehlenden Angaben zum gewünschten Ergebnis
    Und ich sehe die bisherigen Antworten im selben Gespräch

  Szenario: GrillMe hinterfragt eine vorgeschlagene Lösung
    Angenommen ich habe in der Discovery "Eine neue Suche bauen" vorgeschlagen
    Wenn ich GrillMe für diesen Vorschlag starte
    Dann stellt der Agent kritische Rückfragen zu Annahmen und Erfolgskriterien
    Und ich kann die Antworten vor einem Arbeitsauftrag überarbeiten

  Szenario: Erkundung führt erst nach Bestätigung zu Arbeit
    Angenommen Discovery und GrillMe haben einen Entwurf für "Wissensportal verbessern" erstellt
    Wenn ich den Entwurf noch nicht als Auftrag bestätige
    Dann wird daraus keine Aufgabe und kein Agentenlauf gestartet
    Wenn ich den Entwurf als Auftrag bestätige
    Dann sehe ich die daraus angelegte Aufgabe mit Verweis auf das Gespräch
```
