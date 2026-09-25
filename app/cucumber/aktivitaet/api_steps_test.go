package aktivitaet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	sc.Step(`^der lokale Server läuft mit einer neuen SQLite-Datenbank$`, s.fresh)
	sc.Step(`^in "([^"]+)" bestehen das Projekt "([^"]+)" und der Agent "([^"]+)"$`, s.ensureProjectAndAgent)
	sc.Step(`^die Organisation "([^"]+)" besteht$`, s.ensureOrganization)
	sc.Step(`^ich "([^"]+)" im Projekt "([^"]+)" dem Agenten "([^"]+)" über die öffentliche API zuweise$`, s.createTask)
	sc.Step(`^"([^"]+)" ist im Projekt "([^"]+)" dem Agenten "([^"]+)" zugewiesen$`, s.createTask)
	sc.Step(`^ich eine Aufgabe ohne Titel im Projekt "([^"]+)" dem Agenten "([^"]+)" über die öffentliche API zuweise$`, s.createUntitled)
	sc.Step(`^ich eine Aufgabenanlage mit den behaupteten Werten "([^"]+)" und "([^"]+)" für Akteur und Quelle sende$`, s.spoof)
	sc.Step(`^ich "([^"]+)" über das öffentliche Aufgabenformular für "([^"]+)" anlege$`, s.createThroughForm)
	s.registerAssertions(sc)
	s.registerRollback(sc)
	s.registerBrowser(sc)
}

func (s *Suite) registerAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^ist die Aufgabe "([^"]+)" angelegt und "([^"]+)" zugewiesen$`, s.taskCreated)
	sc.Step(`^der Aktivitätsverlauf von "([^"]+)" enthält genau ein Anlageereignis für "([^"]+)"$`, s.createdEvent)
	sc.Step(`^der Aktivitätsverlauf von "([^"]+)" enthält genau ein Zuweisungsereignis an "([^"]+)" für "([^"]+)"$`, s.assignedEvent)
	sc.Step(`^beide Ereignisse nennen die serverseitige Betreiberkennung und "([^"]+)" als Quelle$`, s.actorAndSource)
	sc.Step(`^beide Ereignisse haben einen gültigen Zeitpunkt und einen Deep Link zur gespeicherten Aufgabe$`, s.timestampsAndLinks)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank neu starte$`, s.captureAndRestart)
	sc.Step(`^enthält der Aktivitätsverlauf von "([^"]+)" dieselben zwei Ereignisse mit denselben Kennungen und Zeitpunkten$`, s.sameEvents)
	sc.Step(`^beide Deep Links öffnen die ursprüngliche Aufgabe$`, s.linksOpenTask)
	sc.Step(`^enthält der Aktivitätsverlauf von "([^"]+)" keine Ereignisse aus "([^"]+)"$`, s.noForeignEvents)
	sc.Step(`^keine Kennung eines "([^"]+)"-Ereignisses erscheint im Verlauf von "([^"]+)"$`, s.noForeignIDs)
	sc.Step(`^wird die Aufgabenanlage mit einem Feldfehler abgewiesen$`, s.fieldError)
	sc.Step(`^im Aktivitätsverlauf von "([^"]+)" steht kein Aufgabenereignis$`, s.noEvents)
	sc.Step(`^wird die Aufgabenanlage ohne neues Ereignis abgewiesen$`, s.spoofRejected)
	sc.Step(`^enthält der Aktivitätsverlauf von "([^"]+)" genau zwei Ereignisse mit Quelle "([^"]+)"$`, s.twoSources)
}

func (s *Suite) createTask(title, project, agent string) error {
	if err := s.postTask(title, project, agent, nil); err != nil {
		return err
	}

	return s.taskCreated(title, agent)
}

func (s *Suite) createUntitled(project, agent string) error {
	return s.postTask("", project, agent, nil)
}

func (s *Suite) postTask(title, project, agent string, extra map[string]string) error {
	org := "Nordstern"
	value := map[string]string{"title": title, "description": "Prüfen", "priority": "normal", "assignee_id": s.agents[s.key(org, agent)]}
	for key, item := range extra {
		value[key] = item
	}

	data, _ := json.Marshal(value)
	return s.request("POST", s.taskPath(org, project), string(data))
}

