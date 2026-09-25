# language: de
@story-25 @ms-05
Funktionalität: Registrierten Modellzugang für einen Eino-Agenten auswählen
  Als Ersteller wähle ich eine zulässige Verbindung und ein konkretes Modell.
  Die Anbieterkonfiguration bestimmt Authentifizierung und belegte Fähigkeiten.

  Grundlage:
    Angenommen ein lokaler Server mit kontrollierten Modellkatalogen läuft
    Und die Organisation "Nord" mit dem Eino-Agenten "Mira" besteht

  Szenario: Codex-Abo ist die anfängliche Auswahl
    Wenn ich Miras Modellwahl über HTTP abrufe
    Dann ist "Codex-Abo" als Standardanbieter ausgewiesen
    Und die Ausführungsart von "Mira" ist "Eino"
    Und Verbindungskennungen und Modellkennungen stammen aus dem registrierten Katalog
    Und das Codex-Modell ist ohne Konto-Nachweis als ungeprüft markiert
    Und der Katalog zeigt nur nicht geheime Verbindungsreferenzen

  Szenario: OpenRouter ohne ausdrücklich konfigurierte Verbindung wird abgewiesen
    Angenommen der kontrollierte Katalog enthält OpenRouter ohne Verbindungsreferenz
    Wenn ich einen Serverprozess mit diesem Katalog starte
    Dann endet der Start mit einem verständlichen Hinweis auf die fehlende OpenRouter-Verbindung
    Und es entsteht keine öffentliche Modellwahl-Grenze aus dieser Konfiguration

  Szenario: Freigegebenes OpenRouter-Modell bewusst auswählen
    Angenommen die zentrale OpenRouter-Verbindung ist einsatzbereit und für "Nord" und "Mira" freigegeben
    Wenn ich für "Mira" die OpenRouter-Verbindung und das Modell "openai/gpt-4" über HTTP speichere
    Dann zeigt Miras öffentliche Modellwahlansicht "OpenRouter" und "openai/gpt-4"
    Und die Ausführungsart von "Mira" bleibt "Eino"
    Und die öffentliche Modellwahlantwort enthält den hinterlegten Schlüssel nicht
    Und die Modellwahl bleibt nach einem Serverneustart erhalten

  Szenariogrundriss: Fehlende Freigabe sperrt die Auswahl
    Angenommen die zentrale OpenRouter-Verbindung ist einsatzbereit
    Und ihre Freigabe für <Ebene> fehlt
    Wenn ich für "Mira" die OpenRouter-Verbindung und das Modell "openai/gpt-4" über HTTP speichere
    Dann wird die Modellwahl mit einem Hinweis auf die fehlende <Ebene> abgelehnt
    Und die öffentliche Antwort hat HTTP-Status 403
    Und Miras bisherige Modellwahl bleibt erhalten

    Beispiele:
      | Ebene        |
      | Organisation |
      | Agent        |

  Szenariogrundriss: Fremde Kennungen sind keine auswählbare Modellroute
    Wenn ich für "Mira" <Auswahl> über HTTP speichere
    Dann wird die Modellwahl mit einem verständlichen Kataloghinweis abgelehnt
    Und die öffentliche Antwort hat HTTP-Status 422
    Und Miras bisherige Modellwahl bleibt erhalten

    Beispiele:
      | Auswahl                                      |
      | den unbekannten Anbieter "fremd"            |
      | die fremde Verbindung "andere-verbindung"   |
      | das nicht katalogisierte Modell "fremd/neu" |

  Szenario: Authentifizierung und Fähigkeiten kommen aus der Anbieterkonfiguration
    Angenommen die zentrale OpenRouter-Verbindung ist einsatzbereit und für "Nord" und "Mira" freigegeben
    Wenn ich Miras Modellwahl über HTTP abrufe
    Dann nennt der Katalog für OpenRouter seine konfigurierte Authentifizierungsart
    Und er zeigt für "openai/gpt-4" Modellquelle, Stand und Prüfstatus
    Und er zeigt für Textfähigkeit Quelle, Prüfzeit und Status
    Und er führt Text, Tools, Audioeingabe, Audioausgabe und Echtzeit einzeln auf
    Und nicht belegte Echtzeit-Sprache wird als "Nicht nachgewiesen" gekennzeichnet

  Szenario: Eine fremde Organisation kann Miras Modellwahl nicht lesen oder ändern
    Wenn ich Miras Modellwahl aus einer fremden Organisation über HTTP lese und ändere
    Dann erhalte ich keine Modellwahl oder Katalogdaten von "Mira"
    Und die fremde Antwort entspricht einer unbekannten Agentenkennung mit HTTP 404
    Und Miras bisherige Modellwahl bleibt erhalten
