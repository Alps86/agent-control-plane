# language: de
@story-09 @story-23 @settings @browser
Funktionalität: Modellanbieter über die gemeinsamen Settings erreichen
  Als Betreiber
  möchte ich aus Settings beide Anbieteransichten öffnen und zurückkehren,
  damit ich die jeweilige Verbindung am passenden Ort verwalte.

  Szenario: Gemeinsame Settings als vollständige Seite ausliefern
    Angenommen die lokale Anwendung läuft mit den beiden Anbieteransichten
    Wenn ich GET "/settings" ohne HTMX-Kopfzeile aufrufe
    Dann erhalte ich HTTP 200 und ein vollständiges HTML-Dokument
    Und Settings enthält lokale Links nach "/settings/modelle/codex" und "/settings/modellanbieter/openrouter"
    Und Settings zeigt keine Anmeldung, Nutzung, Budgets oder Geheimnisse

  Szenario: Dieselben Anbieterlinks als HTMX-Inhalt ausliefern
    Angenommen die lokale Anwendung läuft mit den beiden Anbieteransichten
    Wenn ich GET "/settings" mit HTMX-Kopfzeile aufrufe
    Dann erhalte ich HTTP 200 und nur den Settings-Inhalt
    Und Settings enthält lokale Links nach "/settings/modelle/codex" und "/settings/modellanbieter/openrouter"
    Und Settings zeigt keine Anmeldung, Nutzung, Budgets oder Geheimnisse

  Szenariogrundriss: Anbieteransicht im Browser öffnen und nach Settings zurückkehren
    Angenommen die lokale Anwendung läuft mit den beiden Anbieteransichten
    Wenn ich im Browser Settings öffne und den Link "<Anbieter>" wähle
    Dann erreiche ich die Anbieteransicht "<Pfad>"
    Wenn ich dort den Rücklink nach Settings wähle
    Dann sehe ich im Browser wieder "/settings" mit beiden Anbieterlinks

    Beispiele:
      | Anbieter  | Pfad                                   |
      | Codex-Abo | /settings/modelle/codex                |
      | OpenRouter | /settings/modellanbieter/openrouter   |

  Szenario: Unsichere öffentliche Bind-Adresse verhindert den Start
    Angenommen die Anwendung wird mit APP_ADDR "0.0.0.0:0" gestartet
    Wenn der Startvorgang abgeschlossen ist
    Dann ist der Prozess ohne öffentlichen Settings-Listener fehlgeschlagen
