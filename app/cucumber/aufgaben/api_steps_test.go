package aufgaben

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"agentcontrolplane/app/internal/adapter/sqlite"
	appagent "agentcontrolplane/app/internal/app/agent"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainagent "agentcontrolplane/app/internal/domain/agent"
	"github.com/cucumber/godog"
)

func (s *Suite) initialize(sc *godog.ScenarioContext) {
	sc.After(s.after)
	s.registerSetup(sc)
	s.registerTaskActions(sc)
	s.registerTaskAssertions(sc)
	s.registerBrowser(sc)
}

func (s *Suite) registerSetup(sc *godog.ScenarioContext) {
	sc.Step(`^ich starte den lokalen Server mit einer neuen SQLite-Datenbank als Betreiberin$`, s.fresh)
	sc.Step(`^die Organisation "([^"]+)" mit dem Projekt "([^"]+)" besteht$`, s.ensureProject)
	sc.Step(`^die Organisation "([^"]+)" mit dem Projekt "([^"]+)" und dem Agenten "([^"]+)" besteht$`, s.ensureProjectAndAgent)
	sc.Step(`^die Organisation "([^"]+)" besteht$`, s.ensureOrganization)
	sc.Step(`^die Organisationen "([^"]+)" und "([^"]+)" bestehen$`, s.ensureOrganizations)
	sc.Step(`^das Projekt "([^"]+)" besteht in "([^"]+)"$`, s.ensureNamedProject)
	sc.Step(`^der Agent "([^"]+)" mit der Ausführungsart "([^"]+)" besteht in "([^"]+)"$`, s.ensureTypedAgent)
	sc.Step(`^der Agent "([^"]+)" besteht in "([^"]+)"$`, s.ensureAgent)
	sc.Step(`^die Agenten "([^"]+)" und "([^"]+)" bestehen in "([^"]+)"$`, s.ensureAgents)
	sc.Step(`^der Agent "([^"]+)" besteht pausiert in "([^"]+)"$`, s.ensurePausedAgent)
}

func (s *Suite) registerTaskActions(sc *godog.ScenarioContext) {
	sc.Step(`^ich die Aufgabe "([^"]+)" mit der Beschreibung "([^"]+)", der Priorität "([^"]+)" und dem einzigen zuständigen Agenten "([^"]+)" im Projekt "([^"]+)" über HTTP anlege$`, s.createFullTask)
	sc.Step(`^ich eine Aufgabe mit dem Titel "([^"]*)" und dem einzigen zuständigen Agenten "([^"]+)" im Projekt "([^"]+)" über HTTP anlege$`, s.createTitledTask)
	sc.Step(`^ich die Aufgabe "([^"]+)" ohne zuständigen Agenten im Projekt "([^"]+)" über HTTP anlege$`, s.createUnassignedTask)
	sc.Step(`^ich die Aufgabe "([^"]+)" mit "([^"]+)" und "([^"]+)" zugleich als zuständigen Agenten im Projekt "([^"]+)" über HTTP anlege$`, s.createMultiAssignedTask)
	sc.Step(`^ich die Aufgabe "([^"]+)" mit einer unbekannten Agentenkennung als einzigem zuständigen Agenten im Projekt "([^"]+)" über HTTP anlege$`, s.createUnknownAgentTask)
	sc.Step(`^ich die Aufgabe "([^"]+)" mit "([^"]+)" aus "([^"]+)" als einzigem zuständigen Agenten im Projekt "([^"]+)" über HTTP anlege$`, s.createForeignAgentTask)
	sc.Step(`^ich die Aufgabe "([^"]+)" mit "([^"]+)" als einzigem zuständigen Agenten im Projekt "([^"]+)" über HTTP anlege$`, s.createNamedAgentTask)
	sc.Step(`^ich die Aufgabe "([^"]+)" mit "([^"]+)" für eine unbekannte Projektkennung in "([^"]+)" über HTTP anlege$`, s.createUnknownProjectTask)
	sc.Step(`^ich die Aufgabe "([^"]+)" mit "([^"]+)" für das Projekt "([^"]+)" aus "([^"]+)" in "([^"]+)" über HTTP anlege$`, s.createForeignProjectTask)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank erneut starte$`, s.restart)
}

func (s *Suite) registerTaskAssertions(sc *godog.ScenarioContext) {
	sc.Step(`^antwortet der Server mit einer neuen Aufgabenkennung und dem HTTP-Status 201$`, s.createdTask)
	sc.Step(`^die Aufgabenliste des Projekts "([^"]+)" enthält "([^"]+)" genau einmal$`, s.taskOnce)
	sc.Step(`^die Aufgabenliste des Projekts "([^"]+)" enthält "([^"]+)" weiterhin genau einmal$`, s.taskOnce)
	sc.Step(`^das Aufgabendetail zeigt "([^"]+)", "([^"]+)", die Priorität "([^"]+)" und das Projekt "([^"]+)"$`, s.taskDetailFields)
	sc.Step(`^das Aufgabendetail zeigt genau "([^"]+)" als zuständigen Agenten mit der Ausführungsart "([^"]+)"$`, s.taskDetailAssignee)
	sc.Step(`^das Aufgabendetail zeigt den anfänglichen Status "([^"]+)"$`, s.taskOpen)
	sc.Step(`^zeigt dasselbe Aufgabendetail unter derselben Aufgabenkennung alle diese Angaben unverändert$`, s.taskAfterRestart)
	sc.Step(`^antwortet der Server mit einem Hinweis am Aufgabenfeld "([^"]+)"$`, s.taskFieldError)
	sc.Step(`^antwortet der Server mit einem Hinweis am Aufgabenfeld "([^"]+)" ohne fremde Agentendaten$`, s.taskAssigneeErrorNoLeak)
	sc.Step(`^die Aufgabenliste des Projekts "([^"]+)" bleibt leer$`, s.tasksEmpty)
	sc.Step(`^die bestehenden Agenten bleiben unverändert$`, s.agentsUnchanged)
	sc.Step(`^wird die Projektzuordnung ohne fremde Projektdaten abgewiesen$`, s.projectRejected)
	sc.Step(`^in "([^"]+)" entsteht keine Aufgabe$`, s.noTaskInOrganization)
	sc.Step(`^die Aufgabenliste von "([^"]+)" bleibt unverändert$`, s.tasksEmpty)
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
	return s.postOrganization(name)
}

