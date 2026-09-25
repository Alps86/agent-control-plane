package codexprofil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	sc.Step(`^ich starte die Anwendung mit einer neuen SQLite-Datenbank und einem privaten Datenverzeichnis$`, s.fresh)
	sc.Step(`^die Organisation "Nordstern" mit dem Codex-CLI-Agenten "Kai" besteht$`, s.createNordsternAndKai)
	sc.Step(`^ich Kais Codex-Profil über die öffentliche API lese$`, s.readProfile)
	sc.Step(`^ist der Arbeitsbereich ausschließlich serverseitig abgeleitet und gehört zu "Nordstern" und "Kai"$`, s.derivedWorkspace)
	sc.Step(`^Arbeitsbereich und Schreiben sind deaktiviert$`, s.bothDisabled)
	sc.Step(`^das Profil enthält keine vom Client gesetzten Pfade oder direkten Systemfähigkeiten$`, s.noDirectAccess)
	sc.Step(`^Kai ist nicht startbereit$`, s.notReady)
	sc.Step(`^ich Kais Codex-Profil mit aktiviertem Arbeitsbereich und deaktiviertem Schreiben über die öffentliche API speichere$`, s.enableWorkspace)
	sc.Step(`^zeigt das gespeicherte Profil den aktivierten Arbeitsbereich und deaktiviertes Schreiben$`, s.workspaceOnly)
	sc.Step(`^ich die Anwendung mit derselben SQLite-Datenbank neu starte$`, s.restart)
	sc.Step(`^zeigt Kais Codex-Profil weiterhin den aktivierten Arbeitsbereich und deaktiviertes Schreiben$`, s.persistedWorkspace)
	sc.Step(`^Kai ist wegen des fehlenden CLI-Durchsetzungsnachweises nicht startbereit$`, s.unverified)
	sc.Step(`^ich Kais Codex-Profil mit deaktiviertem Arbeitsbereich und aktiviertem Schreiben über die öffentliche API speichere$`, s.writeWithoutWorkspace)
	sc.Step(`^wird die Konfiguration am Feld "write_enabled" ohne Änderung abgewiesen$`, s.writeFieldDenied)
	sc.Step(`^ich Kais Codex-Profil mit dem zusätzlichen JSON-Feld (.+) über die öffentliche API speichere$`, s.injectField)
	sc.Step(`^wird die Konfiguration ohne Änderung abgewiesen$`, s.invalidNoMutation)
	sc.Step(`^ich Kais Codex-Profil mit gültigen Schaltern und mehr als einem MiB nachfolgendem Leerraum speichere$`, s.oversizedJSON)
	sc.Step(`^wird der übergroße JSON-Körper mit HTTP 400 abgewiesen$`, s.oversizedDenied)
	sc.Step(`^die Organisation "Südlicht" besteht$`, s.createSuedlicht)
	sc.Step(`^ich Kais Codex-Profil über die URL von "Südlicht" lese$`, s.readForeign)
	sc.Step(`^ich Kais Codex-Profil über die URL von "Südlicht" ändere$`, s.writeForeign)
	sc.Step(`^entspricht die datenfreie Ablehnung der Antwort auf eine unbekannte Agentenkennung$`, s.sameUnknown)
	sc.Step(`^Kais Profil in "Nordstern" ist unverändert$`, s.bothDisabled)
	sc.Step(`^eine uneindeutige Betreiberidentität Kais Codex-Profil ändert$`, s.ambiguousIdentity)
	sc.Step(`^ein fremder Browser-Ursprung Kais Codex-Profil ändert$`, s.foreignOrigin)
	sc.Step(`^antwortet die Anwendung mit HTTP 403 ohne Profildaten$`, s.forbiddenNoData)
	sc.Step(`^der serverseitig abgeleitete Arbeitsbereich für Kai enthält einen Symlink nach außerhalb$`, s.symlinkEscape)
	sc.Step(`^wird der Arbeitsbereich wegen verletzter Integrität ohne Schreibwirkung abgewiesen$`, s.integrityDenied)
	sc.Step(`^ich Kais Codex-Profil mit aktiviertem Arbeitsbereich und aktiviertem Schreiben über die öffentliche API speichere$`, s.enableBoth)
	sc.Step(`^ich Kais Bereitschaft über die öffentliche Agenten-API abfrage$`, s.readReadiness)
	sc.Step(`^ist Kai wegen des fehlenden CLI-Durchsetzungsnachweises nicht startbereit$`, s.unverified)
	sc.Step(`^ich Kai über die öffentliche Agenten-API starte$`, s.startAgent)
	sc.Step(`^wird der Start mit HTTP 409 ohne Laufkennung verweigert$`, s.startDenied)
	s.registerBrowser(sc)
	s.registerRuntime(sc)
}