func (s *Suite) spoof(actor, source string) error {
	return s.postTask("Manipulierter Auftrag", "Website", "Mira", map[string]string{"actor": actor, "source": source})
}

func (s *Suite) createThroughForm(title, project string) error {
	path := "/organisationen/" + s.organizations["Nordstern"] + "/projekte/" + s.projects[s.key("Nordstern", project)] + "/aufgaben/neu"
	values := url.Values{"title": {title}, "description": {"Prüfen"}, "priority": {"normal"},
		"assignee_id": {s.agents[s.key("Nordstern", "Mira")]}, "project_id": {s.projects[s.key("Nordstern", project)]}, "project_name": {project}}
	request, err := http.NewRequest("POST", s.url(path), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return s.submitForm(request)
}

func (s *Suite) submitForm(request *http.Request) error {
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusSeeOther {
		return fmt.Errorf("Formular HTTP %d", response.StatusCode)
	}

	s.taskID = strings.TrimPrefix(response.Header.Get("Location"), "/organisationen/"+s.organizations["Nordstern"]+"/projekte/"+s.projects[s.key("Nordstern", "Website")]+"/aufgaben/")
	if s.taskID == "" || strings.Contains(s.taskID, "/") {
		return fmt.Errorf("Aufgaben-Location fehlt: %s", response.Header.Get("Location"))
	}

	return nil
}

func (s *Suite) twoSources(org, source string) error {
	events, err := s.events(org)
	if err != nil {
		return err
	}

	if len(events) != 2 {
		return fmt.Errorf("Ereignisanzahl: %+v", events)
	}

	for _, item := range events {
		if item.Source != source || item.Actor != "local-operator" || item.TaskID != s.taskID {
			return fmt.Errorf("Formularquelle falsch: %+v", item)
		}
	}

	return nil
}

func (s *Suite) taskCreated(title, agent string) error {
	if s.status != http.StatusCreated {
		return fmt.Errorf("Aufgabenanlage HTTP %d: %s", s.status, s.body)
	}

	var task struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		AssigneeID string `json:"assignee_id"`
	}
	if err := json.Unmarshal(s.body, &task); err != nil {
		return err
	}

	if task.ID == "" || task.Title != title || task.AssigneeID != s.agents[s.key("Nordstern", agent)] {
		return fmt.Errorf("Aufgabe unvollständig: %+v", task)
	}

	s.taskID = task.ID
	return nil
}

func (s *Suite) events(org string) ([]Event, error) {
	path := "/api/organisationen/" + s.organizations[org] + "/aktivitaet"
	if err := s.request("GET", path, ""); err != nil {
		return nil, err
	}

	if s.status != http.StatusOK {
		return nil, fmt.Errorf("Aktivität HTTP %d: %s", s.status, s.body)
	}

	var list EventList
	if err := json.Unmarshal(s.body, &list); err != nil {
		return nil, err
	}

	return list.Events, nil
}

func (s *Suite) event(org, kind, title string) (Event, error) {
	events, err := s.events(org)
	if err != nil {
		return Event{}, err
	}

	var match Event
	count := 0
	for _, item := range events {
		if item.Kind == kind && item.ObjectTitle == title && item.TaskID == s.taskID {
			match, count = item, count+1
		}
	}

	if count != 1 || len(events) != 2 {
		return Event{}, fmt.Errorf("Ereignisse %s: %+v", kind, events)
	}

	return match, nil
}

func (s *Suite) createdEvent(org, title string) error {
	item, err := s.event(org, "created", title)
	if err != nil {
		return err
	}

	if item.AssigneeID != "" || item.AssigneeName != "" {
		return fmt.Errorf("Anlage enthält verfrühte Zuweisung: %+v", item)
	}

	return nil
}

