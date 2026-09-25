# language: de
@story-24 @ms-04 @prozess @probe
Funktionalität: Kontrollierte OpenRouter-Probe am gebauten Server
  Als Betreiber
  möchte ich einen synthetischen lokalen Anbieter prüfen und danach die Verbindung gezielt freigeben,
  damit nur ein belegter Status und ausdrückliche Freigaben Modellzugriff erlauben.

  Szenariogrundriss: Numerischer Loopback-Anbieter wird echt geprüft und gezielt freigegeben
    Angenommen ein lokaler synthetischer OpenRouter-Anbieter hört auf numerischem <IP-Typ>-Loopback und antwortet auf GET "/api/v1/key" mit HTTP 200 und {"data":{}}
    Und APP_OPENROUTER_PROBE_URL enthält seine HTTP-Adresse mit explizitem Port
    Und ein gebauter Server mit isolierter SQLite-, Credential- und Masterkey-Datei ist gestartet
    Wenn ich einen synthetischen Schlüssel über die öffentliche Settings-API speichere und prüfe
    Dann empfängt der lokale Anbieter genau einen GET "/api/v1/key" mit diesem Schlüssel im Bearer-Header und ohne Requestbody
    Und die öffentliche Settings-API meldet "einsatzbereit" und eine Verbindungsreferenz ohne Schlüssel
    Wenn ich über die öffentlichen APIs die Organisation "Nord" und darin den Eino-Agenten "Mira" anlege
    Und ich die zentrale OpenRouter-Verbindung erst für "Nord" und dann für "Mira" öffentlich freigebe
    Dann meldet die öffentliche Freigabe-API beide Freigaben für dieselbe Verbindungsreferenz
    Und Settings-, Freigabe- und Prozessausgaben enthalten den synthetischen Schlüssel nicht

    Beispiele:
      | IP-Typ |
      | IPv4   |
      | IPv6   |

  Szenariogrundriss: Unsichere Probe-Adresse verhindert den Serverstart
    Angenommen APP_OPENROUTER_PROBE_URL ist "<URL>"
    Wenn ich den gebauten Server mit isolierter SQLite-, Credential- und Masterkey-Datei starte
    Dann beendet er den Start mit einem Fehler vor der öffentlichen HTTP-Bereitschaft
    Und es wurde keine Anfrage an einen OpenRouter-Anbieter gesendet
    Und die Prozessausgabe enthält keinen synthetischen Schlüssel

    Beispiele:
      | Fall                    | URL                                |
      | DNS-Name                | http://localhost:32123             |
      | Nichtloopback IPv4      | http://192.0.2.1:32123             |
      | Nichtloopback IPv6      | http://[2001:db8::1]:32123         |
      | fehlender Port          | http://127.0.0.1                   |
      | Userinfo                | http://user@127.0.0.1:32123        |
      | Query                   | http://127.0.0.1:32123?probe=1     |
      | leere Query             | http://127.0.0.1:32123?            |
      | Fragment                | http://127.0.0.1:32123#probe       |
      | leeres Fragment         | http://127.0.0.1:32123#            |
      | Pfad-Wurzel             | http://127.0.0.1:32123/            |
      | Pfad                    | http://127.0.0.1:32123/fremd       |

  Szenariogrundriss: Redirect der Schlüsselprüfung wird ohne Weitergabe des Schlüssels abgelehnt
    Angenommen ein lokaler synthetischer OpenRouter-Anbieter antwortet auf GET "/api/v1/key" mit HTTP <Status> und einem Redirect zu einem zweiten lokalen Empfänger
    Und APP_OPENROUTER_PROBE_URL enthält die numerische IPv4-Loopback-HTTP-Adresse des ersten Anbieters mit explizitem Port
    Und ein gebauter Server mit isolierter SQLite-, Credential- und Masterkey-Datei ist gestartet
    Wenn ich einen synthetischen Schlüssel über die öffentliche Settings-API speichere und prüfe
    Dann empfängt nur der erste Anbieter genau einen GET "/api/v1/key" mit diesem Schlüssel im Bearer-Header
    Und der zweite Empfänger erhält keine Anfrage und keinen Schlüssel
    Und die öffentliche Settings-API meldet nicht "einsatzbereit"
    Und Settings- und Prozessausgaben enthalten den synthetischen Schlüssel nicht

    Beispiele:
      | Status |
      | 302    |
      | 307    |