func (s *Suite) createNordsternAndKai() error {
	if err := s.createOrganization("Nordstern"); err != nil {
		return err
	}
	return s.createAgent()
}

func (s *Suite) createSuedlicht() error { return s.createOrganization("Südlicht") }
func (s *Suite) readProfile() error     { return s.request("GET", s.profilePath(), "", "") }

func (s *Suite) profile() (Profile, error) {
	var profile Profile
	if s.response.Status != 200 {
		return profile, fmt.Errorf("Profil HTTP %d: %s", s.response.Status, s.response.Body)
	}
	err := json.Unmarshal(s.response.Body, &profile)
	return profile, err
}

func (s *Suite) derivedWorkspace() error {
	profile, err := s.profile()
	if err != nil {
		return err
	}
	if profile.AgentID != s.agentID || profile.OrganizationID != s.orgID {
		return fmt.Errorf("abgeleiteter Arbeitsbereich fehlt: %s", s.response.Body)
	}
	if strings.Contains(string(s.response.Body), filepath.Dir(s.dbPath)) {
		return fmt.Errorf("privater Hostpfad offengelegt: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) bothDisabled() error {
	if err := s.readProfile(); err != nil {
		return err
	}
	profile, err := s.profile()
	if err != nil {
		return err
	}
	if profile.WorkspaceEnabled || profile.WriteEnabled {
		return fmt.Errorf("restriktiver Default verletzt: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) noDirectAccess() error {
	content := string(s.response.Body)
	for _, field := range []string{"workspace_path", "filesystem_write", "shell", "terminal", "interpreter"} {
		if strings.Contains(content, field) {
			return fmt.Errorf("direkter Zugriff im Profil: %s", content)
		}
	}
	return nil
}

func (s *Suite) notReady() error {
	if err := s.readReadiness(); err != nil {
		return err
	}
	var readiness Readiness
	if err := json.Unmarshal(s.response.Body, &readiness); err != nil {
		return err
	}
	if readiness.Ready {
		return fmt.Errorf("CLI fälschlich startbereit: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) putProfile(workspace, write bool) error {
	body, _ := json.Marshal(map[string]bool{"workspace_enabled": workspace, "write_enabled": write})
	return s.request("PUT", s.profilePath(), string(body), "")
}

func (s *Suite) enableWorkspace() error       { return s.putProfile(true, false) }
func (s *Suite) enableBoth() error            { return s.putProfile(true, true) }
func (s *Suite) writeWithoutWorkspace() error { return s.putProfile(false, true) }

func (s *Suite) workspaceOnly() error {
	profile, err := s.profile()
	if err != nil {
		return err
	}
	if !profile.WorkspaceEnabled || profile.WriteEnabled {
		return fmt.Errorf("Profilzustand falsch: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) persistedWorkspace() error {
	if err := s.readProfile(); err != nil {
		return err
	}
	return s.workspaceOnly()
}

func (s *Suite) unverified() error {
	if err := s.readReadiness(); err != nil {
		return err
	}
	var readiness Readiness
	if err := json.Unmarshal(s.response.Body, &readiness); err != nil {
		return err
	}
	if readiness.Ready || readiness.Code != "codex_cli_enforcement_unverified" {
		return fmt.Errorf("CLI-Grenze nicht gesperrt: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) writeFieldDenied() error {
	if s.response.Status != 422 || !strings.Contains(string(s.response.Body), "write_enabled") {
		return fmt.Errorf("Schreibrecht ohne Bereich: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) injectField(field string) error {
	body := `{"workspace_enabled":true,"write_enabled":false,` + field + `}`
	return s.request("PUT", s.profilePath(), body, "")
}

func (s *Suite) invalidNoMutation() error {
	if s.response.Status != 400 && s.response.Status != 422 {
		return fmt.Errorf("eingeschmuggeltes Feld angenommen: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) oversizedJSON() error {
	body := `{"workspace_enabled":true,"write_enabled":true}` + strings.Repeat(" ", (1<<20)+1)
	return s.request("PUT", s.profilePath(), body, "")
}

func (s *Suite) oversizedDenied() error {
	if s.response.Status != 400 || !strings.Contains(string(s.response.Body), "invalid_json") {
		return fmt.Errorf("übergroßes JSON angenommen: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) foreignPath() string {
	return "/api/organisationen/" + s.otherOrgID + "/agenten/" + s.agentID + "/codex-profil"
}

func (s *Suite) readForeign() error { return s.request("GET", s.foreignPath(), "", "") }
func (s *Suite) writeForeign() error {
	return s.request("PUT", s.foreignPath(), `{"workspace_enabled":true,"write_enabled":true}`, "")
}

func (s *Suite) sameUnknown() error {
	if s.response.Status != 404 {
		return fmt.Errorf("Fremdprofil HTTP %d: %s", s.response.Status, s.response.Body)
	}
	foreign := string(s.response.Body)
	path := "/api/organisationen/" + s.otherOrgID + "/agenten/unbekannt/codex-profil"
	method, body := s.response.Method, ""
	if method == "PUT" {
		method, body = "PUT", `{"workspace_enabled":true,"write_enabled":true}`
	}
	if err := s.request(method, path, body, ""); err != nil {
		return err
	}
	if s.response.Status != 404 || string(s.response.Body) != foreign {
		return fmt.Errorf("abweichende Ablehnung: %s / %s", foreign, s.response.Body)
	}
	return nil
}

func (s *Suite) foreignOrigin() error {
	return s.request("PUT", s.profilePath(), `{"workspace_enabled":true,"write_enabled":true}`, "https://fremd.example")
}

func (s *Suite) forbiddenNoData() error {
	if s.response.Status != http.StatusForbidden || strings.Contains(string(s.response.Body), s.agentID) {
		return fmt.Errorf("Zugriff nicht datenfrei gesperrt: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return nil
}

func (s *Suite) symlinkEscape() error {
	workspace := filepath.Join(filepath.Dir(s.dbPath), "codex-workspaces", s.orgID, s.agentID)
	if err := os.MkdirAll(filepath.Dir(workspace), 0700); err != nil {
		return err
	}
	return os.Symlink(s.t.TempDir(), workspace)
}

func (s *Suite) integrityDenied() error {
	if s.response.Status != 409 || !strings.Contains(string(s.response.Body), "profile_conflict") {
		return fmt.Errorf("Symlink nicht verweigert: HTTP %d %s", s.response.Status, s.response.Body)
	}
	return s.bothDisabled()
}

func (s *Suite) readReadiness() error {
	return s.request("GET", "/api/organisationen/"+s.orgID+"/agenten/"+s.agentID+"/bereitschaft", "", "")
}

func (s *Suite) startAgent() error {
	return s.request("POST", "/api/organisationen/"+s.orgID+"/agenten/"+s.agentID+"/start", "", "")
}

func (s *Suite) startDenied() error {
	content := string(s.response.Body)
	if s.response.Status != 409 || strings.Contains(content, "run_id") || strings.Contains(content, "lauf_id") {
		return fmt.Errorf("CLI-Start nicht sicher verweigert: HTTP %d %s", s.response.Status, content)
	}
	return nil
}