func (s *Suite) assignedEvent(org, agent, title string) error {
	item, err := s.event(org, "assigned", title)
	if err != nil {
		return err
	}

	if item.ProjectID != s.projects[s.key(org, "Website")] || item.AssigneeID != s.agents[s.key(org, agent)] || item.AssigneeName != agent {
		return fmt.Errorf("Zuweisungsobjekt fehlt: %+v", item)
	}

	path := s.taskPath(org, "Website") + "/" + item.TaskID
	if err := s.request("GET", path, ""); err != nil {
		return err
	}

	if s.status != http.StatusOK || !strings.Contains(string(s.body), s.agents[s.key(org, agent)]) {
		return fmt.Errorf("Zuweisung im Task fehlt: HTTP %d: %s", s.status, s.body)
	}

	return nil
}

func (s *Suite) actorAndSource(source string) error {
	events, err := s.events("Nordstern")
	if err != nil {
		return err
	}

	for _, item := range events {
		if item.Actor != "local-operator" || item.Source != source {
			return fmt.Errorf("Akteur oder Quelle falsch: %+v", item)
		}
	}

	return nil
}

func (s *Suite) timestampsAndLinks() error {
	events, err := s.events("Nordstern")
	if err != nil {
		return err
	}

	for _, item := range events {
		if err := s.validateEvent(item); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) validateEvent(item Event) error {
	stamp, err := time.Parse(time.RFC3339Nano, item.OccurredAt)
	path := "/organisationen/" + item.OrganizationID + "/projekte/" + item.ProjectID + "/aufgaben/" + item.TaskID
	if err != nil || stamp.Location() != time.UTC || stamp.After(time.Now().Add(time.Minute)) {
		return fmt.Errorf("Zeitpunkt ungültig: %+v", item)
	}

	if item.ID == "" || item.OrganizationID != s.organizations["Nordstern"] || item.TaskID != s.taskID || item.DeepLink != path {
		return fmt.Errorf("Objekt oder Deep Link ungültig: %+v", item)
	}

	return nil
}

func (s *Suite) captureAndRestart() error {
	events, err := s.events("Nordstern")
	if err != nil {
		return err
	}

	s.before = events
	return s.restart()
}

func (s *Suite) sameEvents(org string) error {
	events, err := s.events(org)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(events, s.before) || len(events) != 2 {
		return fmt.Errorf("Ereignisse nach Neustart: vorher %+v, nachher %+v", s.before, events)
	}

	return nil
}

func (s *Suite) linksOpenTask() error {
	for _, item := range s.before {
		if err := s.request("GET", item.DeepLink, ""); err != nil {
			return err
		}

		if s.status != http.StatusOK || !strings.Contains(string(s.body), "Startseite prüfen") {
			return fmt.Errorf("Deep Link %s: HTTP %d", item.DeepLink, s.status)
		}
	}

	return nil
}

func (s *Suite) noForeignEvents(own, foreign string) error {
	events, err := s.events(own)
	if err != nil {
		return err
	}

	for _, item := range events {
		if item.OrganizationID == s.organizations[foreign] || item.TaskID == s.taskID {
			return fmt.Errorf("Fremdes Ereignis: %+v", item)
		}
	}

	return nil
}

func (s *Suite) noForeignIDs(foreign, own string) error {
	foreignEvents, err := s.events(foreign)
	if err != nil {
		return err
	}

	ownEvents, err := s.events(own)
	if err != nil {
		return err
	}

	for _, first := range foreignEvents {
		for _, second := range ownEvents {
			if first.ID == second.ID {
				return fmt.Errorf("Ereigniskennung in beiden Organisationen: %s", first.ID)
			}
		}
	}

	return nil
}

func (s *Suite) fieldError() error {
	if s.status != http.StatusUnprocessableEntity || !strings.Contains(string(s.body), "title") {
		return fmt.Errorf("Feldfehler fehlt: HTTP %d: %s", s.status, s.body)
	}

	return nil
}

func (s *Suite) noEvents(org string) error {
	events, err := s.events(org)
	if err != nil {
		return err
	}

	if len(events) != 0 {
		return fmt.Errorf("Unerwartete Ereignisse: %+v", events)
	}

	return nil
}

func (s *Suite) spoofRejected() error {
	if s.status < 400 || s.status >= 500 {
		return fmt.Errorf("Manipulation akzeptiert: HTTP %d", s.status)
	}

	return s.noEvents("Nordstern")
}
