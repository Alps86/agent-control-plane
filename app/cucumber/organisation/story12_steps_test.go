package organisation

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) registerStory12Steps(sc *godog.ScenarioContext) {
	sc.Step(`^ich lege die Organisation "([^"]*)" mit der Beschreibung "([^"]*)" über die öffentliche JSON-API an$`, s.organizationExists)
	sc.Step(`^ich die Organisationsübersicht im Browser öffne$`, s.openStoredList)
	sc.Step(`^enthält die HTML-Übersicht "([^"]*)" mit "([^"]*)" genau einmal$`, s.storedListOnce)
	sc.Step(`^enthält die HTML-Übersicht "([^"]*)" mit "([^"]*)" erneut genau einmal$`, s.storedListOnce)
	sc.Step(`^die Detailseite dieser Organisation zeigt denselben Namen und dieselbe Beschreibung$`, s.storedDetail)
	sc.Step(`^ich den Server mit derselben SQLite-Datenbank neu starte$`, s.restartStoredPage)
	sc.Step(`^die öffentliche JSON-API liefert dieselbe Organisationskennung$`, s.sameStoredID)
	sc.Step(`^ich die Organisationsübersicht als vollständige HTML-Seite und mit "HX-Request: true" abrufe$`, s.getFullAndFragment)
	sc.Step(`^enthalten beide Antworten "([^"]*)" und "([^"]*)"$`, s.bothContain)
	s.registerStory12Assertions(sc)
}

func (s *Suite) registerStory12Assertions(sc *godog.ScenarioContext) {
	sc.Step(`^die vollständige Antwort enthält ein HTML-Dokument mit Navigation$`, s.fullDocument)
	sc.Step(`^das Fragment enthält nur den Inhalt der Organisationsübersicht ohne HTML-Dokument$`, s.onlyFragment)
	sc.Step(`^ich die Organisationsübersicht und die Detailseite im Browser öffne$`, s.openStoredListAndDetail)
	sc.Step(`^sehe ich beide Fachtexte als Text$`, s.maliciousTextsVisible)
	sc.Step(`^es wurde weder ein Script ausgeführt noch ein Bild-Fehlerhandler ausgelöst$`, s.noExecutedContent)
	sc.Step(`^die HTML-Antworten enthalten kein ausführbares Script oder Bild aus diesen Fachtexten$`, s.htmlTextsEscaped)
	sc.Step(`^eine andere serverseitige Betreiberidentität die Organisationsübersicht und Detailseite direkt über HTTP abruft$`, s.foreignHTML)
	sc.Step(`^enthält ihre Übersicht weder "([^"]*)" noch "([^"]*)"$`, s.foreignListEmpty)
	sc.Step(`^die Detailseite antwortet wie für eine unbekannte Kennung ohne Organisationsdaten$`, s.foreignDetailHidden)
	sc.Step(`^ich ohne gültige serverseitige Betreiberidentität "POST /organisationen" mit Name "([^"]*)" direkt aufrufe$`, s.deniedHTMLPost)
	sc.Step(`^antwortet die HTML-Route mit HTTP-Status 403 ohne Organisationsdaten$`, s.deniedHTMLResponse)
	sc.Step(`^die öffentliche JSON-API der Betreiberin zeigt keine neu angelegte Organisation$`, s.ownListEmpty)
	sc.Step(`^ich "POST /organisationen" mit gültigem Namen "([^"]*)" und unbekanntem Feld "([^"]*)" direkt aufrufe$`, s.postUnknownFormField)
	sc.Step(`^antwortet die HTML-Route mit HTTP-Status 400 oder 422 ohne neue Organisation$`, s.unknownFormFieldRejected)
}

func (s *Suite) openStoredList() error {
	if err := s.startBrowser(); err != nil {
		return err
	}
	if err := s.browserNavigate("/organisationen", "list"); err != nil {
		return err
	}
	s.listPage = s.page
	return nil
}

