package organisation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) initializeScenario(sc *godog.ScenarioContext) {
	sc.After(s.afterScenario)
	s.registerAPISteps(sc)
	s.registerBrowserSteps(sc)
}

func (s *Suite) registerAPISteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin$`, s.freshServer)
	sc.Step(`^ich die Organisationsübersicht über HTTP abrufe$`, s.getList)
	sc.Step(`^antwortet der Server erfolgreich mit einer leeren Organisationsliste$`, s.emptyListResponse)
	sc.Step(`^ich das Formular für eine neue Organisation über HTTP abrufe$`, s.getForm)
	sc.Step(`^enthält das Formular die Felder "Name" und "Beschreibung"$`, s.formFields)
	sc.Step(`^das Formular zeigt keine vorgegebene Organisation$`, s.formHasNoOrganization)
	sc.Step(`^ich die Organisation "([^"]*)" mit der Beschreibung "([^"]*)" über HTTP anlege$`, s.postOrganization)
	sc.Step(`^die Organisation "([^"]*)" mit der Beschreibung "([^"]*)" wurde über HTTP angelegt$`, s.organizationExists)
	sc.Step(`^ich eine Organisation mit dem Namen "([^"]*)" und der Beschreibung "([^"]*)" über HTTP anlege$`, s.postOrganization)
	s.registerAPIAssertions(sc)
}

func (s *Suite) registerAPIAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^antwortet der Server mit einer neuen Organisationskennung$`, s.createdResponse)
	sc.Step(`^die Organisationsübersicht enthält "([^"]*)" mit der Beschreibung "([^"]*)" genau einmal$`, s.listContainsOnce)
	sc.Step(`^enthält die Organisationsübersicht "([^"]*)" mit der Beschreibung "([^"]*)" genau einmal$`, s.listContainsOnce)
	sc.Step(`^die Organisationsdetails zeigen einen restriktiven Standard für Arbeits- und Delegationsübergänge$`, s.detailsRestricted)
	sc.Step(`^die Organisationsdetails zeigen weiterhin den restriktiven Standard für Arbeits- und Delegationsübergänge$`, s.detailsRestricted)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank erneut starte$`, s.restartServer)
	sc.Step(`^antwortet der Server mit einem Hinweis am Feld "Name"$`, s.nameError)
	sc.Step(`^antwortet der Server mit einem Konflikthinweis am Feld "Name"$`, s.nameConflict)
	sc.Step(`^die Organisationsübersicht bleibt leer$`, s.listEmpty)
	sc.Step(`^keine Organisation enthält die Beschreibung "([^"]*)"$`, s.descriptionAbsent)
	sc.Step(`^der Server für einen HTTP-Aufruf keine eindeutige Betreiberidentität ermittelt$`, s.noIdentity)
	sc.Step(`^verweigert er das Lesen der Organisationsübersicht ohne Organisationsdaten$`, s.deniedList)
	sc.Step(`^er verweigert das Anlegen einer Organisation ohne Teilerzeugung$`, s.deniedCreate)
	s.registerSecurityAssertions(sc)
}

