package organisationswechsel

import (
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"net/url"
	"reflect"
	"strings"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	sc.Step(`^ich starte für Story 31 einen lokalen Server mit neuer SQLite-Datenbank$`, s.fresh)
	sc.Step(`^ich lege über die öffentliche API die Organisationen "([^"]+)" und "([^"]+)" an$`, s.createTwo)
	sc.Step(`^ich "([^"]+)" in der Organisationsliste über HTTP suche$`, s.search)
	sc.Step(`^enthält die Suchantwort "([^"]+)" und nicht "([^"]+)"$`, s.searchResult)
	sc.Step(`^ich über HTTP zu "([^"]+)" wechsle$`, s.switchTo)
	sc.Step(`^ist "([^"]+)" der gewählte Organisationskontext$`, s.selected)
	sc.Step(`^die direkte Detail-URL von "([^"]+)" zeigt dessen Namen und Beschreibung$`, s.detail)
	sc.Step(`^ich die Arbeitsregeln von "([^"]+)" über HTTP abfrage$`, s.readRules)
	sc.Step(`^sehe ich getrennte Regeln für Aufgabenstatus, Zuweisung und Delegation mit zuständiger Freigabe$`, s.threeAreas)
	sc.Step(`^ich für "([^"]+)" den Aufgabenstatus-Übergang "([^"]+)" nach "([^"]+)" mit Freigabe "([^"]+)" vorschauprüfe und speichere$`, s.statusChange)
	sc.Step(`^ich für "([^"]+)" den (Zuweisung|Delegation)-Übergang ([^ ]+) nach ([^ ]+) mit Freigabe "([^"]+)" vorschauprüfe und speichere$`, s.namedChange)
	sc.Step(`^nennt die Vorschau betroffene (Aufgabenstatus|Zuweisung|Delegation)-Abläufe und keine (?:Zuweisungs- oder Delegationsänderung|Änderung anderer Regelbereiche)$`, s.previewArea)
	sc.Step(`^die gespeicherte (Aufgabenstatus|Zuweisung|Delegation)-Regel ist nur bei "([^"]+)" sichtbar$`, s.savedArea)
	sc.Step(`^Zuweisungs- und Delegationsregeln bleiben unverändert$`, s.otherAreasStable)
	sc.Step(`^die Arbeitsregeln von "([^"]+)" bleiben unverändert$`, s.rulesStable)
	sc.Step(`^ich für "([^"]+)" eine (unbekannter Aufgabenstatus|fehlende zuständige Freigabe|Rechteausweitung|Workflow-DSL) Regeländerung über HTTP vorschauprüfe und zu speichern versuche$`, s.invalidChange)
	sc.Step(`^benennt die Antwort den Ablehnungsgrund ohne Regeländerung$`, s.invalidReason)
	sc.Step(`^alle drei Regelbereiche von "([^"]+)" bleiben unverändert$`, s.rulesStable)
	sc.Step(`^ich die gespeicherte Aufgabenstatus-Regel für "([^"]+)" über HTTP widerrufe$`, s.revokeStored)
	sc.Step(`^ist diese Regel bei "([^"]+)" nicht mehr freigegeben$`, s.ruleRevoked)
	sc.Step(`^ich den Widerruf mit einer veralteten Revision wiederhole$`, s.revokeStale)
	sc.Step(`^erhalte ich einen Revisionskonflikt ohne Regeländerung$`, s.revisionConflict)
	sc.Step(`^"([^"]+)" ist für die aktuelle serverseitige Identität nicht zugeordnet$`, s.foreignIdentity)
	sc.Step(`^ich dessen direkte API-Detail-URL, HTML-Detail-URL und Wechselroute über HTTP abrufe$`, s.foreignRoutes)
	sc.Step(`^erhalte ich datenfreie Antworten ohne "([^"]+)" und dessen Beschreibung$`, s.noForeignData)
	sc.Step(`^die Antworten sind nicht von einer unbekannten Organisationskennung unterscheidbar$`, s.sameUnknown)
	sc.Step(`^"([^"]+)" bleibt der gewählte Organisationskontext$`, s.selected)
	sc.Step(`^ich eine Delegationsregel für "([^"]+)" über eine direkte URL vorschauprüfe und zu speichern versuche$`, s.foreignChange)
	sc.Step(`^erhalte ich datenfreie Antworten wie für eine unbekannte Organisationskennung$`, s.sameUnknown)
	s.registerUISteps(sc)
}

func (s *Suite) createTwo(a, b string) error {
	if err := s.create(a); err != nil {
		return err
	}
	if err := s.create(b); err != nil {
		return err
	}
	s.initialRules = map[string]map[string]any{}
	for _, name := range []string{a, b} {
		snapshot, err := s.rules(name)
		if err != nil {
			return err
		}
		s.initialRules[name] = snapshot
	}
	return nil
}

