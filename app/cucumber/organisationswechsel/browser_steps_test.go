package organisationswechsel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (s *Suite) registerUISteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich öffne Story 31 mit den Organisationen "([^"]+)" und "([^"]+)" im Browser$`, s.openBrowserFixture)
	sc.Step(`^ich in der Organisationsübersicht nach "([^"]+)" suche$`, s.browserSearch)
	sc.Step(`^ich den regulären Einstieg der Anwendung öffne$`, s.browserEntry)
	sc.Step(`^sehe ich die Aktion "Organisation wechseln"$`, s.browserSwitchAction)
	sc.Step(`^ich die Organisationswahl über diese Aktion öffne$`, s.browserOpenChooser)
	sc.Step(`^sehe ich im Organisationsdetail die Aktion "Arbeitsregeln bearbeiten"$`, s.browserRuleAction)
	sc.Step(`^ich die Arbeitsregeln über diese Aktion öffne$`, s.browserOpenRuleAction)
	sc.Step(`^sehe ich "([^"]+)" und nicht "([^"]+)" in den Suchergebnissen$`, s.browserSearchResult)
	sc.Step(`^ich "([^"]+)" auswähle$`, s.browserSelect)
	sc.Step(`^zeigt die Oberfläche "([^"]+)" als gewählte Organisation$`, s.browserSelected)
	sc.Step(`^ich sehe Namen und Beschreibung im Detail$`, s.browserDetail)
	sc.Step(`^ich die Detail-URL neu lade$`, s.browserReload)
	sc.Step(`^bleibt "([^"]+)" als gewählte Organisation sichtbar$`, s.browserSelected)
	sc.Step(`^ich die Detail-URL von "([^"]+)" direkt im Browser öffne$`, s.browserForeign)
	sc.Step(`^sehe ich keine Daten von "([^"]+)"$`, s.browserNoData)
	sc.Step(`^die Oberfläche zeigt denselben datenfreien Zustand wie für eine unbekannte Kennung$`, s.browserUnknown)
	sc.Step(`^ich die Arbeitsregeln von "([^"]+)" öffne$`, s.browserRules)
	sc.Step(`^sehe ich getrennte Bereiche für Aufgabenstatus, Zuweisung und Delegation$`, s.browserAreas)
	sc.Step(`^ich einen zulässigen Aufgabenstatus-Übergang und die zuständige Freigabe auswähle$`, s.browserValidRule)
	sc.Step(`^ich die Vorschau öffne$`, s.browserPreview)
	sc.Step(`^sehe ich betroffene Aufgabenstatus-Abläufe und keine Änderung der übrigen Bereiche$`, s.browserPreviewText)
	sc.Step(`^ich die Regel speichere und die Seite neu lade$`, s.browserSaveReload)
	sc.Step(`^sehe ich den gespeicherten Übergang bei "([^"]+)"$`, s.browserSavedRule)
	sc.Step(`^bei "([^"]+)" bleibt die bisherige Regel sichtbar$`, s.browserOtherStable)
	sc.Step(`^ich den gespeicherten Übergang widerrufe$`, s.browserRevoke)
	sc.Step(`^sehe ich den Übergang bei "([^"]+)" nicht mehr als freigegeben$`, s.browserRevoked)
	sc.Step(`^die Bereiche Zuweisung und Delegation bleiben sichtbar unverändert$`, s.browserOtherAreas)
	sc.Step(`^ich bei "([^"]+)" (einen ungültigen Übergang|eine Rechteausweitung) vorschauprüfe$`, s.browserInvalidRule)
	sc.Step(`^sehe ich einen konkreten Ablehnungsgrund$`, s.browserError)
	sc.Step(`^ich kann die abgewiesene Regel nicht speichern$`, s.browserNoSave)
	sc.Step(`^nach erneutem Laden bleiben die bisherigen Regeln sichtbar$`, s.browserRulesStable)
}

