# Organisation und team

Quelle: `spec/features/zusammenarbeit/organisation-und-team.feature`

Planungsbeispiel; keine Aussage über eine implementierte oder bestandene Prüfung. Der folgende Gherkin-Text ist vollständig aus der Quelle übernommen.

```gherkin
# language: de
@ZA-02
Funktionalität: Arbeit im Organisations- und Teamkontext steuern
  Als Betreiberin möchte ich sehen, welche Agenten gemeinsam arbeiten und wofür sie zuständig sind.

  Szenario: Ein Team innerhalb einer Organisation zusammenstellen
    Angenommen "Mira" und "Kai" gehören zu "Nordstern"
    Wenn ich das Team "Web" mit "Mira" und "Kai" anlege
    Dann sehe ich "Web" mit seinen Mitgliedern unter "Nordstern"
    Und ich sehe "Web" als ansprechbares Team der Organisation

  Szenario: Zuständigkeit und Arbeitsstand im Team erkennen
    Angenommen das Team "Web" bearbeitet die Aufgabe "Startseite freigeben"
    Und "Mira" ist für "Texte prüfen" zuständig
    Wenn ich die Teamübersicht öffne
    Dann sehe ich die Aufgabe und die Zuständigkeit von "Mira"
    Und ich kann von dort die zugehörige Aufgabe öffnen

  Szenario: Organisationsübersicht zeigt Arbeit über Teams hinweg nur einmal
    Angenommen "Startseite freigeben" gehört zu "Nordstern" und wird von "Web" und "Qualität" begleitet
    Wenn ich die Arbeitsübersicht von "Nordstern" öffne
    Dann sehe ich "Startseite freigeben" genau einmal
    Und ich sehe die beteiligten Teams und den aktuellen Aufgabenstatus

  Szenario: Priorität, Blockade und unzugewiesene Arbeit nach Neustart sehen
    Angenommen "Startseite freigeben" hat hohe Priorität und ist durch "Texte prüfen" blockiert
    Und "Bilder prüfen" in "Nordstern" ist noch keinem Agenten zugewiesen
    Wenn ich die Anwendung neu starte und die Arbeitsübersicht von "Nordstern" öffne
    Dann sehe ich Priorität und offene Blockade von "Startseite freigeben"
    Und ich sehe "Bilder prüfen" als unzugewiesene Arbeit

  Szenario: Organisationsgrenzen bei der Teamzuordnung einhalten
    Angenommen das Team "Web" gehört zu "Nordstern"
    Und "Lena" gehört ausschließlich zu "Südstern"
    Wenn ich "Lena" zu "Web" hinzufügen will
    Dann wird die Zuordnung abgelehnt
    Und "Lena" erhält keinen Zugriff auf die Aufgaben von "Web"
```
