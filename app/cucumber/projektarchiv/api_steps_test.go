package projektarchiv

import (
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
	"strings"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	s.registerSetup(sc)
	s.registerActions(sc)
	s.registerAssertions(sc)
	s.registerBrowser(sc)
}

func (s *Suite) registerSetup(sc *godog.ScenarioContext) {
	sc.Step(`^die Organisation "([^"]+)" mit dem Projekt "([^"]+)" und dem Agenten "([^"]+)" besteht$`, s.projectAndAgent)
	sc.Step(`^die Organisation "([^"]+)" mit dem Projekt "([^"]+)" besteht$`, s.projectSetup)
	sc.Step(`^die Organisationen "([^"]+)" und "([^"]+)" bestehen$`, s.twoOrganizations)
	sc.Step(`^das Projekt "([^"]+)" besteht in "([^"]+)"$`, s.ensureNamedProject)
	sc.Step(`^der Agent "([^"]+)" besteht in "([^"]+)"$`, s.ensureAgent)
	sc.Step(`^die Aufgabe "([^"]+)" besteht im Projekt "([^"]+)"$`, s.ensureTask)
	sc.Step(`^"([^"]+)" ist in "([^"]+)" archiviert$`, s.archiveSetup)
	sc.Step(`^ein Lauf für "([^"]+)" ist reserviert$`, s.reserveRun)
	sc.Step(`^ein Lauf für "([^"]+)" ist mit Vorbereitungsfehler beendet$`, s.failedRun)
	sc.Step(`^der öffentliche Aktivitätsverlauf von "([^"]+)" enthält den Anlageeintrag$`, s.activitySetup)
}

func (s *Suite) registerActions(sc *godog.ScenarioContext) {
	sc.Step(`^ich "([^"]+)" in "([^"]+)" über HTTP archiviere$`, s.archive)
	sc.Step(`^ich "([^"]+)" in "([^"]+)" über HTTP wiederherstelle$`, s.restore)
	sc.Step(`^ich die Aufgabe "([^"]+)" direkt über die Aufgaben-URL von "([^"]+)" anlege$`, s.directTask)
	sc.Step(`^ich die neue Aufgabe "([^"]+)" im Projekt "([^"]+)" anlegen kann$`, s.canCreateTask)
	sc.Step(`^ich kann die neue Aufgabe "([^"]+)" im Projekt "([^"]+)" anlegen$`, s.canCreateTask)
	sc.Step(`^ich kann im archivierten Projekt keine neue Aufgabe anlegen$`, s.cannotCreateTask)
	sc.Step(`^ich den Server mit derselben SQLite-Datenbank neu starte$`, s.restart)
	sc.Step(`^ich "([^"]+)" über die Projekt-URL von "([^"]+)" zu archivieren versuche$`, s.archiveForeign)
	sc.Step(`^ich "([^"]+)" über die Projekt-URL von "([^"]+)" wiederherzustellen versuche$`, s.restoreForeign)
	sc.Step(`^ich ein unbekanntes Projekt über die Archiv-URL von "([^"]+)" zu archivieren versuche$`, s.archiveUnknown)
	sc.Step(`^ich über die öffentliche Lauf-Anwendungsgrenze einen Lauf für "([^"]+)" anfordere$`, s.attemptArchivedRun)
	sc.Step(`^ich "([^"]+)" archiviere und gleichzeitig die Aufgabe "([^"]+)" über die öffentliche HTTP-Grenze anlege$`, s.archiveAndCreateConcurrently)
}

