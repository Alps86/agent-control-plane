# language: de
@story-04 @arc-04
Funktionalität: Einen Aufgabenlauf einmalig und dauerhaft reservieren
  Als berechtigter Betreiber
  möchte ich einen Lauf für eine Aufgabe atomar reservieren und seinen Zustand wiederfinden,
  damit parallele Startanforderungen keinen doppelten Lauf vorbereiten.

  Grundlage:
    Angenommen ein neuer lokaler Datenbestand ist für die Lauf-Anwendung konfiguriert
    Und die Lauf-Anwendung kennt die Aufgabe "aufgabe-eins" in Organisation "org-eigen"
    Und die Rechtegrenze erlaubt dem Betreiber "betreiber-eigen" den Zugriff auf "aufgabe-eins"

  Szenario: Zwei gleichzeitige Startanforderungen reservieren dieselbe Aufgabe höchstens einmal
    Wenn der Betreiber "betreiber-eigen" über die öffentliche Lauf-Anwendungsgrenze gleichzeitig zweimal einen Start für "aufgabe-eins" anfordert
    Dann wird genau eine der beiden Startanforderungen als Reservierung angenommen
    Und die andere Startanforderung meldet einen bereits reservierten Lauf
    Und die öffentliche Laufansicht zeigt genau einen aktiven Lauf für "aufgabe-eins"
    Und kein Lauf ist als erfolgreich abgeschlossen gekennzeichnet

  Szenario: Die Reservierung bleibt nach Neustart sichtbar und verhindert einen Doppelstart
    Wenn der Betreiber "betreiber-eigen" über die öffentliche Lauf-Anwendungsgrenze einen Start für "aufgabe-eins" anfordert
    Dann erhält er eine Reservierung mit einer Laufkennung und dem Zustand "Reserviert"
    Wenn ich die Lauf-Anwendung mit demselben lokalen Datenbestand neu starte
    Dann sieht der Betreiber "betreiber-eigen" dieselbe Laufkennung im Zustand "Reserviert"
    Wenn der Betreiber "betreiber-eigen" erneut einen Start für "aufgabe-eins" anfordert
    Dann meldet die Lauf-Anwendung den bereits reservierten Lauf
    Und die öffentliche Laufansicht zeigt weiterhin genau einen aktiven Lauf für "aufgabe-eins"

  Szenario: Ein Vorbereitungsfehler bleibt nach Neustart sichtbar
    Angenommen der Betreiber "betreiber-eigen" hat einen Lauf für "aufgabe-eins" reserviert
    Wenn die Lauf-Anwendung vor einem Adapterstart einen Vorbereitungsfehler für diesen Lauf meldet
    Dann zeigt die öffentliche Laufansicht denselben Lauf im Zustand "Fehlgeschlagen" mit dem Vorbereitungsfehler
    Wenn ich die Lauf-Anwendung mit demselben lokalen Datenbestand neu starte
    Dann zeigt die öffentliche Laufansicht denselben Lauf im Zustand "Fehlgeschlagen" mit dem Vorbereitungsfehler
    Und kein Lauf ist als erfolgreich abgeschlossen gekennzeichnet

  Szenario: Ein abgewiesener Zustandswechsel meldet keinen Scheinerfolg
    Angenommen der Betreiber "betreiber-eigen" hat einen Lauf für "aufgabe-eins" reserviert
    Wenn die Lauf-Anwendung diesen Lauf ohne tatsächlich gestarteten Adapter als "Erfolgreich" markieren soll
    Dann weist die Lauf-Anwendung den Zustandswechsel zurück
    Und die öffentliche Laufansicht zeigt den Lauf weiterhin im Zustand "Reserviert"
    Wenn ich die Lauf-Anwendung mit demselben lokalen Datenbestand neu starte
    Dann zeigt die öffentliche Laufansicht den Lauf weiterhin im Zustand "Reserviert"

  Szenario: Eine organisationsfremde Startanforderung legt keinen Lauf an
    Angenommen die Rechtegrenze verweigert dem Betreiber "betreiber-fremd" den Zugriff auf "aufgabe-eins"
    Wenn der Betreiber "betreiber-fremd" über die öffentliche Lauf-Anwendungsgrenze einen Start für "aufgabe-eins" anfordert
    Dann verweigert die Lauf-Anwendung den Start ohne Aufgabendaten preiszugeben
    Und die öffentliche Laufansicht für "betreiber-eigen" zeigt keinen Lauf für "aufgabe-eins"