func (s *Suite) search(q string) error {
	return s.call("GET", "/api/organisationswechsel?q="+url.QueryEscape(q), nil)
}

func (s *Suite) searchResult(want, absent string) error {
	if s.response.Status != 200 {
		return fmt.Errorf("Suche: %d %s", s.response.Status, s.response.Body)
	}
	var list struct {
		Organizations []Organization `json:"organizations"`
	}
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	seen := false
	for _, org := range list.Organizations {
		if org.Name == want {
			seen = true
		}
		if org.Name == absent {
			return fmt.Errorf("Fremdtreffer %s", absent)
		}
	}
	if !seen {
		return fmt.Errorf("Suchtreffer %s fehlt: %s", want, s.response.Body)
	}
	return nil
}

func (s *Suite) switchTo(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	return s.call("POST", "/api/organisationen/"+url.PathEscape(org.ID)+"/wechsel", map[string]any{})
}

func (s *Suite) selected(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Wechselstatus %d: %s", s.response.Status, s.response.Body)
	}
	var selection struct {
		ID string `json:"selected_organization_id"`
	}
	if err := json.Unmarshal(s.response.Body, &selection); err != nil {
		return err
	}
	if selection.ID != org.ID {
		return fmt.Errorf("gewählt %q, erwartet %q", selection.ID, org.ID)
	}
	if len(s.response.Header.Values("Set-Cookie")) == 0 {
		return fmt.Errorf("Wechsel speichert keinen Browserkontext")
	}
	return nil
}

func (s *Suite) detail(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	if err := s.call("GET", "/api/organisationen/"+url.PathEscape(org.ID), nil); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Detail: %d %s", s.response.Status, s.response.Body)
	}
	var got Organization
	if err := json.Unmarshal(s.response.Body, &got); err != nil {
		return err
	}
	if got.ID != org.ID || got.Name != org.Name || got.Description != org.Description {
		return fmt.Errorf("Detailabweichung: %+v / %+v", got, org)
	}
	return nil
}

func (s *Suite) rules(name string) (map[string]any, error) {
	org, err := s.organization(name)
	if err != nil {
		return nil, err
	}
	if err := s.call("GET", "/api/organisationen/"+url.PathEscape(org.ID)+"/arbeitsregeln", nil); err != nil {
		return nil, err
	}
	if s.response.Status != 200 {
		return nil, fmt.Errorf("Regeln: %d %s", s.response.Status, s.response.Body)
	}
	var snapshot map[string]any
	if err := json.Unmarshal(s.response.Body, &snapshot); err != nil {
		return nil, err
	}
	if snapshot["organization_id"] != org.ID {
		return nil, fmt.Errorf("falsche Organisation in Regeln: %v", snapshot["organization_id"])
	}
	return snapshot, nil
}

func (s *Suite) readRules(name string) error {
	snapshot, err := s.rules(name)
	if err != nil {
		return err
	}
	s.baseline = snapshot
	s.lastOrganization = name
	return nil
}

func (s *Suite) threeAreas() error {
	if s.baseline == nil {
		return fmt.Errorf("kein Regel-Snapshot")
	}
	rules, ok := s.baseline["rules"].([]any)
	if !ok {
		return fmt.Errorf("rules fehlt: %v", s.baseline)
	}
	for _, item := range rules {
		rule, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("ungültige Rule %v", item)
		}
		area, _ := rule["area"].(string)
		if area != "status" && area != "assignment" && area != "delegation" {
			return fmt.Errorf("unbekannter Bereich %s", area)
		}
	}
	if _, ok := s.baseline["revision"].(float64); !ok {
		return fmt.Errorf("Revision fehlt")
	}
	areas, ok := s.baseline["areas"].([]any)
	if !ok || len(areas) != 3 {
		return fmt.Errorf("drei Bereichskennungen fehlen: %v", s.baseline["areas"])
	}
	for index, want := range []string{"status", "assignment", "delegation"} {
		if areas[index] != want {
			return fmt.Errorf("Bereich %d: %v statt %s", index, areas[index], want)
		}
	}
	return nil
}

func (s *Suite) statusChange(name, from, to, approver string) error {
	return s.change(name, "status", from, to, approver)
}
func (s *Suite) namedChange(name, area, from, to, approver string) error {
	areas := map[string]string{"Zuweisung": "assignment", "Delegation": "delegation"}
	return s.change(name, areas[area], from, to, approver)
}