func (s *Suite) openBrowserFixture(a, b string) error {
	if err := s.fresh(); err != nil {
		return err
	}
	if err := s.createTwo(a, b); err != nil {
		return err
	}
	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}
	s.browser = exec.Command("node", script)
	s.browser.Stderr = os.Stderr
	in, err := s.browser.StdinPipe()
	if err != nil {
		return err
	}
	s.browserInput = in
	out, err := s.browser.StdoutPipe()
	if err != nil {
		return err
	}
	s.stdin = bufio.NewWriter(in)
	s.stdout = bufio.NewScanner(out)
	if err := s.browser.Start(); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"url": "http://" + s.address + "/organisationswechsel", "mode": "Organisation wählen"})
}

func (s *Suite) browserCommand(command map[string]any) error {
	data, err := json.Marshal(command)
	if err != nil {
		return err
	}
	if _, err := s.stdin.Write(append(data, '\n')); err != nil {
		return err
	}
	if err := s.stdin.Flush(); err != nil {
		return err
	}
	if !s.stdout.Scan() {
		return fmt.Errorf("Browser ohne Antwort: %v", s.stdout.Err())
	}
	var reply struct {
		OK    bool        `json:"ok"`
		Error string      `json:"error"`
		Page  BrowserPage `json:"page"`
	}
	if err := json.Unmarshal(s.stdout.Bytes(), &reply); err != nil {
		return err
	}
	if !reply.OK {
		return fmt.Errorf("Browser: %s", reply.Error)
	}
	s.page = reply.Page
	return nil
}

func (s *Suite) browserSearch(q string) error {
	if err := s.browserCommand(map[string]any{"fields": map[string]string{"q": q}}); err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"submit": "form[role=search]", "mode": "Organisation wählen", "urlContains": "q=" + url.QueryEscape(q)})
}

func (s *Suite) browserEntry() error {
	return s.browserCommand(map[string]any{"url": "http://" + s.address + "/", "mode": "Organisationsübersicht"})
}