func (s *Suite) storedListOnce(name, description string) error {
	count := 0
	for _, card := range s.listPage.Cards {
		if card.Name == name && card.Description == description {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("Browserliste %q/%q: %d Karten: %+v", name, description, count, s.listPage.Cards)
	}
	return nil
}

func (s *Suite) storedDetail() error {
	if err := s.browserNavigate("/organisationen/"+s.organization.ID, "detail"); err != nil {
		return err
	}
	if s.page.Heading != s.organization.Name || !strings.Contains(s.page.Text, s.organization.Description) {
		return fmt.Errorf("Browserdetail weicht von JSON ab: %+v", s.page)
	}
	return nil
}

func (s *Suite) restartStoredPage() error {
	if err := s.restartServer(); err != nil {
		return err
	}
	return s.openStoredList()
}

func (s *Suite) sameStoredID() error {
	if err := s.getList(); err != nil {
		return err
	}
	list, err := s.parseList()
	if err != nil {
		return err
	}
	if len(list.Organizations) != 1 || list.Organizations[0].ID != s.organization.ID {
		return fmt.Errorf("Kennung nach Neustart: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) getFullAndFragment() error {
	if err := s.request("GET", "/organisationen", "", ""); err != nil {
		return err
	}
	s.fullPage = s.response
	req, err := http.NewRequest("GET", s.baseURL()+"/organisationen", nil)
	if err != nil {
		return err
	}
	req.Header.Set("HX-Request", "true")
	return s.fetchFragment(req)
}

func (s *Suite) fetchFragment(req *http.Request) error {
	response, err := s.client.Do(req)
	if err != nil {
		return err
	}
	if err := s.recordResponse(response); err != nil {
		return err
	}
	s.fragment = s.response
	return nil
}

func (s *Suite) bothContain(name, description string) error {
	for _, response := range []*HTTPResponse{s.fullPage, s.fragment} {
		if response == nil || response.Status != 200 {
			return fmt.Errorf("HTML-Antwort fehlt oder Status nicht 200: %+v", response)
		}
		if !strings.Contains(string(response.Body), html.EscapeString(name)) || !strings.Contains(string(response.Body), html.EscapeString(description)) {
			return fmt.Errorf("Fachdaten fehlen: %s", response.Body)
		}
	}
	return nil
}

func (s *Suite) fullDocument() error {
	body := strings.ToLower(string(s.fullPage.Body))
	if !strings.Contains(body, "<!doctype html>") || !strings.Contains(body, "<nav>") || !strings.Contains(body, "<html") {
		return fmt.Errorf("Vollseite ohne Dokument/Navigation: %s", s.fullPage.Body)
	}
	return nil
}

func (s *Suite) onlyFragment() error {
	body := strings.ToLower(string(s.fragment.Body))
	if strings.Contains(body, "<!doctype") || strings.Contains(body, "<html") || strings.Contains(body, "<nav") {
		return fmt.Errorf("Fragment enthält Dokumenthülle: %s", s.fragment.Body)
	}
	if !strings.Contains(body, "organisationsübersicht") {
		return fmt.Errorf("Fragment ohne Übersicht: %s", s.fragment.Body)
	}
	return nil
}

func (s *Suite) openStoredListAndDetail() error {
	if err := s.openStoredList(); err != nil {
		return err
	}
	return s.storedDetail()
}

func (s *Suite) maliciousTextsVisible() error {
	if err := s.storedListOnce(s.organization.Name, s.organization.Description); err != nil {
		return err
	}
	return s.storedDetail()
}

func (s *Suite) noExecutedContent() error {
	for _, page := range []BrowserPage{s.listPage, s.page} {
		if page.AlertCalls != 0 || page.Images != 0 {
			return fmt.Errorf("Browser führte Inhalt aus: alerts=%d images=%d", page.AlertCalls, page.Images)
		}
	}
	return nil
}

func (s *Suite) htmlTextsEscaped() error {
	if err := s.request("GET", "/organisationen", "", ""); err != nil {
		return err
	}
	s.fullPage = s.response
	if err := s.request("GET", "/organisationen/"+s.organization.ID, "", ""); err != nil {
		return err
	}
	s.detailPage = s.response
	return s.assertEscapedResponses()
}

func (s *Suite) assertEscapedResponses() error {
	for _, response := range []*HTTPResponse{s.fullPage, s.detailPage} {
		body := string(response.Body)
		if response.Status != 200 || strings.Contains(body, s.organization.Name) || strings.Contains(body, s.organization.Description) {
			return fmt.Errorf("Fachtext nicht escaped: %s", body)
		}
		if !strings.Contains(body, "&lt;script&gt;") || !strings.Contains(body, "&lt;img") {
			return fmt.Errorf("escaped Texte fehlen: %s", body)
		}
	}
	return nil
}

func (s *Suite) foreignHTML() error {
	if err := s.getUnassigned(); err != nil {
		return err
	}
	response, err := s.facadeGet("/organisationen")
	if err != nil {
		return err
	}
	s.foreignList = response
	s.foreignDetail, err = s.facadeGet("/organisationen/" + s.organization.ID)
	return err
}

func (s *Suite) foreignListEmpty(name, description string) error {
	if s.foreignList.Status != 200 {
		return fmt.Errorf("fremde Übersicht: %d: %s", s.foreignList.Status, s.foreignList.Body)
	}
	body := string(s.foreignList.Body)
	if strings.Contains(body, name) || strings.Contains(body, description) {
		return fmt.Errorf("fremde Daten in Übersicht: %s", body)
	}
	return nil
}

func (s *Suite) foreignDetailHidden() error {
	if s.foreignDetail.Status != 404 {
		return fmt.Errorf("fremdes Detail: %d: %s", s.foreignDetail.Status, s.foreignDetail.Body)
	}
	if strings.Contains(string(s.foreignDetail.Body), s.organization.Name) || strings.Contains(string(s.foreignDetail.Body), s.organization.ID) {
		return fmt.Errorf("fremde Daten in Detail: %s", s.foreignDetail.Body)
	}
	unknown, err := s.facadeGet("/organisationen/gibt-es-nicht")
	if err != nil {
		return err
	}
	if unknown.Status != s.foreignDetail.Status || string(unknown.Body) != string(s.foreignDetail.Body) {
		return fmt.Errorf("fremdes Detail unterscheidet sich von unbekannter Kennung")
	}
	return nil
}

func (s *Suite) deniedHTMLPost(name string) error {
	if err := s.noIdentity(); err != nil {
		return err
	}
	request, err := http.NewRequest("POST", s.negative.Server.URL+"/organisationen", strings.NewReader(url.Values{"name": {name}}.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	return s.recordResponse(response)
}

func (s *Suite) deniedHTMLResponse() error {
	if s.response.Status != 403 || strings.Contains(string(s.response.Body), "Unerlaubt") {
		return fmt.Errorf("HTML-POST: %d: %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) ownListEmpty() error {
	if err := s.getList(); err != nil {
		return err
	}
	return s.emptyListResponse()
}

func (s *Suite) postUnknownFormField(name, field string) error {
	values := url.Values{"name": {name}, "description": {"Nicht übernehmen"}, field: {"9999"}}
	return s.request("POST", "/organisationen", values.Encode(), "application/x-www-form-urlencoded")
}

func (s *Suite) unknownFormFieldRejected() error {
	if s.response.Status != 400 && s.response.Status != 422 {
		return fmt.Errorf("unbekanntes Formularfeld angenommen: %d, Location %q", s.response.Status, s.response.Location)
	}
	if strings.Contains(string(s.response.Body), "budgetLimit") || strings.Contains(string(s.response.Body), "Budgetversuch") {
		return fmt.Errorf("abgewiesene Eingabe im Fehlertext: %s", s.response.Body)
	}
	return nil
}