func (s *Suite) change(name, area, from, to, approver string) error {
	before, err := s.rules(name)
	if err != nil {
		return err
	}
	org, _ := s.organization(name)
	change := map[string]any{"operation": "allow", "area": area, "from": from, "to": to, "approver": approver, "revision": before["revision"]}
	path := "/api/organisationen/" + url.PathEscape(org.ID) + "/arbeitsregeln"
	if err := s.call("POST", path+"/vorschau", change); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Vorschau: %d %s", s.response.Status, s.response.Body)
	}
	var preview map[string]any
	if err := json.Unmarshal(s.response.Body, &preview); err != nil {
		return err
	}
	s.lastPreview = preview
	s.lastArea = area
	s.baseline = before
	s.lastOrganization = name
	if err := s.call("PUT", path, change); err != nil {
		return err
	}
	if s.response.Status != 200 {
		return fmt.Errorf("Speichern: %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) previewArea(label string) error {
	areas := map[string]string{"Aufgabenstatus": "status", "Zuweisung": "assignment", "Delegation": "delegation"}
	if s.lastPreview["area"] != areas[label] {
		return fmt.Errorf("falscher Vorschau-Bereich: %v", s.lastPreview)
	}
	flows, ok := s.lastPreview["affected_flows"].([]any)
	if !ok || len(flows) == 0 {
		return fmt.Errorf("betroffene Abläufe fehlen: %v", s.lastPreview)
	}
	for _, flow := range flows {
		description, _ := flow.(string)
		if !strings.HasPrefix(description, label+":") {
			return fmt.Errorf("Vorschau nennt anderen Bereich: %q", description)
		}
	}
	return nil
}

func (s *Suite) savedArea(label, name string) error {
	areas := map[string]string{"Aufgabenstatus": "status", "Zuweisung": "assignment", "Delegation": "delegation"}
	if s.lastArea != areas[label] {
		return fmt.Errorf("gespeicherter Bereich %s, erwartet %s", s.lastArea, areas[label])
	}
	current, err := s.rules(name)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(current["rules"], s.baseline["rules"]) {
		return fmt.Errorf("Regeln unverändert nach Speichern")
	}
	for _, area := range []string{"status", "assignment", "delegation"} {
		if area != s.lastArea && !reflect.DeepEqual(s.areaRules(current, area), s.areaRules(s.baseline, area)) {
			return fmt.Errorf("anderer Regelbereich %s verändert", area)
		}
	}
	return nil
}

func (s *Suite) otherAreasStable() error {
	current, err := s.rules(s.lastOrganization)
	if err != nil {
		return err
	}
	for _, area := range []string{"assignment", "delegation"} {
		if !reflect.DeepEqual(s.areaRules(s.baseline, area), s.areaRules(current, area)) {
			return fmt.Errorf("%s wurde mit verändert", area)
		}
	}
	return nil
}

func (s *Suite) areaRules(snapshot map[string]any, area string) []any {
	rules, _ := snapshot["rules"].([]any)
	selected := []any{}
	for _, item := range rules {
		rule, ok := item.(map[string]any)
		if ok && rule["area"] == area {
			selected = append(selected, item)
		}
	}
	return selected
}

func (s *Suite) rulesStable(name string) error {
	current, err := s.rules(name)
	if err != nil {
		return err
	}
	initial := s.initialRules[name]
	if initial == nil {
		return fmt.Errorf("Ausgangsregeln für %s fehlen", name)
	}
	if !reflect.DeepEqual(current["rules"], initial["rules"]) {
		return fmt.Errorf("Regeln von %s unerwartet geändert", name)
	}
	return nil
}

func (s *Suite) invalidChange(name, kind string) error {
	before, err := s.rules(name)
	if err != nil {
		return err
	}
	s.baseline = before
	s.lastOrganization = name
	org, _ := s.organization(name)
	change := map[string]any{"operation": "allow", "area": "status", "from": "todo", "to": "in_progress", "approver": "betreiber", "revision": before["revision"]}
	switch kind {
	case "unbekannter Aufgabenstatus":
		change["to"] = "unbekannt"
	case "fehlende zuständige Freigabe":
		change["approver"] = ""
	case "Rechteausweitung":
		change["area"] = "delegation"
		change["from"] = "requested"
		change["to"] = "approved"
		change["approver"] = "agent"
	case "Workflow-DSL":
		change["to"] = "in_progress => allow all"
	}
	path := "/api/organisationen/" + url.PathEscape(org.ID) + "/arbeitsregeln"
	if err := s.call("POST", path+"/vorschau", change); err != nil {
		return err
	}
	preview := s.response
	if err := s.call("PUT", path, change); err != nil {
		return err
	}
	if preview.Status < 400 || s.response.Status < 400 {
		return fmt.Errorf("ungültige Regel akzeptiert: Vorschau %d, Save %d", preview.Status, s.response.Status)
	}
	return nil
}

func (s *Suite) invalidReason() error {
	if s.response.Status < 400 || s.response.Status >= 500 {
		return fmt.Errorf("Ablehnung: %d %s", s.response.Status, s.response.Body)
	}
	if !strings.Contains(string(s.response.Body), "error") && !strings.Contains(string(s.response.Body), "Fehler") {
		return fmt.Errorf("Ablehnungsgrund fehlt: %s", s.response.Body)
	}
	return nil
}