func (s *Suite) registerSecurityAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^ich "([^"]+)" mit dem Namen "([^"]*)" und der Beschreibung "([^"]*)" von Origin "([^"]+)" per HTTP POST aufrufe$`, s.foreignOriginPost)
	sc.Step(`^ich "([^"]+)" mit dem Namen "([^"]*)" und der Beschreibung "([^"]*)" über Host "([^"]+)" per HTTP POST aufrufe$`, s.foreignHostPost)
	sc.Step(`^antwortet der Server mit dem HTTP-Status (\d+)$`, s.statusCode)
	sc.Step(`^ich eine unbekannte Organisationskennung über HTTP abfrage$`, s.getUnknown)
	sc.Step(`^antwortet der Server mit einem datenfreien HTTP-Status 404$`, s.unknown404)
	sc.Step(`^ich die vorhandene Organisationskennung als nicht zugeordnete Betreiberin über HTTP abfrage$`, s.getUnassigned)
	sc.Step(`^antwortet der Server mit demselben datenfreien HTTP-Status 404$`, s.same404)
}

func (s *Suite) foreignOriginPost(path, name, description, origin string) error {
	return s.postFrom(path, name, description, origin, "")
}

func (s *Suite) foreignHostPost(path, name, description, host string) error {
	return s.postFrom(path, name, description, "", host)
}

func (s *Suite) postFrom(path, name, description, origin, host string) error {
	body, contentType, err := s.postBody(path, name, description)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", s.baseURL()+path, strings.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", contentType)
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	if host != "" {
		request.Host = host
	}
	return s.sendPost(request)
}

func (s *Suite) sendPost(request *http.Request) error {
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	return s.recordResponse(response)
}

func (s *Suite) postBody(path, name, description string) (string, string, error) {
	if path == "/organisationen" {
		values := url.Values{"name": {name}, "description": {description}}
		return values.Encode(), "application/x-www-form-urlencoded", nil
	}

	body, err := json.Marshal(map[string]string{"name": name, "description": description})
	return string(body), "application/json", err
}

func (s *Suite) statusCode(status int) error {
	if s.response.Status != status {
		return fmt.Errorf("HTTP-Status %d statt %d: %s", s.response.Status, status, s.response.Body)
	}
	return nil
}

func (s *Suite) getList() error { return s.request("GET", "/api/organisationen", "", "") }
func (s *Suite) getForm() error { return s.request("GET", "/organisationen/neu", "", "") }

func (s *Suite) recordResponse(response *http.Response) error {
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	s.response = &HTTPResponse{Status: response.StatusCode, Location: response.Header.Get("Location"), Body: data, Header: response.Header}
	return nil
}

func (s *Suite) emptyListResponse() error {
	if s.response.Status != 200 {
		return fmt.Errorf("leere Liste: Status %d: %s", s.response.Status, s.response.Body)
	}
	var list OrganizationList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	if list.Organizations == nil || len(list.Organizations) != 0 {
		return fmt.Errorf("Liste nicht explizit leer: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) formFields() error {
	if s.response.Status != 200 {
		return fmt.Errorf("Formular: Status %d: %s", s.response.Status, s.response.Body)
	}
	text := string(s.response.Body)
	for _, fragment := range []string{`<label for="organisation-name">Name</label>`, `name="name"`, `<label for="organisation-description">Beschreibung</label>`, `name="description"`} {
		if !strings.Contains(text, fragment) {
			return fmt.Errorf("Formularfeld fehlt: %s", fragment)
		}
	}
	return nil
}

func (s *Suite) formHasNoOrganization() error {
	text := string(s.response.Body)
	if strings.Contains(text, `value="Nordstern"`) || strings.Contains(text, "Wissensarbeit für das Team") {
		return fmt.Errorf("Formular enthält Organisation")
	}
	return nil
}

func (s *Suite) postOrganization(name, description string) error {
	body, err := json.Marshal(map[string]string{"name": name, "description": description})
	if err != nil {
		return err
	}
	return s.request("POST", "/api/organisationen", string(body), "application/json")
}

func (s *Suite) organizationExists(name, description string) error {
	if err := s.postOrganization(name, description); err != nil {
		return err
	}
	return s.createdResponse()
}

func (s *Suite) createdResponse() error {
	if s.response.Status != http.StatusCreated {
		return fmt.Errorf("Erstellung: Status %d: %s", s.response.Status, s.response.Body)
	}
	if err := json.Unmarshal(s.response.Body, &s.organization); err != nil {
		return err
	}
	if s.organization.ID == "" {
		return fmt.Errorf("keine Organisationskennung: %s", s.response.Body)
	}
	if s.response.Location != "/api/organisationen/"+s.organization.ID {
		return fmt.Errorf("Location %q", s.response.Location)
	}
	return s.assertPolicy(s.organization)
}

func (s *Suite) listContainsOnce(name, description string) error {
	if err := s.getList(); err != nil {
		return err
	}
	list, err := s.parseList()
	if err != nil {
		return err
	}
	count := 0
	for _, item := range list.Organizations {
		if item.Name == name && item.Description == description {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("%q/%q: %d Einträge in %s", name, description, count, s.response.Body)
	}
	return nil
}

func (s *Suite) parseList() (OrganizationList, error) {
	var list OrganizationList
	if s.response.Status != 200 {
		return list, fmt.Errorf("Liste: Status %d: %s", s.response.Status, s.response.Body)
	}
	err := json.Unmarshal(s.response.Body, &list)
	return list, err
}

func (s *Suite) detailsRestricted() error {
	if err := s.request("GET", "/api/organisationen/"+s.organization.ID, "", ""); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Detail: Status %d: %s", s.response.Status, s.response.Body)
	}
	var item Organization
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID != s.organization.ID || item.Name != s.organization.Name {
		return fmt.Errorf("Detail abweichend: %s", s.response.Body)
	}
	return s.assertPolicy(item)
}

func (s *Suite) assertPolicy(item Organization) error {
	if item.WorkflowPolicy.WorkTransitions == nil || item.WorkflowPolicy.DelegationTransitions == nil {
		return fmt.Errorf("Policy nicht sichtbar: %+v", item.WorkflowPolicy)
	}
	if len(item.WorkflowPolicy.WorkTransitions) != 0 || len(item.WorkflowPolicy.DelegationTransitions) != 0 {
		return fmt.Errorf("Policy nicht restriktiv: %+v", item.WorkflowPolicy)
	}
	return nil
}

func (s *Suite) nameError() error    { return s.assertNameError(http.StatusUnprocessableEntity) }
func (s *Suite) nameConflict() error { return s.assertNameError(http.StatusConflict) }

func (s *Suite) assertNameError(status int) error {
	if s.response.Status != status {
		return fmt.Errorf("Name-Fehler: Status %d: %s", s.response.Status, s.response.Body)
	}
	var failure FieldError
	if err := json.Unmarshal(s.response.Body, &failure); err != nil {
		return err
	}
	if failure.FieldErrors["name"] == "" {
		return fmt.Errorf("kein Feldhinweis: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) listEmpty() error {
	if err := s.getList(); err != nil {
		return err
	}
	return s.emptyListResponse()
}

func (s *Suite) descriptionAbsent(description string) error {
	if err := s.getList(); err != nil {
		return err
	}
	list, err := s.parseList()
	if err != nil {
		return err
	}
	for _, item := range list.Organizations {
		if item.Description == description {
			return fmt.Errorf("unerlaubte Beschreibung gespeichert: %s", s.response.Body)
		}
	}
	return nil
}