func (s *Suite) registerAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^meldet der Server den Status "([^"]+)" für "([^"]+)"$`, s.projectStatus)
	sc.Step(`^die aktive Projektliste von "([^"]+)" enthält "([^"]+)" nicht$`, s.activeAbsent)
	sc.Step(`^die aktive Projektliste von "([^"]+)" enthält "([^"]+)" genau einmal$`, s.activeOnce)
	sc.Step(`^das Projektarchiv von "([^"]+)" enthält "([^"]+)" genau einmal$`, s.archiveOnce)
	sc.Step(`^das Projektarchiv von "([^"]+)" enthält "([^"]+)" nicht$`, s.archiveAbsent)
	sc.Step(`^die bestehende Aufgabe "([^"]+)" bleibt unter derselben Kennung erreichbar$`, s.taskPreserved)
	sc.Step(`^wird neue Arbeit mit einem konkreten Archivhinweis verweigert$`, s.archivedTaskRejected)
	sc.Step(`^in "([^"]+)" entsteht keine neue Aufgabe$`, s.noNewTask)
	sc.Step(`^erhalte ich eine datenfreie Ablehnung$`, s.notFound)
	sc.Step(`^"([^"]+)" bleibt in "([^"]+)" aktiv$`, s.remainsActive)
	sc.Step(`^"([^"]+)" bleibt im Archiv von "([^"]+)"$`, s.remainsArchived)
	sc.Step(`^wird die Archivierung mit Hinweis auf den aktiven Lauf abgewiesen$`, s.activeRunRejected)
	sc.Step(`^bleibt der abgeschlossene Lauf unter derselben Kennung im Verlauf von "([^"]+)" sichtbar$`, s.runPreserved)
	sc.Step(`^enthält der öffentliche Aktivitätsverlauf von "([^"]+)" denselben Anlageeintrag$`, s.activityPreserved)
	sc.Step(`^wird die neue Laufreservierung verweigert$`, s.runReservationDenied)
	sc.Step(`^der Verlauf von "([^"]+)" enthält keinen neuen Lauf$`, s.noNewRun)
	sc.Step(`^ist "([^"]+)" genau einmal archiviert$`, s.raceArchivedOnce)
	sc.Step(`^entweder ist "([^"]+)" mit Aufgabe und beiden Aktivitätseinträgen vollständig angelegt oder mit Archivhinweis ganz abgewiesen$`, s.raceOutcome)
	sc.Step(`^zeigt der öffentliche Aktivitätsverlauf von "([^"]+)" keinen Eintrag von "([^"]+)"$`, s.activityNotLeaked)
}

func (s *Suite) key(org, name string) string { return org + "\x00" + name }
func (s *Suite) orgPath(org string) string   { return "/api/organisationen/" + s.organizations[org] }
func (s *Suite) projectPath(org, project string) string {
	return s.orgPath(org) + "/projekte/" + s.projects[s.key(org, project)]
}
func (s *Suite) taskPath(org, project string) string {
	return s.projectPath(org, project) + "/aufgaben"
}

func (s *Suite) ensureOrganization(name string) error {
	if s.address == "" {
		if err := s.fresh(); err != nil {
			return err
		}
	}
	if s.organizations[name] != "" {
		s.activeOrg = name
		return nil
	}
	body, _ := json.Marshal(map[string]string{"name": name, "description": "Archivtest"})
	if err := s.request("POST", "/api/organisationen", string(body)); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var item Organization
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID == "" || item.Name != name {
		return fmt.Errorf("Organisation: %s", s.response.Body)
	}
	s.organizations[name], s.activeOrg = item.ID, name
	return nil
}

func (s *Suite) twoOrganizations(first, second string) error {
	if err := s.ensureOrganization(first); err != nil {
		return err
	}
	return s.ensureOrganization(second)
}

func (s *Suite) projectSetup(org, project string) error {
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	return s.ensureProject(org, project)
}

func (s *Suite) projectAndAgent(org, project, agent string) error {
	if err := s.projectSetup(org, project); err != nil {
		return err
	}
	return s.ensureAgent(agent, org)
}

func (s *Suite) ensureNamedProject(project, org string) error { return s.projectSetup(org, project) }

