# language: de
@story-23 @ms-03
Funktionalität: Eine zentrale OpenRouter-Verbindung sicher verwalten
  Als Betreiber
  möchte ich genau eine OpenRouter-Verbindung für diese Installation verwalten,
  damit OpenRouter nur nach bewusster Einrichtung und späterer Freigabe nutzbar wird.

  Grundlage:
    Angenommen der lokale Server verwendet einen neuen geschützten Credentials-Speicher
    Und ein kontrollierter OpenRouter-Testanbieter steht bereit

  Szenario: Eine zentrale Verbindung mit gültigem Schlüssel speichern und prüfen
    Angenommen noch keine OpenRouter-Verbindung eingerichtet ist
    Wenn ich als Betreiber einen gültigen OpenRouter-Schlüssel über die öffentliche Settings-Grenze speichere
    Dann zeigt die Settings-Antwort genau eine zentrale OpenRouter-Verbindungsreferenz
    Und der gespeicherte Schlüssel ist in jeder öffentlichen Antwort maskiert
    Wenn ich den Verbindungsstatus ausdrücklich prüfe
    Dann meldet die Verbindung mit dem kontrollierten Testanbieter "einsatzbereit"
    Und die Prüfung verwendet nur den eingerichteten OpenRouter-Anbieter

  Szenario: Erneutes Speichern ersetzt den Schlüssel derselben Verbindung
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Wenn ich als Betreiber einen anderen gültigen OpenRouter-Schlüssel speichere
    Dann zeigt die Settings-Antwort weiterhin genau dieselbe Verbindungsreferenz
    Wenn ich den Verbindungsstatus ausdrücklich prüfe
    Dann erreicht nur der neue Schlüssel den kontrollierten Testanbieter
    Und der alte Schlüssel wird nicht erneut verwendet

  Szenario: Ein fehlender Schlüssel richtet keine Verbindung ein
    Angenommen noch keine OpenRouter-Verbindung eingerichtet ist
    Wenn ich als Betreiber einen leeren OpenRouter-Schlüssel über die öffentliche Settings-Grenze speichere
    Dann wird die Eingabe mit dem Hinweis "Schlüssel fehlt" abgelehnt
    Und die OpenRouter-Verbindung ist nicht einsatzbereit
    Und der kontrollierte Testanbieter erhält keine Anfrage

  Szenario: Ein ungültiger Schlüssel wird nicht als einsatzbereit bestätigt
    Angenommen noch keine OpenRouter-Verbindung eingerichtet ist
    Wenn ich als Betreiber einen ungültigen OpenRouter-Schlüssel speichere
    Und ich den Verbindungsstatus ausdrücklich prüfe
    Dann meldet die Verbindung "nicht einsatzbereit" mit einem verständlichen Authentifizierungsgrund
    Und es erfolgt keine Anfrage an einen anderen kostenpflichtigen Anbieter

  Szenario: Prüfung einer nicht eingerichteten Verbindung vermeidet Anbieterzugriff
    Angenommen noch keine OpenRouter-Verbindung eingerichtet ist
    Wenn ich den Verbindungsstatus ausdrücklich prüfe
    Dann meldet die Verbindung "nicht eingerichtet"
    Und der kontrollierte Testanbieter erhält keine Anfrage

  Szenario: Verbindung trennen und gespeicherten Schlüssel entfernen
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Wenn ich als Betreiber die OpenRouter-Verbindung über die öffentliche Settings-Grenze trenne
    Dann zeigt die Settings-Antwort "Verbindung getrennt"
    Und die OpenRouter-Verbindung ist nicht einsatzbereit
    Wenn ich den Verbindungsstatus ausdrücklich prüfe
    Dann erhält der kontrollierte Testanbieter keine weitere Anfrage

  Szenario: Verbindung und Status über einen Neustart erhalten
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Und ich den Verbindungsstatus erfolgreich geprüft habe
    Wenn ich den lokalen Server mit demselben geschützten Credentials-Speicher neu starte
    Dann zeigt die Settings-Antwort dieselbe zentrale OpenRouter-Verbindungsreferenz
    Und die Settings-Antwort enthält keinen Klartextschlüssel
    Wenn ich den Verbindungsstatus ausdrücklich prüfe
    Dann meldet die Verbindung mit dem kontrollierten Testanbieter "einsatzbereit"

  Szenario: Ein gespeicherter Schlüssel erscheint in keiner öffentlichen Antwort
    Angenommen eine zentrale OpenRouter-Verbindung mit eindeutigem Testschlüssel eingerichtet ist
    Wenn ich die öffentliche Settings-Ansicht und ihre JSON-Antwort abrufe
    Dann enthalten HTML und JSON den eindeutigen Testschlüssel nicht
    Und HTML und JSON enthalten keine teilweise entschlüsselten Schlüsselwerte
    Und ich sehe nur die zentrale Verbindungsreferenz und nichtgeheime Statusdaten

  Szenario: Die zentrale Einrichtung zeigt keine Nutzungsfreigaben an
    Angenommen noch keine OpenRouter-Verbindung eingerichtet ist
    Wenn ich als Betreiber eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel einrichte
    Dann enthält die öffentliche Settings-Antwort keine Organisations- oder Agentenfreigabe
    Und sie zeigt nur die zentrale Verbindungsreferenz ohne Schlüssel

  Szenario: Das Speichern löst keine kostenpflichtige Anfrage aus
    Angenommen noch keine OpenRouter-Verbindung eingerichtet ist
    Wenn ich als Betreiber einen gültigen OpenRouter-Schlüssel über die öffentliche Settings-Grenze speichere
    Dann erhält der kontrollierte Testanbieter noch keine Anfrage
    Und erst meine ausdrückliche Statusprüfung darf den Anbieter kontaktieren

  Szenariogrundriss: Fremde oder fehlende Herkunft darf die Verbindung nicht ändern
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Wenn ich einen anderen Schlüssel mit "<Herkunft>" an die öffentliche Settings-Grenze sende
    Dann wird die Änderung mit HTTP 403 verweigert
    Und der kontrollierte Testanbieter erhält durch die verweigerte Änderung keine Anfrage
    Und die zentrale Verbindung verwendet weiterhin den bisherigen Schlüssel

    Beispiele:
      | Herkunft       |
      | fremdem Origin |
      | ohne Origin    |

  Szenario: Fremder Host darf eine Statusprüfung nicht auslösen
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Wenn ich eine Statusprüfung mit fremdem Host an die öffentliche Settings-Grenze sende
    Dann wird die Prüfung mit HTTP 403 verweigert
    Und der kontrollierte Testanbieter erhält keine Anfrage

  Szenario: Späte Statusprüfung darf einen ersetzten Schlüssel nicht wiederherstellen
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Und der kontrollierte Testanbieter hält die nächste Statusprüfung zurück
    Wenn ich die Statusprüfung über die öffentliche Settings-Grenze nebenläufig starte
    Und der Testanbieter den bisherigen Schlüssel empfängt und die Antwort zurückhält
    Und ich das Ersetzen des Schlüssels über die öffentliche Settings-Grenze nebenläufig starte
    Und der Schlüsseltausch bleibt an der öffentlichen HTTP-Grenze bis zur Anbieterantwort offen
    Und ich die zurückgehaltene Anbieterantwort freigebe
    Dann schließen Statusprüfung und Schlüsseltausch ohne Fehler ab
    Und die öffentliche Settings-Antwort zeigt dieselbe Verbindung mit Status "nicht geprüft"
    Wenn ich den Verbindungsstatus erneut ausdrücklich prüfe
    Dann erreicht ausschließlich der neue Schlüssel den Testanbieter

  Szenario: Späte Statusprüfung darf eine getrennte Verbindung nicht wiederherstellen
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Und der kontrollierte Testanbieter hält die nächste Statusprüfung zurück
    Wenn ich die Statusprüfung über die öffentliche Settings-Grenze nebenläufig starte
    Und der Testanbieter den bisherigen Schlüssel empfängt und die Antwort zurückhält
    Und ich das Trennen der Verbindung über die öffentliche Settings-Grenze nebenläufig starte
    Und die Trennung bleibt an der öffentlichen HTTP-Grenze bis zur Anbieterantwort offen
    Und ich die zurückgehaltene Anbieterantwort freigebe
    Dann schließen Statusprüfung und Trennung ohne Fehler ab
    Und die öffentliche Settings-Antwort zeigt "nicht eingerichtet"
    Wenn ich den Verbindungsstatus erneut ausdrücklich prüfe
    Dann erhält der kontrollierte Testanbieter nach der Trennung keine zusätzliche Anfrage

  Szenariogrundriss: Fremder TCP-Client erhält trotz gefälschtem lokalem Host keine Settings-Daten
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Wenn ein fremder TCP-Client "GET" an "<Pfad>" mit gefälschtem lokalem Host und passendem Origin sendet
    Dann antwortet die öffentliche Grenze mit HTTP 403 ohne Schlüssel oder Verbindungsstatus
    Und der kontrollierte Testanbieter erhält keine Anfrage
    Und die Verbindung bleibt über den lokalen Settings-Zugriff unverändert

    Beispiele:
      | Pfad |
      | /api/settings/modellanbieter/openrouter |
      | /settings/modellanbieter/openrouter |

  Szenariogrundriss: Fremder TCP-Client löst trotz gefälschtem lokalem Host keine Mutation aus
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Wenn ein fremder TCP-Client "<Methode>" an "<Pfad>" mit gefälschtem lokalem Host und passendem Origin sendet
    Dann antwortet die öffentliche Grenze mit HTTP 403 ohne Schlüssel oder Verbindungsstatus
    Und der kontrollierte Testanbieter erhält keine Anfrage
    Und die Verbindung bleibt über den lokalen Settings-Zugriff unverändert

    Beispiele:
      | Methode | Pfad |
      | POST | /api/settings/modellanbieter/openrouter/pruefen |
      | POST | /api/settings/modellanbieter/openrouter |
      | DELETE | /api/settings/modellanbieter/openrouter |

  Szenariogrundriss: Unsichere Server-Bindadresse sperrt Settings bereits an der HTTP-Grenze
    Angenommen eine zentrale OpenRouter-Verbindung mit gültigem Schlüssel eingerichtet ist
    Wenn ich einen lokalen Settings-Server mit Bindadresse "<Bindadresse>" über "<Methode>" an "<Pfad>" aufrufe
    Dann antwortet die öffentliche Grenze mit HTTP 403 ohne Schlüssel oder Verbindungsstatus
    Und der kontrollierte Testanbieter erhält keine Anfrage
    Und die Verbindung bleibt über den lokalen Settings-Zugriff unverändert

    Beispiele:
      | Bindadresse | Methode | Pfad |
      | 0.0.0.0:8080 | GET | /api/settings/modellanbieter/openrouter |
      | 0.0.0.0:8080 | POST | /api/settings/modellanbieter/openrouter |
      | [::]:8080 | GET | /api/settings/modellanbieter/openrouter |
      | [::]:8080 | POST | /api/settings/modellanbieter/openrouter/pruefen |
      | localhost:8080 | GET | /api/settings/modellanbieter/openrouter |
      | localhost:8080 | POST | /api/settings/modellanbieter/openrouter |
      | ungültig | GET | /api/settings/modellanbieter/openrouter |
      | ungültig | POST | /api/settings/modellanbieter/openrouter/pruefen |