func (s *Suite) postOrganization(name string) error {
	body, _ := json.Marshal(map[string]string{"name": name, "description": "Testorganisation"})
	if err := s.request("POST", "/api/organisationen", string(body), "application/json"); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var organization Organization
	if err := json.Unmarshal(s.response.Body, &organization); err != nil {
		return err
	}
	if organization.ID == "" || organization.Name != name {
		return fmt.Errorf("Organisation: %s", s.response.Body)
	}
	s.organizations[name], s.activeOrg = organization.ID, name
	return nil
}

func (s *Suite) ensureOrganizations(first, second string) error {
	if err := s.ensureOrganization(first); err != nil {
		return err
	}
	return s.ensureOrganization(second)
}

func (s *Suite) key(org, name string) string { return org + "\x00" + name }
func (s *Suite) projectPath(org, project string) string {
	return "/api/organisationen/" + s.organizations[org] + "/projekte/" + s.projects[s.key(org, project)]
}
func (s *Suite) taskPath(org, project string) string {
	return s.projectPath(org, project) + "/aufgaben"
}
func (s *Suite) taskDetailPath(org, project, title string) string {
	return s.taskPath(org, project) + "/" + s.tasks[s.key(org, title)]
}

func (s *Suite) ensureProject(org, project string) error {
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	if s.projects[s.key(org, project)] != "" {
		return nil
	}
	goalID, err := s.createGoal(org, project)
	if err != nil {
		return err
	}
	return s.postProject(org, project, goalID)
}

func (s *Suite) createGoal(org, project string) (string, error) {
	goalBody, _ := json.Marshal(map[string]string{"name": "Ziel " + project})
	goalPath := "/api/organisationen/" + s.organizations[org] + "/ziele"
	if err := s.request("POST", goalPath, string(goalBody), "application/json"); err != nil {
		return "", err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return "", err
	}
	var goal Goal
	if err := json.Unmarshal(s.response.Body, &goal); err != nil {
		return "", err
	}
	return goal.ID, nil
}