func (s *Suite) ensureProject(org, project string) error {
	if s.projects[s.key(org, project)] != "" {
		s.activeOrg, s.activeProject = org, project
		return nil
	}
	goal, err := s.createGoal(org, project)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{"name": project, "description": "Archivtestprojekt", "goal_id": goal})
	if err := s.request("POST", s.orgPath(org)+"/projekte", string(body)); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var item Project
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID == "" || item.Name != project {
		return fmt.Errorf("Projekt: %s", s.response.Body)
	}
	s.projects[s.key(org, project)], s.activeOrg, s.activeProject = item.ID, org, project
	return nil
}

func (s *Suite) createGoal(org, project string) (string, error) {
	body, _ := json.Marshal(map[string]string{"name": "Ziel " + project})
	if err := s.request("POST", s.orgPath(org)+"/ziele", string(body)); err != nil {
		return "", err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return "", err
	}
	var item Goal
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return "", err
	}
	if item.ID == "" {
		return "", fmt.Errorf("Ziel: %s", s.response.Body)
	}
	return item.ID, nil
}

func (s *Suite) ensureAgent(name, org string) error {
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	if s.agents[s.key(org, name)] != "" {
		return nil
	}
	body, _ := json.Marshal(map[string]string{"name": name, "template_id": "recherche", "execution_kind": "eino"})
	if err := s.request("POST", s.orgPath(org)+"/agenten", string(body)); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var item Agent
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID == "" {
		return fmt.Errorf("Agent: %s", s.response.Body)
	}
	s.agents[s.key(org, name)] = item.ID
	return nil
}

func (s *Suite) postTask(org, project, title string) error {
	body, _ := json.Marshal(map[string]string{"title": title, "description": "Archiv-Nachweis", "priority": "normal", "assignee_id": s.agents[s.key(org, "Mira")]})
	s.activeOrg, s.activeProject = org, project
	return s.request("POST", s.taskPath(org, project), string(body))
}

func (s *Suite) ensureTask(title, project string) error {
	if err := s.postTask(s.activeOrg, project, title); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var item Task
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID == "" || item.Title != title {
		return fmt.Errorf("Aufgabe: %s", s.response.Body)
	}
	s.tasks[s.key(s.activeOrg, title)] = item.ID
	return nil
}

func (s *Suite) archiveSetup(project, org string) error {
	if err := s.archive(project, org); err != nil {
		return err
	}
	return s.status(http.StatusOK)
}

func (s *Suite) archive(project, org string) error {
	s.activeOrg, s.activeProject = org, project
	return s.request("POST", s.projectPath(org, project)+"/archivieren", "")
}

func (s *Suite) restore(project, org string) error {
	s.activeOrg, s.activeProject = org, project
	return s.request("POST", s.projectPath(org, project)+"/wiederherstellen", "")
}

func (s *Suite) directTask(title, project string) error {
	s.rejectedTitle = title
	return s.postTask(s.activeOrg, project, title)
}

func (s *Suite) canCreateTask(title, project string) error {
	if err := s.postTask(s.activeOrg, project, title); err != nil {
		return err
	}
	return s.status(http.StatusCreated)
}

func (s *Suite) cannotCreateTask() error {
	if err := s.postTask(s.activeOrg, s.activeProject, "Verbotener Auftrag"); err != nil {
		return err
	}
	return s.archivedTaskRejected()
}

func (s *Suite) archiveForeign(project, foreignOrg string) error {
	return s.request("POST", s.orgPath(foreignOrg)+"/projekte/"+s.projects[s.key("Nordstern", project)]+"/archivieren", "")
}

func (s *Suite) restoreForeign(project, foreignOrg string) error {
	return s.request("POST", s.orgPath(foreignOrg)+"/projekte/"+s.projects[s.key("Nordstern", project)]+"/wiederherstellen", "")
}

func (s *Suite) archiveUnknown(org string) error {
	return s.request("POST", s.orgPath(org)+"/projekte/unbekannt/archivieren", "")
}

