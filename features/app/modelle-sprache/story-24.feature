# language: de
@story-24 @ms-04
Funktionalität: Zentrale OpenRouter-Verbindung für Organisation und Eino-Agent freigeben
  Als Betreiber
  möchte ich die Nutzung der zentralen Verbindung ausdrücklich je Organisation und Eino-Agent erlauben,
  damit ein vorhandener Schlüssel allein keinem Agenten Modellzugriff gewährt.
  Ein kontrollierter Modellzugang beobachtet Freigabeentscheidungen ohne echten OpenRouter-Schlüssel oder kostenpflichtigen Aufruf.

  Grundlage:
    Angenommen eine zentrale OpenRouter-Verbindung mit synthetischem Testschlüssel eingerichtet ist
    Und die Organisationen "Nord" und "Süd" über die öffentliche Organisationsgrenze angelegt sind
    Und die Eino-Agenten "Mira" und "Nora" in "Nord" sowie "Sami" in "Süd" über die öffentliche Agentengrenze angelegt sind
    Und ein kontrollierter OpenRouter-Modellzugang bereitsteht

  Szenario: Erst Organisations- und Agentenfreigabe erlauben die Verbindung
    Wenn ich als Betreiber die zentrale Verbindung für "Nord" freigebe
    Dann zeigt die öffentliche Freigabeansicht "Nord" als freigegeben und "Mira" als nicht freigegeben
    Und ein Modellzugangsversuch für "Mira" wird vor Kontakt zum kontrollierten Modellzugang wegen fehlender Agentenfreigabe abgewiesen
    Wenn ich als Betreiber die Verbindung zusätzlich für "Mira" freigebe
    Dann ist die Verbindung für "Mira" zulässig
    Und ein Modellzugangsversuch für "Mira" erreicht den kontrollierten OpenRouter-Modellzugang genau einmal
    Und Modellzugangsversuche für "Nora" und "Sami" erreichen ihn nicht

  Szenariogrundriss: Fehlende Freigabe sperrt bereits die Zuordnung
    Angenommen die Verbindung ist für "Nord" auf "<Ebene>" freigegeben
    Wenn ich die zentrale Verbindung dem Eino-Agenten "Mira" zuordnen möchte
    Dann wird die Zuordnung mit dem Grund "<Grund>" abgelehnt
    Und ein Modellzugangsversuch für "Mira" erreicht den kontrollierten Modellzugang nicht

    Beispiele:
      | Ebene                 | Grund                           |
      | keiner                | Organisationsfreigabe fehlt     |
      | Organisation allein  | Agentenfreigabe fehlt          |

  Szenario: Fremder Agent und fremde Organisation erhalten keine Freigabe durch Referenzkenntnis
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Wenn ich die Verbindungsreferenz für "Sami" aus "Süd" oder "Nora" aus "Nord" bei der öffentlichen Modellzugangsgrenze verwende
    Dann werden beide Versuche vor Kontakt zum kontrollierten Modellzugang abgewiesen
    Und die Antwort enthält keine Schlüssel- oder fremden Organisationsdaten

  Szenario: Agentenfreigabe entziehen stoppt einen neuen Aufruf
    Angenommen die Verbindung ist für "Nord", "Mira" und "Nora" ausdrücklich freigegeben
    Wenn ich als Betreiber die Freigabe für "Mira" widerrufe
    Dann zeigt die öffentliche Freigabeansicht "Mira" als nicht freigegeben
    Und ein neuer Modellzugangsversuch für "Mira" wird vor Kontakt zum kontrollierten Modellzugang abgewiesen
    Und ein Modellzugangsversuch für "Nora" erreicht ihn weiterhin

  Szenario: Organisationsfreigabe entziehen sperrt ihre Agenten
    Angenommen die Verbindung ist für "Nord", "Mira" und "Nora" ausdrücklich freigegeben
    Wenn ich als Betreiber die Freigabe für "Nord" widerrufe
    Dann werden neue Modellzugangsversuche für "Mira" und "Nora" vor Kontakt zum kontrollierten Modellzugang abgewiesen
    Und die öffentliche Freigabeansicht nennt die fehlende Organisationsfreigabe

  Szenariogrundriss: Entzug vor einem noch nicht ausgeführten Aufruf wird erneut geprüft
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Und ein Modellzugangsversuch für "Mira" wartet vor dem tatsächlichen Anbieteraufruf
    Wenn ich als Betreiber die Freigabe für "<Ebene>" widerrufe
    Und der wartende Modellzugangsversuch fortgesetzt wird
    Dann wird der Aufruf mit dem Grund "<Grund>" abgewiesen
    Und der kontrollierte Modellzugang hat keinen Aufruf erhalten

    Beispiele:
      | Ebene | Grund                       |
      | Mira  | Agentenfreigabe fehlt       |
      | Nord  | Organisationsfreigabe fehlt |

  Szenariogrundriss: Widerruf nach erster Autorisierung sperrt die finale Providergrenze
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Und die öffentliche Aufrufgrenze hat einen Modellzugangsversuch für "Mira" positiv autorisiert
    Und der aufrufende Teil hält diesen autorisierten Aufruf vor der finalen serverseitigen Schlüsselauflösung und Providergrenze zurück
    Wenn ich als Betreiber die Freigabe für "<Ebene>" widerrufe
    Und der aufrufende Teil den zurückgehaltenen Aufruf fortsetzt
    Dann wird die Schlüsselauflösung mit dem Grund "<Grund>" verweigert
    Und der kontrollierte Modellzugang hat keine synthetische Provider-POST-Anfrage aus diesem Aufruf erhalten

    Beispiele:
      | Ebene | Grund                       |
      | Mira  | Agentenfreigabe fehlt       |
      | Nord  | Organisationsfreigabe fehlt |

  Szenario: Neue Organisation und neuer Agent erben keine Freigabe
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Wenn ich eine weitere Organisation und darin einen Eino-Agenten über die öffentlichen Grenzen anlege
    Und ich einen weiteren Eino-Agenten in "Nord" über die öffentliche Agentengrenze anlege
    Dann haben die neue Organisation und beide neuen Agenten keine OpenRouter-Freigabe
    Und ihre Modellzugangsversuche erreichen den kontrollierten Modellzugang nicht

  Szenario: Delegierter Auftrag erbt keine Verbindung
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Wenn ein Modellzugangsversuch mit "Mira" als delegierendem und "Nora" als ausführendem Eino-Agenten ankommt
    Dann wird der Versuch wegen fehlender Freigabe für "Nora" vor Kontakt zum kontrollierten Modellzugang abgewiesen
    Und die Freigabe für "Mira" wird nicht auf "Nora" übertragen

  Szenario: Schlüsseltausch erweitert keine Freigabe
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Wenn ich als Betreiber den zentralen Testschlüssel über Settings ersetze
    Dann bleiben nur "Nord" und "Mira" freigegeben
    Und Modellzugangsversuche für "Nora" und "Sami" erreichen den kontrollierten Modellzugang nicht

  Szenario: Trennen und erneutes Verbinden belebt alte Freigaben nicht wieder
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Und ein Modellzugangsversuch für "Mira" erreicht den kontrollierten Modellzugang
    Wenn ich als Betreiber die zentrale Verbindung über Settings trenne
    Und ich sie mit einem neuen synthetischen Testschlüssel erneut verbinde
    Dann zeigt Settings dieselbe zentrale Verbindungsreferenz ohne Schlüssel
    Und die früheren Freigaben für "Nord" und "Mira" sind nicht mehr wirksam
    Und ein Modellzugangsversuch für "Mira" erreicht den kontrollierten Modellzugang nicht
    Wenn ich als Betreiber zuerst "Nord" und danach "Mira" für die erneuerte Verbindung ausdrücklich freigebe
    Dann erreicht ein neuer Modellzugangsversuch für "Mira" den kontrollierten Modellzugang genau einmal
    Und öffentliche Settings-, Freigabe- und Modellzugangsantworten enthalten keinen der Testschlüssel

  Szenario: Alte autorisierte Bindung wird nach Trennen und Neuverbinden vor Providerkontakt abgewiesen
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Und die öffentliche Aufrufgrenze hat einen Modellzugangsversuch für "Mira" positiv autorisiert
    Und der aufrufende Teil hält diese autorisierte Bindung vor dem tatsächlichen Anbieteraufruf zurück
    Wenn ich als Betreiber die zentrale Verbindung über Settings trenne
    Und ich sie unter derselben Verbindungsreferenz mit einem neuen synthetischen Testschlüssel erneut verbinde
    Und der aufrufende Teil die zurückgehaltene Bindung am serverseitigen Verbindungsresolver einlöst
    Dann weist der Resolver die alte Bindung als widerrufen ab
    Und der kontrollierte Modellzugang hat keinen Aufruf aus dieser Bindung erhalten
    Und die früheren Freigaben für "Nord" und "Mira" sind nicht wieder wirksam

  Szenario: Öffentliche Freigabeantwort enthält nur Referenz und Status
    Angenommen die Verbindung ist für "Nord" und "Mira" ausdrücklich freigegeben
    Wenn ich die öffentlichen Freigabeansichten und JSON-Antworten für Organisation und Agent abrufe
    Dann zeigen sie die Verbindungsreferenz und den jeweiligen Freigabestatus
    Und sie enthalten weder den synthetischen Testschlüssel noch einen entschlüsselten Teil davon
    Und der Testschlüssel erscheint nicht in Agentenkonfiguration oder Modellzugangsfehlern
