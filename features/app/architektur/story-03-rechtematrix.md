# Story-03 · ARC-03: Bootstrap-Rechtematrix

Diese Matrix beschreibt nur den öffentlichen Beispiel-Leseweg von ARC-03. Die
fachlichen Daten-, Tool- und Delegationsrechte werden in PERM-01 bis PERM-05
festgelegt. Die Beispielressourcen sind serverseitige Bootstrap-Daten; sie
setzen weder eine Organisationsverwaltung noch einen Agentenlauf voraus.

## Entscheidungsdaten

| Feld | Bedeutung im Bootstrap |
| --- | --- |
| Akteur | Genau ein vom Server ermittelter Betreiber oder Agent. HTTP-Parameter und Header dürfen keine Akteuridentität behaupten. |
| Organisationszuordnung | Genau eine serverseitige Zuordnung des Akteurs zu einer Organisation. Für einen Agenten muss diese Zuordnung zur eigenen Organisation gehören. |
| Ressource | Genau eine serverseitig bekannte Beispielressource mit Organisationskennung; für Agenten zusätzlich ausdrücklich dem Agenten zugeordnet. |
| Aktion | `lesen` auf dem Beispiel-Leseweg. Andere Aktionen besitzen in ARC-03 keine Freigabe. |

| Akteur | Zuordnung | Zielressource | Aktion | Entscheidung |
| --- | --- | --- | --- | --- |
| Betreiber | Eindeutig zu Organisation A | In Organisation A | Lesen | Erlauben |
| Agent | Eindeutig zu Organisation A und Ressource R | R in Organisation A | Lesen | Erlauben |
| Betreiber oder Agent | Eindeutig zu Organisation A | In Organisation B | Lesen | Verweigern; Existenz nicht offenlegen |
| Agent | Eindeutig zu Organisation A, ohne Zuordnung zu R | R in Organisation A | Lesen | Verweigern; Existenz nicht offenlegen |
| Betreiber oder Agent | Keine oder mehrere Akteuridentitäten | Beliebig | Lesen | Verweigern |
| Betreiber oder Agent | Keine oder mehrere Organisationszuordnungen | Beliebig | Lesen | Verweigern |
| Betreiber oder Agent | Eindeutig | Unbekannte ID | Lesen | Verweigern; gleiche Antwort wie bei fremder Ressource |
| Betreiber oder Agent | Beliebig | Beliebig | Schreiben, Toolausführung, Delegation | Keine ARC-03-Freigabe; spätere PERM-Stories |

Die Entscheidung erfolgt an der Anwendungsgrenze unmittelbar vor der Ausgabe
der Ressourcendaten. Ein erlaubter Aufruf liefert die eigene Ressource. Ein
Aufruf für eine fremde, nicht zugeordnete oder unbekannte Ressource liefert
identisch `404` und `{"error":"resource_not_found"}`. Fehlt eine eindeutige
Akteuridentität oder Organisationszuordnung, liefert der Weg `403` und
`{"error":"access_denied"}`. Beide Fehlerkörper enthalten keine
Ressourcenkennung, keinen Ressourcennamen und keine fremde
Organisationskennung. Die Identität und Zuordnungen stammen nur aus
serverseitiger Konfiguration; ein vom Client gesetzter Header oder Parameter
erweitert sie nicht.

## Öffentlicher Bootstrap-Vertrag

`GET /api/architektur/rechte/ressourcen/{id}` liest eine Beispielressource
über die zentrale Rechteentscheidung. Der Server kann für die Blackbox-Abnahme
mit einer eindeutig konfigurierten Betreiber- oder Agentenidentität und einem
festen Satz von Beispielressourcen gestartet werden. Diese Route ist nur ein
Bootstrap-Nachweis; spätere Fachrouten nutzen dieselbe Entscheidungsgrenze,
ohne die Beispielressourcen zu übernehmen.
