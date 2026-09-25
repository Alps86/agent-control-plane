package projektort

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

type location struct {
	Configured bool    `json:"configured"`
	Kind       string  `json:"kind"`
	Path       *string `json:"path"`
	StartReady bool    `json:"start_ready"`
	Ready      bool    `json:"ready"`
	Code       string  `json:"code"`
	Reason     string  `json:"reason"`
}

func (s *suite) steps(sc *godog.ScenarioContext) {
	sc.After(s.after)
	sc.Step(`^ich starte die Anwendung mit einer neuen SQLite-Datenbank und privatem Datenverzeichnis$`, s.fresh)
	sc.Step(`^ich lege die Organisation "([^"]*)" und das Projekt "([^"]*)" über die öffentliche API an$`, s.createFixture)
	sc.Step(`^ich lege die Organisation "([^"]*)" mit dem Projekt "([^"]*)" über die öffentliche API an$`, s.createFixture)
	sc.Step(`^ich den Ausführungsort von "([^"]*)" über die öffentliche API abrufe$`, s.getLocation)
	sc.Step(`^zeigt die Antwort configured=false, path=null und code=project_location_missing$`, s.missingLocation)
	sc.Step(`^ich den privaten Ausführungsort von "([^"]*)" über die öffentliche API aktiviere$`, s.enableLocation)
	sc.Step(`^ich habe den privaten Ausführungsort von "([^"]*)" über die öffentliche API aktiviert$`, s.enableLocation)
	sc.Step(`^ist der Ausführungsort für "([^"]*)" aktiv und gehört nur zu "([^"]*)"$`, s.activeLocation)
	sc.Step(`^die öffentliche Startprüfung für "([^"]*)" bestätigt den freigegebenen privaten Ort als bereit$`, s.readyStartCheck)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte$`, s.restart)
	sc.Step(`^bleibt der Ausführungsort von "([^"]*)" aktiv$`, s.activeAfterRestart)
	sc.Step(`^ich die öffentliche Startprüfung für "([^"]*)" anfordere$`, s.startCheck)
	sc.Step(`^wird die Startvorbereitung ohne Laufkennung wegen fehlendem Ausführungsort verweigert$`, s.missingStartCheck)
	sc.Step(`^die Antwort erklärt, wie ich den privaten Projektort aktiviere$`, s.explainsCorrection)
	sc.Step(`^ich für "([^"]*)" den Ausführungsort mit (.*) über die öffentliche API speichere$`, s.injectLocation)
	sc.Step(`^wird die Pfadfreigabe ohne Änderung des Ausführungsorts abgewiesen$`, s.rejectedInjection)
	sc.Step(`^die öffentliche Startprüfung für "([^"]*)" erhält keinen fremden Pfad$`, s.noForeignPath)
	sc.Step(`^ich den Ausführungsort von "([^"]*)" über "([^"]*)" abrufe$`, s.getThroughWrongOrg)
	sc.Step(`^erhalte ich eine datenfreie Verweigerung ohne Pfadangabe$`, s.opaqueNotFound)
	sc.Step(`^der private Ort von "([^"]*)" enthält einen Symlink nach außerhalb$`, s.insertSymlink)
	sc.Step(`^der private Ort von "([^"]*)" enthält einen Hardlink auf eine Datei außerhalb$`, s.insertHardlink)
	sc.Step(`^wird die Startvorbereitung wegen verletzter Pfadintegrität ohne Pfadübergabe abgewiesen$`, s.invalidStartCheck)
	s.registerBrowserSteps(sc)
}

func (s *suite) createFixture(org, project string) error {
	if err := s.call(http.MethodPost, "/api/organisationen", map[string]string{"name": org, "description": "Testorganisation"}); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var created record
	if err := json.Unmarshal(s.response.body, &created); err != nil {
		return err
	}
	if created.ID == "" {
		return fmt.Errorf("Organisationskennung fehlt: %s", s.response.body)
	}
	s.organizations[org] = created.ID
	if err := s.call(http.MethodPost, "/api/organisationen/"+created.ID+"/ziele", map[string]string{"name": "Projektziel"}); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var goal record
	if err := json.Unmarshal(s.response.body, &goal); err != nil {
		return err
	}
	if goal.ID == "" {
		return fmt.Errorf("Zielkennung fehlt: %s", s.response.body)
	}
	if err := s.call(http.MethodPost, "/api/organisationen/"+created.ID+"/projekte", map[string]string{"name": project, "description": "Testprojekt", "goal_id": goal.ID}); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var p record
	if err := json.Unmarshal(s.response.body, &p); err != nil {
		return err
	}
	if p.ID == "" {
		return fmt.Errorf("Projektkennung fehlt: %s", s.response.body)
	}
	s.projects[s.key(org, project)] = p.ID
	s.lastOrg, s.lastProject = org, project
	return nil
}

