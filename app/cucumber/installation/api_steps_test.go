package installation

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) registerAPISteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich ohne Cookie oder Authorization-Header über die öffentliche Organisations-API "([^"]*)" mit dem zusätzlichen JSON-Feld (.+) anlege$`, s.createWithBudget)
	sc.Step(`^wird die Eingabe als ungültig abgewiesen$`, s.invalidInput)
	sc.Step(`^es wurde keine Organisation "([^"]*)" angelegt$`, s.notCreated)
	sc.Step(`^die Antwort enthält weder das übermittelte Budget noch ein Budgetlimit$`, s.noBudgetEcho)
	sc.Step(`^eine Organisation "([^"]*)" wurde über die öffentliche API ohne App-Anmeldung angelegt$`, s.createViaAPI)
	sc.Step(`^ich sie ohne Cookie oder Authorization-Header über die öffentliche Organisations-API lese$`, s.readViaAPI)
	sc.Step(`^antwortet die API mit der Organisation "([^"]*)"$`, s.apiName)
	sc.Step(`^ich über eine Nicht-Loopback-Adresse ohne Origin mit Host "localhost" eine Organisation "([^"]*)" anlege$`, s.remoteCreate)
	sc.Step(`^ich über Loopback ohne Origin mit Host "localhost" eine Organisation "([^"]*)" anlege$`, s.wildcardCreate)
	sc.Step(`^wird der Schreibzugriff mit HTTP 403 abgewiesen$`, s.remoteForbidden)
}

func (s *Suite) remoteCreate(name string) error {
	remote, err := s.nonLoopbackAddress()
	if err != nil {
		return err
	}

	return s.createWithHost(remote, name)
}

func (s *Suite) wildcardCreate(name string) error {
	return s.createWithHost("127.0.0.1", name)
}

func (s *Suite) createWithHost(target, name string) error {
	_, port, _ := net.SplitHostPort(s.bindAddress)
	payload := fmt.Sprintf(`{"name":%q,"description":"Fremdzugriff"}`, name)
	request, err := http.NewRequest(http.MethodPost, "http://"+net.JoinHostPort(target, port)+"/api/organisationen", strings.NewReader(payload))
	if err != nil {
		return err
	}

	request.Host = "localhost:" + port
	request.Header.Set("Content-Type", "application/json")
	return s.remotePost(request)
}

func (s *Suite) nonLoopbackAddress() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addresses {
		ip, _, err := net.ParseCIDR(address.String())
		if err == nil && ip.To4() != nil && ip.IsGlobalUnicast() {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("keine Nicht-Loopback-Adresse für öffentlichen HTTP-Test verfügbar")
}

func (s *Suite) remotePost(request *http.Request) error {
	client := &http.Client{Timeout: time.Second, Transport: &http.Transport{Proxy: nil}}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	s.status, s.header = response.StatusCode, response.Header.Clone()
	s.body, err = io.ReadAll(response.Body)
	return err
}

func (s *Suite) remoteForbidden() error {
	if s.status != http.StatusForbidden || !strings.Contains(string(s.body), "access_denied") {
		return fmt.Errorf("entfernter POST: HTTP %d, Antwort %s", s.status, s.body)
	}

	return nil
}

func (s *Suite) createWithBudget(name, extra string) error {
	s.budget = "100"
	payload := fmt.Sprintf(`{"name":%q,"description":"Wissensarbeit",%s}`, name, extra)
	return s.post("/api/organisationen", "application/json", strings.NewReader(payload))
}

func (s *Suite) invalidInput() error {
	if s.status != http.StatusBadRequest && s.status != http.StatusUnprocessableEntity {
		return fmt.Errorf("Budgeteingabe nicht als ungültig abgewiesen: HTTP %d", s.status)
	}

	s.rejected = append([]byte(nil), s.body...)
	s.rejectedHeader = s.header.Clone()
	return nil
}

func (s *Suite) notCreated(name string) error {
	if err := s.get("/api/organisationen"); err != nil {
		return err
	}

	if s.status != http.StatusOK || !strings.HasPrefix(s.header.Get("Content-Type"), "application/json") {
		return fmt.Errorf("Organisationsliste nicht lesbar: HTTP %d", s.status)
	}

	return s.assertNotListed(name)
}

func (s *Suite) assertNotListed(name string) error {
	var body map[string]any
	if err := json.Unmarshal(s.body, &body); err != nil {
		return err
	}

	organizations, ok := body["organizations"].([]any)
	if !ok {
		return fmt.Errorf("API-Liste enthält kein organizations-Array")
	}

	if s.hasName(organizations, name) {
		return fmt.Errorf("abgewiesene Organisation %q trotzdem gespeichert", name)
	}

	return nil
}

func (s *Suite) hasName(entries []any, name string) bool {
	for _, child := range entries {
		organization, ok := child.(map[string]any)
		if ok && organization["name"] == name {
			return true
		}
	}

	return false
}

func (s *Suite) noBudgetEcho() error {
	if strings.Contains(string(s.rejected), s.budget) {
		return fmt.Errorf("übermittelter Budgetwert wird ausgegeben")
	}

	if err := s.checkRejectedBudget(); err != nil {
		return err
	}

	return s.noBudget()
}

func (s *Suite) checkRejectedBudget() error {
	contentType := s.rejectedHeader.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		return s.checkNoBudgetJSON(s.rejected)
	}

	if !strings.HasPrefix(contentType, "text/html") {
		return fmt.Errorf("Validierungsantwort ohne HTML/JSON: %q", contentType)
	}

	if strings.Contains(strings.ToLower(string(s.rejected)), "budget") {
		return fmt.Errorf("Budgetinhalt in HTML-Validierungsantwort")
	}

	return nil
}

func (s *Suite) createViaAPI(name string) error {
	if err := s.start(); err != nil {
		return err
	}

	payload := fmt.Sprintf(`{"name":%q,"description":"Wissensarbeit"}`, name)
	if err := s.post("/api/organisationen", "application/json", strings.NewReader(payload)); err != nil {
		return err
	}

	if s.status != http.StatusCreated {
		return fmt.Errorf("API-Anlage: HTTP %d statt 201", s.status)
	}

	return s.captureID()
}

func (s *Suite) captureID() error {
	var body map[string]any
	if err := json.Unmarshal(s.body, &body); err != nil {
		return err
	}

	id, ok := body["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("API-Anlage liefert keine Organisations-ID")
	}

	s.orgID = id
	if s.header.Get("Location") != "/api/organisationen/"+id {
		return fmt.Errorf("API-Location passt nicht zur Organisations-ID: %q", s.header.Get("Location"))
	}

	return nil
}

func (s *Suite) readViaAPI() error {
	if s.orgID == "" {
		return fmt.Errorf("keine zuvor angelegte Organisations-ID")
	}

	return s.get("/api/organisationen/" + s.orgID)
}

func (s *Suite) apiName(name string) error {
	if s.status != http.StatusOK || !strings.HasPrefix(s.header.Get("Content-Type"), "application/json") {
		return fmt.Errorf("Organisation nicht per API lesbar: HTTP %d", s.status)
	}

	var body map[string]any
	if err := json.Unmarshal(s.body, &body); err != nil {
		return err
	}

	if body["name"] != name {
		return fmt.Errorf("API-Name: %v statt %q", body["name"], name)
	}

	if body["description"] != "Wissensarbeit" {
		return fmt.Errorf("API-Beschreibung fehlt oder wurde verändert")
	}

	return nil
}
