package projektarchiv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func (s *Suite) readActivity(org string) (ActivityList, error) {
	var list ActivityList
	if err := s.request("GET", s.orgPath(org)+"/aktivitaet", ""); err != nil {
		return list, err
	}
	if err := s.status(http.StatusOK); err != nil {
		return list, err
	}
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return list, err
	}
	if list.Events == nil {
		return list, fmt.Errorf("Aktivitätsliste fehlt: %s", s.response.Body)
	}
	return list, nil
}

func (s *Suite) activitySetup(title string) error {
	list, err := s.readActivity(s.activeOrg)
	if err != nil {
		return err
	}
	for _, event := range list.Events {
		if event.TaskID != s.tasks[s.key(s.activeOrg, title)] || event.Kind != "created" {
			continue
		}
		if err := s.validateActivity(event, title); err != nil {
			return err
		}
		s.activity = event
		return nil
	}
	return fmt.Errorf("Anlageeintrag für %q fehlt: %s", title, s.response.Body)
}

func (s *Suite) validateActivity(event ActivityEvent, title string) error {
	pattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !pattern.MatchString(event.ID) {
		return fmt.Errorf("ungültige Aktivitätskennung: %+v", event)
	}
	_, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil || !strings.HasSuffix(event.OccurredAt, "Z") {
		return fmt.Errorf("UTC-Zeit fehlt: %+v", event)
	}
	if event.OrganizationID != s.organizations[s.activeOrg] || event.ProjectID != s.projects[s.key(s.activeOrg, s.activeProject)] {
		return fmt.Errorf("Aktivitätsbezug: %+v", event)
	}
	if event.ObjectTitle != title || event.Actor != "local-operator" || event.Source != "api" {
		return fmt.Errorf("Aktivitätsherkunft: %+v", event)
	}
	if !strings.Contains(event.DeepLink, s.tasks[s.key(s.activeOrg, title)]) {
		return fmt.Errorf("Aufgabenlink fehlt: %+v", event)
	}
	return nil
}

func (s *Suite) activityPreserved(title string) error {
	list, err := s.readActivity(s.activeOrg)
	if err != nil {
		return err
	}
	for _, event := range list.Events {
		if event.ID != s.activity.ID {
			continue
		}
		if !reflect.DeepEqual(event, s.activity) {
			return fmt.Errorf("Aktivität verändert: %+v statt %+v", event, s.activity)
		}
		return s.validateActivity(event, title)
	}
	return fmt.Errorf("Aktivität %s nach Archivierung/Neustart fehlt: %s", s.activity.ID, s.response.Body)
}

func (s *Suite) activityNotLeaked(org, title string) error {
	list, err := s.readActivity(org)
	if err != nil {
		return err
	}
	for _, event := range list.Events {
		if event.ID == s.activity.ID || event.TaskID == s.activity.TaskID || event.ObjectTitle == title {
			return fmt.Errorf("fremde Aktivität sichtbar: %+v", event)
		}
	}
	if strings.Contains(string(s.response.Body), s.activity.DeepLink) {
		return fmt.Errorf("fremder Link sichtbar: %s", s.response.Body)
	}
	return nil
}