func (s *Suite) browserSwitchAction() error {
	if !strings.Contains(s.page.Text, "Organisation wechseln") {
		return fmt.Errorf("Wechselaktion fehlt: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserOpenChooser() error {
	return s.browserCommand(map[string]any{"click": "a[href=\"/organisationswechsel\"]", "mode": "Organisation wählen"})
}

func (s *Suite) browserRuleAction() error {
	if !strings.Contains(s.page.Text, "Arbeitsregeln bearbeiten") {
		return fmt.Errorf("Regellink fehlt: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserOpenRuleAction() error {
	org, err := s.organization("Nordstern")
	if err != nil {
		return err
	}
	selector := "a[href=\"/organisationen/" + org.ID + "/arbeitsregeln\"]"
	return s.browserCommand(map[string]any{"click": selector, "mode": "Arbeitsregeln"})
}

func (s *Suite) browserSearchResult(want, absent string) error {
	if !strings.Contains(s.page.Text, want) || strings.Contains(s.page.Text, absent) {
		return fmt.Errorf("Suchergebnis: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserSelect(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	selector := "form[action=\"/organisationswechsel/" + org.ID + "\"]"
	return s.browserCommand(map[string]any{"submit": selector, "mode": name, "urlContains": "/organisationen/" + org.ID})
}

func (s *Suite) browserSelected(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	if !strings.Contains(s.page.URL, "/organisationen/"+org.ID) || !strings.Contains(s.page.Text, name) {
		return fmt.Errorf("Organisation nicht gewählt: %v", s.page)
	}
	return s.browserSelectionMarker(name)
}

func (s *Suite) browserSelectionMarker(name string) error {
	previous := s.page.URL
	if err := s.browserCommand(map[string]any{"url": "http://" + s.address + "/organisationswechsel", "mode": "Organisation wählen"}); err != nil {
		return err
	}
	if len(s.page.SelectedNames) != 1 || s.page.SelectedNames[0] != name {
		return fmt.Errorf("Auswahlmarkierung falsch: %v", s.page.SelectedNames)
	}
	if err := s.browserCommand(map[string]any{"url": previous, "mode": name}); err != nil {
		return err
	}
	return nil
}

func (s *Suite) browserDetail() error {
	for _, org := range s.organizations {
		if strings.Contains(s.page.URL, org.ID) {
			if strings.Contains(s.page.Text, org.Name) && strings.Contains(s.page.Text, org.Description) {
				return nil
			}
			return fmt.Errorf("Detail unvollständig: %s", s.page.Text)
		}
	}
	return fmt.Errorf("kein Detail: %s", s.page.URL)
}

func (s *Suite) browserReload() error {
	return s.browserCommand(map[string]any{"url": s.page.URL, "mode": s.page.Heading})
}

func (s *Suite) browserForeign(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	if s.foreignServer == "" {
		return fmt.Errorf("fremde HTTP-Testinstanz fehlt")
	}
	return s.browserCommand(map[string]any{"url": s.foreignServer + "/organisationen/" + url.PathEscape(org.ID)})
}

func (s *Suite) browserNoData(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	if strings.Contains(s.page.Text, name) || strings.Contains(s.page.Text, org.Description) {
		return fmt.Errorf("fremde Daten im Browser: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserUnknown() error {
	previous := s.page.Text
	if err := s.browserCommand(map[string]any{"url": s.foreignServer + "/organisationen/unbekannt"}); err != nil {
		return err
	}
	if s.page.Text != previous {
		return fmt.Errorf("fremd/unbekannt verschieden: %q / %q", previous, s.page.Text)
	}
	return nil
}

func (s *Suite) browserRules(name string) error {
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"url": "http://" + s.address + "/organisationen/" + url.PathEscape(org.ID) + "/arbeitsregeln", "mode": "Arbeitsregeln"})
}

func (s *Suite) browserAreas() error {
	for _, name := range []string{"Aufgabenstatus", "Zuweisung", "Delegation"} {
		if !strings.Contains(s.page.Text, name) {
			return fmt.Errorf("Bereich %s fehlt: %s", name, s.page.Text)
		}
	}
	return nil
}

func (s *Suite) browserValidRule() error {
	return s.browserCommand(map[string]any{"fields": map[string]string{"area": "status", "from": "todo", "to": "in_progress", "approver": "betreiber"}})
}
func (s *Suite) browserPreview() error {
	org, err := s.organization("Nordstern")
	if err != nil {
		return err
	}
	return s.browserCommand(map[string]any{"submit": "form[action=\"/organisationen/" + org.ID + "/arbeitsregeln/vorschau\"]", "mode": "Arbeitsregeln", "textContains": "Betroffene Abläufe"})
}
func (s *Suite) browserPreviewText() error {
	if !strings.Contains(s.page.Text, "Betroffene Abläufe") || !strings.Contains(s.page.Text, "Aufgabenstatus") || !strings.Contains(s.page.Text, "todo") || !strings.Contains(s.page.Text, "in_progress") {
		return fmt.Errorf("Vorschau unvollständig: %s", s.page.Text)
	}
	return nil
}
func (s *Suite) browserSaveReload() error {
	org, err := s.organization("Nordstern")
	if err != nil {
		return err
	}
	if err := s.browserCommand(map[string]any{"submit": "form[action=\"/organisationen/" + org.ID + "/arbeitsregeln\"]", "mode": "Arbeitsregeln", "textAbsent": "Betroffene Abläufe"}); err != nil {
		return err
	}
	return s.browserReload()
}
func (s *Suite) browserSavedRule(name string) error {
	if !strings.Contains(s.page.Text, name) || !strings.Contains(s.page.Text, "todo → in_progress") {
		return fmt.Errorf("Regel nicht sichtbar: %s", s.page.Text)
	}
	return nil
}
func (s *Suite) browserOtherStable(name string) error {
	if err := s.browserRules(name); err != nil {
		return err
	}
	if strings.Contains(s.page.Text, "todo → in_progress") {
		return fmt.Errorf("Regel in fremder Organisation sichtbar")
	}
	return nil
}

func (s *Suite) browserRevoke() error {
	if !strings.Contains(s.page.Text, "todo → in_progress") {
		return fmt.Errorf("gespeicherte Regel vor Browser-Widerruf fehlt")
	}
	org, err := s.organization("Nordstern")
	if err != nil {
		return err
	}
	previewSelector := "form[action=\"/organisationen/" + org.ID + "/arbeitsregeln/vorschau\"]:has(input[name=operation][value=revoke])"
	if err := s.browserCommand(map[string]any{"submit": previewSelector, "mode": "Arbeitsregeln", "textContains": "Widerruf: betroffene Abläufe"}); err != nil {
		return err
	}
	confirmSelector := "form[action=\"/organisationen/" + org.ID + "/arbeitsregeln/widerruf\"]"
	return s.browserCommand(map[string]any{"submit": confirmSelector, "mode": "Arbeitsregeln", "textAbsent": "Widerruf: betroffene Abläufe"})
}

func (s *Suite) browserRevoked(name string) error {
	if !strings.Contains(s.page.Text, name) {
		return fmt.Errorf("Organisation im Widerruf fehlt")
	}
	if strings.Contains(s.page.Text, "todo → in_progress") {
		return fmt.Errorf("widerrufene Regel noch sichtbar: %s", s.page.Text)
	}
	return nil
}

func (s *Suite) browserOtherAreas() error {
	if !strings.Contains(s.page.Text, "Zuweisung") || !strings.Contains(s.page.Text, "Delegation") {
		return fmt.Errorf("andere Bereiche fehlen")
	}
	if strings.Count(s.page.Text, "Keine Übergänge freigegeben.") < 3 {
		return fmt.Errorf("andere Bereiche nach Widerruf geändert: %s", s.page.Text)
	}
	return nil
}
func (s *Suite) browserInvalidRule(name, kind string) error {
	if err := s.browserRules(name); err != nil {
		return err
	}
	fields := map[string]string{"area": "status", "from": "todo", "to": "unbekannt", "approver": "betreiber"}
	if kind == "eine Rechteausweitung" {
		fields["to"] = "in_progress"
		fields["approver"] = "agent"
	}
	command := map[string]any{"fields": fields}
	if kind == "eine Rechteausweitung" {
		command["injectOption"] = map[string]string{"field": "approver", "value": "agent"}
	}
	if err := s.browserCommand(command); err != nil {
		return err
	}
	org, err := s.organization(name)
	if err != nil {
		return err
	}
	expected := "nicht zulässig"
	if kind == "eine Rechteausweitung" {
		expected = "Rechte"
	}
	return s.browserCommand(map[string]any{"submit": "form[action=\"/organisationen/" + org.ID + "/arbeitsregeln/vorschau\"]", "mode": "Arbeitsregeln", "textContains": expected})
}
func (s *Suite) browserError() error {
	if !strings.Contains(s.page.Text, "ungültig") && !strings.Contains(s.page.Text, "nicht erlaubt") && !strings.Contains(s.page.Text, "Unzulässig") && !strings.Contains(s.page.Text, "nicht zulässig") && !strings.Contains(s.page.Text, "Rechte über die erlaubte Grenze") {
		return fmt.Errorf("Ablehnungsgrund fehlt: %s", s.page.Text)
	}
	return nil
}
func (s *Suite) browserNoSave() error {
	if strings.Contains(s.page.Text, "Regel speichern") {
		return fmt.Errorf("abgewiesene Regel speicherbar")
	}
	return nil
}
func (s *Suite) browserRulesStable() error {
	if err := s.browserRules("Nordstern"); err != nil {
		return err
	}
	if strings.Contains(s.page.Text, "todo → unbekannt") || strings.Contains(s.page.Text, "todo → in_progress") {
		return fmt.Errorf("ungültige Regel persistiert")
	}
	return nil
}