func (s *suite) getLocation(project string) error {
	return s.call(http.MethodGet, s.path(s.lastOrg, project), nil)
}
func (s *suite) decodeLocation() (location, error) {
	var l location
	err := json.Unmarshal(s.response.body, &l)
	return l, err
}
func (s *suite) missingLocation() error {
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if l.Configured || l.Path != nil || l.StartReady || l.Code != "project_location_missing" {
		return fmt.Errorf("ungebundener Ort falsch: %+v", l)
	}
	return nil
}
func (s *suite) enableLocation(project string) error {
	return s.call(http.MethodPut, s.path(s.lastOrg, project), map[string]any{"kind": "managed_directory"})
}
func (s *suite) activeLocation(project, org string) error {
	if s.lastOrg != org {
		return fmt.Errorf("andere Organisation als erwartet: %s", s.lastOrg)
	}
	if err := s.call(http.MethodGet, s.path(org, project), nil); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if !l.Configured || l.Kind != "managed_directory" || l.Path == nil || *l.Path == "" {
		return fmt.Errorf("gebundener Ort falsch: %+v", l)
	}
	if !filepath.IsAbs(*l.Path) {
		return fmt.Errorf("Projektort nicht kanonisch: %q", *l.Path)
	}
	if filepath.Clean(*l.Path) != *l.Path {
		return fmt.Errorf("Projektort nicht bereinigt: %q", *l.Path)
	}
	expected := filepath.Join(filepath.Dir(s.dbPath), "project-workspaces", s.organizations[org], s.projects[s.key(org, project)])
	if *l.Path != expected {
		return fmt.Errorf("Projektort %q statt privatem Projektarbeitsordner %q", *l.Path, expected)
	}
	return nil
}
func (s *suite) activeAfterRestart(project string) error { return s.activeLocation(project, s.lastOrg) }
func (s *suite) startCheck(project string) error {
	return s.call(http.MethodPost, s.path(s.lastOrg, project)+"/startpruefung", map[string]any{})
}
func (s *suite) readyStartCheck(project string) error {
	if err := s.getLocation(project); err != nil {
		return err
	}
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if l.Path == nil {
		return fmt.Errorf("kein gebundener Pfad: %s", s.response.body)
	}
	expected := *l.Path
	if err := s.startCheck(project); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	checked, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if !checked.Ready || checked.Code != "project_location_ready" || checked.Path == nil || *checked.Path != expected {
		return fmt.Errorf("öffentliche Startprüfung nicht bereit oder Pfad inkonsistent: %+v; erwartet %q", checked, expected)
	}
	return nil
}
func (s *suite) missingStartCheck() error {
	if err := s.status(http.StatusConflict); err != nil {
		return err
	}
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if l.Ready || l.Path != nil || l.Code != "project_location_missing" || strings.Contains(string(s.response.body), "run_id") {
		return fmt.Errorf("fehlende Bindung nicht sicher verweigert: %s", s.response.body)
	}
	return nil
}
func (s *suite) explainsCorrection() error {
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(l.Reason), "aktiv") {
		return fmt.Errorf("Korrekturhinweis fehlt: %s", s.response.body)
	}
	return nil
}
func (s *suite) injectLocation(project, input string) error {
	request := map[string]any{"kind": "managed_directory"}
	switch input {
	case "einem Home-Pfadfeld":
		request["path"] = os.Getenv("HOME")
	case "einer Repository-URL":
		request["repository_url"] = "https://example.invalid/repo.git"
	case "einem Pfad mit Aufwärtssegment":
		request["path"] = "../other"
	case "einem unbekannten Feld":
		request["extra"] = "allowed"
	default:
		return fmt.Errorf("unbekannter Angriff: %s", input)
	}
	return s.call(http.MethodPut, s.path(s.lastOrg, project), request)
}
func (s *suite) rejectedInjection() error {
	if s.response.status < 400 || s.response.status >= 500 {
		return fmt.Errorf("Pfadfeld nicht abgewiesen: %d %s", s.response.status, s.response.body)
	}
	return s.missingAfterRejection()
}
func (s *suite) missingAfterRejection() error {
	if err := s.getLocation(s.lastProject); err != nil {
		return err
	}
	return s.missingLocation()
}
func (s *suite) noForeignPath(project string) error {
	if err := s.startCheck(project); err != nil {
		return err
	}
	return s.missingStartCheck()
}
func (s *suite) getThroughWrongOrg(project, org string) error {
	path := "/api/organisationen/" + s.organizations[org] + "/projekte/" + s.projects[s.key(s.lastOrg, project)] + "/ausfuehrungsort"
	return s.call(http.MethodGet, path, nil)
}
func (s *suite) opaqueNotFound() error {
	if err := s.status(http.StatusNotFound); err != nil {
		return err
	}
	if strings.Contains(string(s.response.body), "path") || strings.Contains(string(s.response.body), s.lastProject) {
		return fmt.Errorf("Fremdprojekt preisgegeben: %s", s.response.body)
	}
	return nil
}
func (s *suite) insertSymlink(project string) error {
	if err := s.getLocation(project); err != nil {
		return err
	}
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if l.Path == nil {
		return fmt.Errorf("privater Ort fehlt")
	}
	outside := s.t.TempDir()
	return os.Symlink(outside, filepath.Join(*l.Path, "escape"))
}

func (s *suite) insertHardlink(project string) error {
	if err := s.getLocation(project); err != nil {
		return err
	}
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if l.Path == nil {
		return fmt.Errorf("privater Ort fehlt")
	}
	outside := filepath.Join(s.t.TempDir(), "fremd.txt")
	if err := os.WriteFile(outside, []byte("fremder Inhalt"), 0600); err != nil {
		return err
	}
	return os.Link(outside, filepath.Join(*l.Path, "fremd.txt"))
}
func (s *suite) invalidStartCheck() error {
	if err := s.status(http.StatusConflict); err != nil {
		return err
	}
	l, err := s.decodeLocation()
	if err != nil {
		return err
	}
	if l.Ready || l.Path != nil || l.Code != "project_location_invalid" {
		return fmt.Errorf("verletzter Pfad wurde übergeben: %s", s.response.body)
	}
	return nil
}
