# Story-29 / MS-07 – Modellverbindungen dauerhaft verwenden

Stand: 25. September 2026. Dieser Nachweis betrifft kontrollierte lokale Anbieter, geschützte Test-Credentials und die öffentliche HTTP- beziehungsweise App-Service-Grenze. Es wurden keine echten Codex- oder OpenRouter-Modellaufrufe ausgeführt.

## Ausgeführte Prüfungen

- `npm ci --offline && npm run build` in `ui/web/` war erfolgreich. Das vollständige erzeugte `dist/` wurde für den lokalen Server-Build nach `ui/bridge/dist/` kopiert; die UI-Quellen wurden nicht geändert.
- `cd app && GOCACHE=/tmp/story29-gocache go test ./cucumber/modelllebenszyklus -count=1 -run TestFeatures` bestand mit allen neun Szenarien. `go test -race` mit denselben Paket- und Testargumenten bestand ebenfalls.
- Ein gebauter lokaler Serverprozess wurde mit derselben SQLite-Datei sowie derselben verschlüsselten Credential- und Masterkey-Datei beendet und neu gestartet. Die öffentliche Codex-Settings-Antwort zeigte danach die bestehende Verbindung; die getrennt erneut geöffnete serverseitige Zugangsgrenze lieferte Token und Konto-ID zusammen. Der Prozess durfte externe Ziele nur über einen absichtlich unerreichbaren Proxy erreichen.
- Zwei gleichzeitige serverseitige Codex-Zugangsanforderungen gegen eine kontrollierte OAuth-HTTP-Gegenstelle lösten genau eine zurückgehaltene Erneuerungsanfrage aus. Beide erhielten denselben erneuerten Zugang für dasselbe Konto; nach erneutem Öffnen des geschützten Speichers blieb er nutzbar. Eine endgültig verweigerte Erneuerung gab kein Zugangspaar aus, und die öffentliche Statusantwort verlangte eine neue Anmeldung.
- Die öffentliche OpenRouter-Settings-HTTP-Grenze zeigte einen widerrufenen Schlüssel als nicht einsatzbereit und erlaubte dessen bewussten Ersatz unter derselben Referenz. Die zweite kontrollierte Anbieterprüfung verwendete nur den neuen Schlüssel. Echte HTML- und JSON-Antworten des gebauten Serverprozesses enthielten weder Schlüssel noch Tokens.
- Über `openrouterverbindung.Service.Binding` und `Resolve` blieb eine geprüfte Bindung bei Rotation und erneut geöffnetem geschütztem Speicher gültig. Trennen und neues Verbinden widerriefen die alte Bindung und erzeugten eine neue. Eine geschützte Bestandsverbindung ohne frühere Bindungsgeneration blieb nach der Migration über dieselbe öffentliche Servicegrenze auflösbar und nach erneutem Öffnen gültig.
- Nach Integration von Story-24 legte das neunte Szenario eine Organisation und einen Eino-Agenten über öffentliche HTTP-Grenzen an, speicherte und prüfte den OpenRouter-Key über Settings-HTTP und erteilte Organisations- und Agentenfreigaben über deren öffentliche HTTP-Grenze. Ein kontrollierter `modellfreigabe.Caller` löste `Access` unmittelbar vor seinem lokalen Provider-POST auf. Der erste `Invoke` führte genau einen POST aus; nach Settings-Disconnect wurde ein neuer `Invoke` wegen fehlender Verbindung und nach Reconnect unter derselben Referenz ohne neue Freigaben wegen fehlender Organisationsfreigabe abgelehnt. Der Anbieter erhielt in beiden Fällen keinen weiteren POST. Das zuvor ausstehende Szenario ist damit grün und sein Pending-Tag entfernt.

## Offene Gates

- Der echte Eino-Modelltransport über ein Codex-Abonnement und dessen Nutzung durch einen Produktagenten bleiben beim externen Story-08-Gate. Die kontrollierten OAuth- und Resolver-Prüfungen belegen diesen Transport nicht.
