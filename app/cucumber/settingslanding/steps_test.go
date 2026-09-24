package settingslanding

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) InitializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.cleanup)
	sc.Step(`^die lokale Anwendung läuft mit den beiden Anbieteransichten$`, s.startLocal)
	sc.Step(`^ich GET "/settings" ohne HTMX-Kopfzeile aufrufe$`, s.fullSettings)
	sc.Step(`^ich GET "/settings" mit HTMX-Kopfzeile aufrufe$`, s.fragmentSettings)
	sc.Step(`^erhalte ich HTTP 200 und ein vollständiges HTML-Dokument$`, s.fullDocument)
	sc.Step(`^erhalte ich HTTP 200 und nur den Settings-Inhalt$`, s.contentOnly)
	sc.Step(`^Settings enthält lokale Links nach "/settings/modelle/codex" und "/settings/modellanbieter/openrouter"$`, s.localLinks)
	sc.Step(`^Settings zeigt keine Anmeldung, Nutzung, Budgets oder Geheimnisse$`, s.noPrivateData)
	sc.Step(`^ich im Browser Settings öffne und den Link "([^"]*)" wähle$`, s.browseProvider)
	sc.Step(`^erreiche ich die Anbieteransicht "([^"]*)"$`, s.providerPage)
	sc.Step(`^ich dort den Rücklink nach Settings wähle$`, s.returnLink)
	sc.Step(`^sehe ich im Browser wieder "/settings" mit beiden Anbieterlinks$`, s.returnedSettings)
	sc.Step(`^die Anwendung wird mit APP_ADDR "([^"]*)" gestartet$`, s.unsafeAddress)
	sc.Step(`^der Startvorgang abgeschlossen ist$`, s.startupFinished)
	sc.Step(`^ist der Prozess ohne öffentlichen Settings-Listener fehlgeschlagen$`, s.noPublicListener)
}

func (s *Suite) fullSettings() error     { return s.getSettings(false) }
func (s *Suite) fragmentSettings() error { return s.getSettings(true) }

func (s *Suite) getSettings(htmx bool) error {
	request, err := http.NewRequest(http.MethodGet, s.baseURL+"/settings", nil)
	if err != nil {
		return err
	}
	if htmx {
		request.Header.Set("HX-Request", "true")
	}
	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	s.status, s.contentType = response.StatusCode, response.Header.Get("Content-Type")
	body, err := io.ReadAll(response.Body)
	s.body = string(body)
	return err
}

func (s *Suite) fullDocument() error {
	if s.status != http.StatusOK || !strings.HasPrefix(s.contentType, "text/html") || !strings.Contains(strings.ToLower(s.body), "<!doctype html>") || !strings.Contains(s.body, "</html>") {
		return fmt.Errorf("Settings full response: HTTP %d, Content-Type %q", s.status, s.contentType)
	}
	return nil
}

func (s *Suite) contentOnly() error {
	if s.status != http.StatusOK || !strings.HasPrefix(s.contentType, "text/html") || strings.Contains(strings.ToLower(s.body), "<!doctype html>") || !strings.Contains(s.body, "<h1>Settings</h1>") {
		return fmt.Errorf("Settings HTMX response: HTTP %d, Content-Type %q", s.status, s.contentType)
	}
	return nil
}

func (s *Suite) localLinks() error {
	links := map[string]string{"settings-codex-link": "/settings/modelle/codex", "settings-openrouter-link": "/settings/modellanbieter/openrouter"}
	for id, path := range links {
		if err := s.checkLink(id, path); err != nil {
			return err
		}
	}
	return nil
}

func (s *Suite) checkLink(id, path string) error {
	tag := regexp.MustCompile(`(?is)<a\b[^>]*\bid="` + regexp.QuoteMeta(id) + `"[^>]*>`).FindString(s.body)
	if tag == "" || !strings.Contains(tag, `href="`+path+`"`) {
		return fmt.Errorf("Settings link %s does not point to %s", id, path)
	}
	return nil
}

func (s *Suite) noPrivateData() error {
	lower := strings.ToLower(s.body)
	for _, term := range []string{"access_token", "refresh_token", "api_key", "account_id", "user_code", "budget", "nutzung", "anmeldung läuft", "schlüssel", "sk-"} {
		if strings.Contains(lower, term) {
			return fmt.Errorf("Settings landing exposes %q", term)
		}
	}
	if strings.Contains(lower, "<form") {
		return fmt.Errorf("Settings landing includes credential form")
	}
	return nil
}