func (s *Suite) projectStatus(label, project string) error {
	if err := s.request("GET", s.projectPath(s.activeOrg, project)+"/status", ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var item StatusResponse
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	want := map[string]string{"Aktiv": "active", "Archiviert": "archived"}[label]
	if item.Status != want {
		return fmt.Errorf("Projektstatus: %+v", item)
	}
	return nil
}

func (s *Suite) listProjects(org string, archive bool) (ProjectList, error) {
	var list ProjectList
	path := s.orgPath(org) + "/projekte"
	if archive {
		path += "/archiv"
	}
	if err := s.request("GET", path, ""); err != nil {
		return list, err
	}
	if err := s.status(http.StatusOK); err != nil {
		return list, err
	}
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return list, err
	}
	if list.Projects == nil {
		return list, fmt.Errorf("Projektliste fehlt: %s", s.response.Body)
	}
	return list, nil
}

func (s *Suite) countProject(org, project string, archive bool, want int) error {
	list, err := s.listProjects(org, archive)
	if err != nil {
		return err
	}
	count := 0
	for _, item := range list.Projects {
		if item.ID == s.projects[s.key(org, project)] && item.Name == project {
			count++
		}
	}
	if count != want {
		return fmt.Errorf("Projekt %s %d-mal statt %d: %s", project, count, want, s.response.Body)
	}
	return nil
}

func (s *Suite) activeAbsent(org, project string) error {
	return s.countProject(org, project, false, 0)
}
func (s *Suite) activeOnce(org, project string) error { return s.countProject(org, project, false, 1) }
func (s *Suite) archiveAbsent(org, project string) error {
	return s.countProject(org, project, true, 0)
}
func (s *Suite) archiveOnce(org, project string) error     { return s.countProject(org, project, true, 1) }
func (s *Suite) remainsActive(project, org string) error   { return s.activeOnce(org, project) }
func (s *Suite) remainsArchived(project, org string) error { return s.archiveOnce(org, project) }

func (s *Suite) taskPreserved(title string) error {
	path := s.taskPath(s.activeOrg, s.activeProject) + "/" + s.tasks[s.key(s.activeOrg, title)]
	if err := s.request("GET", path, ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var item Task
	if err := json.Unmarshal(s.response.Body, &item); err != nil {
		return err
	}
	if item.ID != s.tasks[s.key(s.activeOrg, title)] || item.Title != title {
		return fmt.Errorf("Aufgabe verändert: %+v", item)
	}
	return nil
}

func (s *Suite) archivedTaskRejected() error {
	if s.response.Status != http.StatusConflict && s.response.Status != http.StatusUnprocessableEntity {
		return fmt.Errorf("Archiv-Ablehnung fehlt: %d %s", s.response.Status, s.response.Body)
	}
	if !strings.Contains(strings.ToLower(string(s.response.Body)), "archiv") {
		return fmt.Errorf("Archivhinweis fehlt: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) noNewTask(project string) error {
	if err := s.request("GET", s.taskPath(s.activeOrg, project), ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var list TaskList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	for _, item := range list.Tasks {
		if item.Title == s.rejectedTitle {
			return fmt.Errorf("Teilerzeugung: %+v", item)
		}
	}
	return nil
}

func (s *Suite) notFound() error {
	if err := s.status(http.StatusNotFound); err != nil {
		return err
	}
	if strings.Contains(string(s.response.Body), "Website") || strings.Contains(string(s.response.Body), s.projects[s.key("Nordstern", "Website")]) {
		return fmt.Errorf("Fremddaten: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) activeRunRejected() error {
	if s.response.Status != http.StatusConflict {
		return fmt.Errorf("Aktiver Lauf: %d %s", s.response.Status, s.response.Body)
	}
	if !strings.Contains(string(s.response.Body), "active_run") {
		return fmt.Errorf("Laufhinweis fehlt: %s", s.response.Body)
	}
	return nil
}
