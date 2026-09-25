package organisationswechsel

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
)

func (s *Suite) revokeStored(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	current, err := s.rules(name)
	if err != nil {
		return err
	}
	if !s.hasStatusRule(current) {
		return fmt.Errorf("gespeicherte Kante vor Widerruf fehlt")
	}
	s.revokeRevision = current["revision"]
	body := map[string]any{"operation": "revoke", "area": "status", "from": "todo", "to": "in_progress", "revision": s.revokeRevision}
	path := "/api/organisationen/" + url.PathEscape(org.ID) + "/arbeitsregeln"
	if err := s.call("POST", path+"/vorschau", body); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Widerrufsvorschau: %d %s", s.response.Status, s.response.Body)
	}
	var preview map[string]any
	if err := json.Unmarshal(s.response.Body, &preview); err != nil {
		return err
	}
	if preview["operation"] != "revoke" || preview["area"] != "status" {
		return fmt.Errorf("Widerrufsvorschau falsch: %v", preview)
	}
	delete(body, "operation")
	if err := s.call("DELETE", path, body); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Widerruf: %d %s", s.response.Status, s.response.Body)
	}
	s.revokedSnapshot, err = s.rules(name)
	return err
}

func (s *Suite) hasStatusRule(snapshot map[string]any) bool {
	for _, item := range s.areaRules(snapshot, "status") {
		rule, _ := item.(map[string]any)
		if rule["from"] == "todo" && rule["to"] == "in_progress" && rule["approver"] == "betreiber" {
			return true
		}
	}

	return false
}

func (s *Suite) ruleRevoked(name string) error {
	current, err := s.rules(name)
	if err != nil {
		return err
	}
	if s.hasStatusRule(current) {
		return fmt.Errorf("widerrufene Kante weiterhin freigegeben")
	}
	if s.revokedSnapshot == nil || !reflect.DeepEqual(current["rules"], s.revokedSnapshot["rules"]) {
		return fmt.Errorf("Widerruf-Snapshot instabil")
	}
	return nil
}

func (s *Suite) revokeStale() error {
	org, err := s.organization("Nordstern")
	if err != nil {
		return err
	}
	body := map[string]any{"area": "status", "from": "todo", "to": "in_progress", "revision": s.revokeRevision}
	return s.call("DELETE", "/api/organisationen/"+url.PathEscape(org.ID)+"/arbeitsregeln", body)
}

func (s *Suite) revisionConflict() error {
	if s.response.Status != 409 {
		return fmt.Errorf("veraltete Revision: %d %s", s.response.Status, s.response.Body)
	}
	current, err := s.rules("Nordstern")
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, s.revokedSnapshot) {
		return fmt.Errorf("Revisionskonflikt änderte Regeln")
	}
	return nil
}
