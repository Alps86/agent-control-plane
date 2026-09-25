# language: de
@story-42 @task-05
Funktionalität: Kommentare im organisationsgebundenen Aufgabenthread speichern
  Als Betreiberin und berechtigter Agent
  möchte ich eine Aufgabe kommentieren,
  damit Aussage, verlässliche Herkunft, Zeitpunkt und Bezüge dauerhaft nachvollziehbar sind.

  Hintergrund:
    Angenommen ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin
    Und die Organisation "Nordstern" mit dem Projekt "Website" und der Aufgabe "Texte prüfen" besteht
    Und die Organisation "Suedstern" mit dem Projekt "Fremd" und der Aufgabe "Fremdaufgabe" besteht
    Und "Mira" ist ein Agent von "Nordstern" mit Lese- und Schreibfreigabe für "Website"
    Und "OhneRecht" ist ein Agent von "Nordstern" ohne Schreibfreigabe für "Website"
    Und "Fremd" ist ein Agent von "Suedstern"

  Szenario: Die Betreiberin ergänzt den Thread und findet ihn nach Neustart wieder
    Wenn ich als lokale Betreiberin "Fachprüfung folgt" über die öffentliche Aufgaben-API an "Texte prüfen" kommentiere
    Dann liefert die API eine stabile Kommentarkennung und die Aufgabenkennung von "Texte prüfen"
    Und der Aufgabenthread zeigt "Fachprüfung folgt" mit Quelle "Betreiberin" und einem serverseitigen Zeitpunkt
    Wenn ich den Server beende und mit derselben SQLite-Datenbank erneut starte
    Dann zeigt derselbe Aufgabenthread denselben Kommentar mit Kennung, Quelle, Zeitpunkt und Aufgabenkennung genau einmal

  Szenario: Ein berechtigter Agent kommentiert mit serverseitig bestimmter Identität
    Wenn "Mira" über die öffentliche Agenten-Fassade "Entwurf ist bereit" an "Texte prüfen" kommentiert
    Dann zeigt der Aufgabenthread "Entwurf ist bereit" mit Quelle "Mira" und einem serverseitigen Zeitpunkt
    Und die Kommentarantwort enthält die stabile Kennung von "Texte prüfen"
    Wenn ein Client über die öffentliche Aufgaben-API eine Betreiberkennung als Autor für einen neuen Kommentar mitsendet
    Dann wird die Autorenangabe abgewiesen
    Und es entsteht kein Kommentar mit gefälschter Quelle

  Szenariogrundriss: Leerer Kommentar wird ohne Threadänderung abgewiesen
    Wenn ich als lokale Betreiberin <Inhalt> an "Texte prüfen" kommentiere
    Dann erhalte ich einen Feldfehler für den Kommentarinhalt
    Und im Aufgabenthread entsteht kein neuer Kommentar

    Beispiele:
      | Inhalt |
      | ""     |
      | "   "  |

  Szenariogrundriss: Fehlende, fremde oder nicht freigegebene Aufgabe verrät keine fremden Daten
    Wenn <Akteur> über die öffentliche Kommentar-Fassade einen Kommentar an <Aufgabe> senden möchte
    Dann wird der Schreibzugriff ohne Kommentar abgewiesen
    Und die Antwort enthält weder Inhalt noch Kennung einer fremden Aufgabe
    Und der Aufgabenthread von <Aufgabe> bleibt unverändert

    Beispiele:
      | Akteur       | Aufgabe                         |
      | "Mira"       | "Fremdaufgabe" aus "Suedstern" |
      | "OhneRecht"  | "Texte prüfen"                 |
      | "Fremd"      | "Texte prüfen"                 |
      | "Mira"       | eine unbekannte Aufgabenkennung |

  Szenario: Ein Agent kann einen fremden Thread nicht lesen
    Wenn "Fremd" über die öffentliche Agenten-Fassade den Thread von "Texte prüfen" abruft
    Dann wird der Lesezugriff ohne Kommentarinhalt und ohne Aufgabenangaben abgewiesen

  Szenario: Entzogene Schreibfreigabe stoppt auch die Änderung eines eigenen Kommentars
    Angenommen "Mira" hat "Erster Entwurf" an "Texte prüfen" kommentiert
    Wenn ich "Mira" die Schreibfreigabe für "Website" entziehe
    Und "Mira" ihren Kommentar über die öffentliche Agenten-Fassade in "Geänderter Entwurf" ändern möchte
    Dann wird die Änderung verweigert
    Und der Thread zeigt weiterhin "Erster Entwurf" mit ursprünglicher Quelle und Zeitpunkt

  Szenario: Ein Agent kann den Kommentar der Betreiberin nicht ändern
    Angenommen ich habe "Fachprüfung folgt" an "Texte prüfen" kommentiert
    Wenn "Mira" diesen Kommentar über die öffentliche Agenten-Fassade ändern möchte
    Dann wird die Änderung ohne Überschreiben verweigert
    Und der Thread zeigt weiterhin "Fachprüfung folgt" mit Quelle "Betreiberin"

  Szenario: Ein verifizierter Aufgabenbezug bleibt am Kommentar erhalten
    Angenommen die weitere Aufgabe "Freigabe" besteht im Projekt "Website" von "Nordstern"
    Wenn ich als lokale Betreiberin "Bitte Freigabe prüfen" mit einem Aufgabenbezug auf "Freigabe" an "Texte prüfen" kommentiere
    Dann zeigt der Thread Kommentar, Quellaufgabenkennung und die stabile Kennung von "Freigabe" als verifizierten Aufgabenlink
    Wenn ich den Server mit derselben SQLite-Datenbank neu starte
    Dann bleibt der Aufgabenlink unverändert am selben Kommentar

  Szenario: Ein fremder Aufgabenbezug wird atomar abgewiesen
    Wenn ich als lokale Betreiberin "Bitte prüfen" mit einem Aufgabenbezug auf "Fremdaufgabe" an "Texte prüfen" kommentiere
    Dann wird der gesamte Kommentar ohne fremde Aufgabendaten abgewiesen
    Und der Aufgabenthread von "Texte prüfen" bleibt unverändert

  Szenario: Eine typisierte Artefaktreferenz bleibt ohne Artefakterzeugung erhalten
    Angenommen der vertrauenswürdige Metadatenresolver kennt "entwurf-7" als Artefakt "Textentwurf" von "Nordstern"
    Wenn ich als lokale Betreiberin "Siehe Entwurf" mit der Artefaktkennung "entwurf-7" an "Texte prüfen" kommentiere
    Dann zeigt der Thread die aufgelöste Artefaktkennung "entwurf-7", den Anzeigenamen "Textentwurf" und einen internen Link am Kommentar
    Wenn ich den Server mit derselben SQLite-Datenbank neu starte
    Dann bleibt diese typisierte Referenz unverändert am selben Kommentar
    Und der Kommentaraufruf hat kein Artefakt angelegt oder seinen Inhalt ausgeliefert

  Szenariogrundriss: Unaufgelöste und fremde Artefaktkennung wird abgewiesen
    Angenommen der vertrauenswürdige Metadatenresolver meldet <Metadaten> für "entwurf-7"
    Wenn ich als lokale Betreiberin "Siehe Entwurf" mit der Artefaktkennung "entwurf-7" an "Texte prüfen" kommentiere
    Dann wird der gesamte Kommentar ohne fremde Artefaktdaten abgewiesen
    Und der Aufgabenthread von "Texte prüfen" bleibt unverändert

    Beispiele:
      | Metadaten                 |
      | keine Metadaten           |
      | ein Artefakt von "Suedstern" |

  Szenario: Ein archiviertes Projekt nimmt keine neuen oder geänderten Kommentare an
    Angenommen "Mira" hat "Erster Entwurf" an "Texte prüfen" kommentiert
    Wenn ich das Projekt "Website" archiviere
    Dann kann ich den bestehenden Kommentar weiter im Thread lesen
    Wenn ich als lokale Betreiberin "Noch eine Prüfung" an "Texte prüfen" kommentiere
    Dann wird die Änderung ohne neuen Kommentar abgewiesen
    Wenn "Mira" ihren Kommentar in "Geänderter Entwurf" ändern möchte
    Dann bleibt "Erster Entwurf" unverändert
    Wenn ich das Projekt "Website" wiederherstelle
    Und ich als lokale Betreiberin "Noch eine Prüfung" an "Texte prüfen" kommentiere
    Dann erscheint der neue Kommentar im Thread