func (s *Suite) postProject(org, project, goalID string) error {
	body, _ := json.Marshal(map[string]string{"name": project, "description": "Testprojekt", "goal_id": goalID})
	path := "/api/organisationen/" + s.organizations[org] + "/projekte"
	if err := s.request("POST", path, string(body), "application/json"); err != nil {
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

func (s *Suite) ensureNamedProject(project, org string) error { return s.ensureProject(org, project) }

func (s *Suite) ensureProjectAndAgent(org, project, agent string) error {
	if err := s.ensureProject(org, project); err != nil {
		return err
	}
	return s.ensureAgent(agent, org)
}

func (s *Suite) ensureAgent(name, org string) error {
	return s.ensureTypedAgent(name, "Eino", org)
}

func (s *Suite) ensureAgents(first, second, org string) error {
	if err := s.ensureAgent(first, org); err != nil {
		return err
	}
	return s.ensureAgent(second, org)
}

func (s *Suite) ensureTypedAgent(name, kind, org string) error {
	if err := s.ensureOrganization(org); err != nil {
		return err
	}
	if s.agents[s.key(org, name)] != "" {
		return nil
	}
	template, execution := "recherche", "eino"
	if kind == "Codex CLI" {
		template, execution = "codex-cli", "codex_cli"
	}
	body, _ := json.Marshal(map[string]string{"name": name, "template_id": template, "execution_kind": execution})
	return s.postAgent(name, org, string(body))
}

func (s *Suite) postAgent(name, org, body string) error {
	path := "/api/organisationen/" + s.organizations[org] + "/agenten"
	if err := s.request("POST", path, body, "application/json"); err != nil {
		return err
	}
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if agent.ID == "" || agent.Name != name {
		return fmt.Errorf("Agent: %s", s.response.Body)
	}
	s.agents[s.key(org, name)] = agent.ID
	return nil
}

func (s *Suite) ensurePausedAgent(name, org string) error {
	if err := s.ensureAgent(name, org); err != nil {
		return err
	}
	s.stopServer()
	if err := s.pauseAgentInStore(name, org); err != nil {
		return err
	}
	return s.start()
}

func (s *Suite) pauseAgentInStore(name, org string) error {
	ctx := context.Background()
	db, err := sqlite.OpenApplication(ctx, s.database)
	if err != nil {
		return err
	}
	service := appagent.NewService(db, apporganisation.NewLocalIdentity(), sqlite.NewOrganizationStore(db))
	profile, err := service.SetStatus(ctx, s.organizations[org], s.agents[s.key(org, name)], domainagent.StatusPaused)
	if err != nil {
		db.Close()
		return err
	}
	if profile.Status != domainagent.StatusPaused {
		db.Close()
		return fmt.Errorf("Agent nicht pausiert: %+v", profile)
	}
	return db.Close()
}

func (s *Suite) status(expected int) error {
	if s.response.Status != expected {
		return fmt.Errorf("HTTP %d statt %d: %s", s.response.Status, expected, s.response.Body)
	}
	return nil
}

func (s *Suite) priority(label string) string {
	if label == "Hoch" {
		return "high"
	}
	return "normal"
}

func (s *Suite) postTask(org, project, title, description, priority string, assignee any) error {
	s.activeOrg, s.activeProject = org, project
	body, _ := json.Marshal(map[string]any{"title": title, "description": description, "priority": priority, "assignee_id": assignee})
	return s.request("POST", s.taskPath(org, project), string(body), "application/json")
}

func (s *Suite) createFullTask(title, description, priority, agent, project string) error {
	return s.postTask(s.activeOrg, project, title, description, s.priority(priority), s.agents[s.key(s.activeOrg, agent)])
}

func (s *Suite) createTitledTask(title, agent, project string) error {
	return s.postTask(s.activeOrg, project, title, "Nur Versuch", "normal", s.agents[s.key(s.activeOrg, agent)])
}

func (s *Suite) createUnassignedTask(title, project string) error {
	return s.postTask(s.activeOrg, project, title, "Nur Versuch", "normal", "")
}

func (s *Suite) createMultiAssignedTask(title, first, second, project string) error {
	agents := []string{s.agents[s.key(s.activeOrg, first)], s.agents[s.key(s.activeOrg, second)]}
	body := fmt.Sprintf(`{"title":%q,"priority":"normal","assignee_id":%q,"assignee_id":%q}`, title, agents[0], agents[1])
	if err := s.request("POST", s.taskPath(s.activeOrg, project), body, "application/json"); err != nil {
		return err
	}
	if err := s.taskFieldError("Zuständig"); err != nil {
		return err
	}
	if err := s.tasksEmpty(project); err != nil {
		return err
	}
	return s.postTask(s.activeOrg, project, title, "Nur Versuch", "normal", agents)
}

func (s *Suite) createUnknownAgentTask(title, project string) error {
	return s.postTask(s.activeOrg, project, title, "Nur Versuch", "normal", "unbekannt")
}

func (s *Suite) createForeignAgentTask(title, agent, foreignOrg, project string) error {
	return s.postTask(s.activeOrg, project, title, "Nur Versuch", "normal", s.agents[s.key(foreignOrg, agent)])
}

func (s *Suite) createNamedAgentTask(title, agent, project string) error {
	return s.postTask(s.activeOrg, project, title, "Nur Versuch", "normal", s.agents[s.key(s.activeOrg, agent)])
}

func (s *Suite) createUnknownProjectTask(title, agent, org string) error {
	s.activeOrg = org
	body, _ := json.Marshal(map[string]any{"title": title, "priority": "normal", "assignee_id": s.agents[s.key(org, agent)]})
	return s.request("POST", "/api/organisationen/"+s.organizations[org]+"/projekte/unbekannt/aufgaben", string(body), "application/json")
}

func (s *Suite) createForeignProjectTask(title, agent, project, foreignOrg, org string) error {
	s.activeOrg = org
	body, _ := json.Marshal(map[string]any{"title": title, "priority": "normal", "assignee_id": s.agents[s.key(org, agent)]})
	path := "/api/organisationen/" + s.organizations[org] + "/projekte/" + s.projects[s.key(foreignOrg, project)] + "/aufgaben"
	return s.request("POST", path, string(body), "application/json")
}

func (s *Suite) createdTask() error {
	if err := s.status(http.StatusCreated); err != nil {
		return err
	}
	var task Task
	if err := json.Unmarshal(s.response.Body, &task); err != nil {
		return err
	}
	if task.ID == "" || task.ProjectID != s.projects[s.key(s.activeOrg, s.activeProject)] {
		return fmt.Errorf("Aufgabe: %s", s.response.Body)
	}
	s.createdID, s.tasks[s.key(s.activeOrg, task.Title)] = task.ID, task.ID
	s.createdAssignee = task.AssigneeID
	return nil
}

func (s *Suite) getTaskList(org, project string) (TaskList, error) {
	var list TaskList
	if err := s.request("GET", s.taskPath(org, project), "", ""); err != nil {
		return list, err
	}
	if err := s.status(http.StatusOK); err != nil {
		return list, err
	}
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return list, err
	}
	if list.Tasks == nil {
		return list, fmt.Errorf("Aufgabenliste fehlt: %s", s.response.Body)
	}
	return list, nil
}

func (s *Suite) taskOnce(project, title string) error {
	list, err := s.getTaskList(s.activeOrg, project)
	if err != nil {
		return err
	}
	count := 0
	for _, task := range list.Tasks {
		if task.Title == title {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("Aufgabe %q %d-mal: %s", title, count, s.response.Body)
	}
	return nil
}

func (s *Suite) taskDetail() (Task, error) {
	var task Task
	if err := s.request("GET", s.taskPath(s.activeOrg, s.activeProject)+"/"+s.createdID, "", ""); err != nil {
		return task, err
	}
	if err := s.status(http.StatusOK); err != nil {
		return task, err
	}
	if err := json.Unmarshal(s.response.Body, &task); err != nil {
		return task, err
	}
	return task, nil
}

func (s *Suite) taskDetailFields(title, description, priority, project string) error {
	task, err := s.taskDetail()
	if err != nil {
		return err
	}
	if task.ID != s.createdID || task.Title != title || task.Description != description || task.Priority != s.priority(priority) || task.ProjectID != s.projects[s.key(s.activeOrg, project)] {
		return fmt.Errorf("Aufgabendetail: %+v", task)
	}
	return nil
}

func (s *Suite) taskDetailAssignee(agent, kind string) error {
	task, err := s.taskDetail()
	if err != nil {
		return err
	}
	if task.AssigneeID != s.agents[s.key(s.activeOrg, agent)] {
		return fmt.Errorf("Zuständigkeit: %+v", task)
	}
	return s.agentKind(agent, kind)
}

func (s *Suite) agentKind(name, kind string) error {
	path := "/api/organisationen/" + s.organizations[s.activeOrg] + "/agenten/" + s.agents[s.key(s.activeOrg, name)]
	if err := s.request("GET", path, "", ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var agent Agent
	if err := json.Unmarshal(s.response.Body, &agent); err != nil {
		return err
	}
	if kind == "Eino" && agent.ExecutionKind == "eino" {
		return nil
	}
	if kind == "Codex CLI" && agent.ExecutionKind == "codex_cli" {
		return nil
	}
	return fmt.Errorf("Ausführungsart: %+v", agent)
}

func (s *Suite) taskOpen(status string) error {
	task, err := s.taskDetail()
	if err != nil {
		return err
	}
	if status != "Offen" || task.Status != "open" {
		return fmt.Errorf("Anfangsstatus: %+v", task)
	}
	return nil
}

func (s *Suite) taskAfterRestart() error {
	task, err := s.taskDetail()
	if err != nil {
		return err
	}
	if task.ID != s.createdID || task.Title != "Startseite prüfen" || task.Description != "Navigation und Texte prüfen" || task.Priority != "high" || task.Status != "open" || task.ProjectID != s.projects[s.key(s.activeOrg, s.activeProject)] {
		return fmt.Errorf("Aufgabe nach Neustart: %+v", task)
	}
	if task.AssigneeID == "" || task.AssigneeID != s.createdAssignee {
		return fmt.Errorf("Empfänger nach Neustart fehlt: %+v", task)
	}
	return nil
}

func (s *Suite) taskFieldError(field string) error {
	if s.response.Status < 400 || s.response.Status >= 500 {
		return fmt.Errorf("Feldfehler HTTP %d: %s", s.response.Status, s.response.Body)
	}
	var problem FieldError
	if err := json.Unmarshal(s.response.Body, &problem); err != nil {
		return err
	}
	key := map[string]string{"Titel": "title", "Zuständig": "assignee_id", "Projekt": "project_id"}[field]
	if key == "" || problem.FieldErrors[key] == "" {
		return fmt.Errorf("Feldfehler %s fehlt: %s", field, s.response.Body)
	}
	return nil
}

func (s *Suite) taskAssigneeErrorNoLeak(field string) error {
	if err := s.taskFieldError(field); err != nil {
		return err
	}
	if strings.Contains(string(s.response.Body), "Fremd") || strings.Contains(string(s.response.Body), "Südstern") {
		return fmt.Errorf("Fremde Agentendaten: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) tasksEmpty(project string) error {
	org := s.activeOrg
	if project == "Fremdprojekt" {
		org = "Südstern"
	}
	list, err := s.getTaskList(org, project)
	if err != nil {
		return err
	}
	if len(list.Tasks) != 0 {
		return fmt.Errorf("Unerwartete Aufgaben: %+v", list.Tasks)
	}
	return nil
}

func (s *Suite) agentsUnchanged() error {
	for key, id := range s.agents {
		parts := strings.SplitN(key, "\x00", 2)
		path := "/api/organisationen/" + s.organizations[parts[0]] + "/agenten/" + id
		if err := s.request("GET", path, "", ""); err != nil {
			return err
		}
		if err := s.status(http.StatusOK); err != nil {
			return err
		}
		var agent Agent
		if err := json.Unmarshal(s.response.Body, &agent); err != nil {
			return err
		}
		if agent.ID != id || agent.Name != parts[1] || agent.OrganizationID != s.organizations[parts[0]] {
			return fmt.Errorf("Agent verändert: %+v", agent)
		}
	}
	return nil
}

func (s *Suite) projectRejected() error {
	if err := s.taskFieldError("Projekt"); err != nil {
		return err
	}
	if strings.Contains(string(s.response.Body), "Fremdprojekt") || strings.Contains(string(s.response.Body), "Südstern") {
		return fmt.Errorf("Fremde Projektdaten: %s", s.response.Body)
	}
	return nil
}

func (s *Suite) noTaskInOrganization(org string) error {
	path := "/api/organisationen/" + s.organizations[org] + "/projekte"
	if err := s.request("GET", path, "", ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var list ProjectList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	for _, project := range list.Projects {
		if err := s.checkProjectTasksEmpty(org, project.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Suite) checkProjectTasksEmpty(org, id string) error {
	path := "/api/organisationen/" + s.organizations[org] + "/projekte/" + id + "/aufgaben"
	if err := s.request("GET", path, "", ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var list TaskList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	if len(list.Tasks) != 0 {
		return fmt.Errorf("Teilaufgabe angelegt: %s", s.response.Body)
	}
	return nil
}
