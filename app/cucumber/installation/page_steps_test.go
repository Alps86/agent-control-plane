package installation

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerPageSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich ohne Cookie oder Authorization-Header die App-Startseite öffne$`, s.openHome)
	sc.Step(`^sehe ich die echte Organisationsübersicht oder ihren Leerzustand$`, s.overview)
	sc.Step(`^ich werde nicht auf eine App-Anmeldung oder Registrierung geleitet$`, s.noAuthRedirect)
	sc.Step(`^die Seite enthält keine App-Anmelde- oder Registrierungsaktion$`, s.noAuthAction)
	sc.Step(`^die Seite zeigt keine Budgetfelder oder Budgetlimits$`, s.noBudgetPage)
	sc.Step(`^ich ohne Cookie oder Authorization-Header das Organisationsformular öffne$`, s.openForm)
	sc.Step(`^kann ich Name und Beschreibung eingeben$`, s.formFields)
	sc.Step(`^ich sehe keine Eingabe für Budget oder Budgetlimit$`, s.noBudgetInput)
	sc.Step(`^ich die Organisation "([^"]*)" ohne App-Anmeldung anlege$`, s.createForm)
	sc.Step(`^sehe ich "([^"]*)" in der echten Organisationsübersicht$`, s.seeInOverview)
	sc.Step(`^ich sehe "([^"]*)" in der echten Organisationsübersicht$`, s.seeInOverview)
	sc.Step(`^die Antwort enthält keine Budgetfelder oder Budgetlimits$`, s.noBudgetPage)
	sc.Step(`^ich habe ohne Cookie oder Authorization-Header das öffentliche Organisationsformular geöffnet$`, s.openedForm)
	sc.Step(`^ich das Formular mit Name "([^"]*)", einer Beschreibung und dem zusätzlich eingeschleusten Feld (.+) als HTML-Form-POST absende$`, s.postInjectedForm)
	s.registerProviderSteps(sc)
}

func (s *Suite) registerProviderSteps(sc *godog.ScenarioContext) {
	sc.Step(`^es ist noch kein Modellanbieter verbunden$`, s.noProvider)
	sc.Step(`^ich ohne Cookie oder Authorization-Header die App-Startseite und das Organisationsformular öffne$`, s.openBoth)
	sc.Step(`^kann ich eine Organisation "([^"]*)" ohne App-Anmeldung anlegen$`, s.canCreate)
	sc.Step(`^die fehlende Modellanbieter-Verbindung hat die lokale Organisationsanlage nicht gesperrt$`, s.providerDidNotBlock)
}

func (s *Suite) openHome() error {
	return s.get("/organisationen")
}

func (s *Suite) overview() error {
	if s.status != http.StatusOK || !strings.HasPrefix(s.header.Get("Content-Type"), "text/html") {
		return fmt.Errorf("Organisationsübersicht: HTTP %d, Content-Type %q", s.status, s.header.Get("Content-Type"))
	}

	if !strings.Contains(strings.ToLower(string(s.body)), "organisation") {
		return fmt.Errorf("keine Organisationsübersicht im HTML")
	}

	return nil
}

func (s *Suite) noAuthRedirect() error {
	if s.status != http.StatusOK || s.header.Get("Location") != "" {
		return fmt.Errorf("Startseite erfordert Anmeldung oder leitet weiter: HTTP %d, Location %q", s.status, s.header.Get("Location"))
	}

	return nil
}

func (s *Suite) noAuthAction() error {
	page := strings.ToLower(string(s.body))
	for _, token := range []string{"/login", "/register", "/registrierung", "type=\"password\""} {
		if strings.Contains(page, token) {
			return fmt.Errorf("App-Anmeldeaktion %q gefunden", token)
		}
	}

	return nil
}

func (s *Suite) noBudgetPage() error {
	if strings.Contains(strings.ToLower(string(s.body)), "budget") {
		return fmt.Errorf("Budgetinhalt in öffentlicher Antwort")
	}

	return nil
}

func (s *Suite) openForm() error {
	return s.get("/organisationen/neu")
}

func (s *Suite) openedForm() error {
	if err := s.openForm(); err != nil {
		return err
	}

	return s.formFields()
}

func (s *Suite) formFields() error {
	if s.status != http.StatusOK || !strings.HasPrefix(s.header.Get("Content-Type"), "text/html") {
		return fmt.Errorf("Organisationsformular: HTTP %d", s.status)
	}

	name := regexp.MustCompile(`(?is)<input\b[^>]*\bname\s*=\s*["']name["']`).Match(s.body)
	description := regexp.MustCompile(`(?is)<textarea\b[^>]*\bname\s*=\s*["']description["']`).Match(s.body)
	if !name || !description {
		return fmt.Errorf("Name/Beschreibung fehlen im Organisationsformular")
	}

	return nil
}

func (s *Suite) noBudgetInput() error {
	return s.noBudgetPage()
}

func (s *Suite) createForm(name string) error {
	form := url.Values{"name": {name}, "description": {"Wissensarbeit"}}
	if err := s.post("/organisationen", "application/x-www-form-urlencoded", strings.NewReader(form.Encode())); err != nil {
		return err
	}

	if s.status != http.StatusSeeOther {
		return fmt.Errorf("Organisationsanlage: HTTP %d statt 303", s.status)
	}

	s.created = true
	return nil
}

func (s *Suite) postInjectedForm(name, field string) error {
	key, value, ok := strings.Cut(field, "=")
	if !ok || value == "" {
		return fmt.Errorf("ungültiges zusätzliches Formularfeld %q", field)
	}

	s.budget = value
	form := url.Values{"name": {name}, "description": {"Wissensarbeit"}, key: {value}}
	return s.post("/organisationen", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
}

func (s *Suite) seeInOverview(name string) error {
	if err := s.openHome(); err != nil {
		return err
	}

	if err := s.overview(); err != nil {
		return err
	}

	if !strings.Contains(string(s.body), name) {
		return fmt.Errorf("Organisation %q fehlt in echter Übersicht", name)
	}

	return nil
}

func (s *Suite) noProvider() error {
	// Die Testinstallation startet mit neuer Datenbank; kein Step verbindet einen Anbieter.
	if !s.freshDB {
		return fmt.Errorf("Anbieterfreiheit ist nur für die frische Testinstallation belegt")
	}

	return nil
}

func (s *Suite) openBoth() error {
	if err := s.openHome(); err != nil {
		return err
	}

	if err := s.overview(); err != nil {
		return err
	}

	if err := s.openForm(); err != nil {
		return err
	}

	return s.formFields()
}

func (s *Suite) canCreate(name string) error {
	return s.createForm(name)
}

func (s *Suite) providerDidNotBlock() error {
	if !s.created {
		return fmt.Errorf("ohne Anbieter-Verbindung wurde keine Organisation angelegt")
	}

	return nil
}
