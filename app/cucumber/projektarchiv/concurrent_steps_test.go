package projektarchiv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

func (s *Suite) archiveAndCreateConcurrently(project, title string) error {
	s.raceTitle = title
	org := s.activeOrg
	path := s.projectPath(org, project)
	body, _ := json.Marshal(map[string]string{"title": title, "description": "Gleichzeitiger Auftrag", "priority": "normal", "assignee_id": s.agents[s.key(org, "Mira")]})
	results := make(chan ConcurrentResult, 2)
	var wait sync.WaitGroup
	wait.Add(2)
	go s.parallelRequest(&wait, results, "archive", "POST", path+"/archivieren", "")
	go s.parallelRequest(&wait, results, "task", "POST", path+"/aufgaben", string(body))
	wait.Wait()
	close(results)
	return s.collectRace(results)
}

func (s *Suite) parallelRequest(wait *sync.WaitGroup, results chan<- ConcurrentResult, kind, method, path, body string) {
	defer wait.Done()
	response, err := s.rawRequest(method, path, body)
	results <- ConcurrentResult{Kind: kind, Response: response, Err: err}
}

func (s *Suite) rawRequest(method, path, body string) (Response, error) {
	req, err := http.NewRequest(method, s.url(path), strings.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	reply, err := s.client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer reply.Body.Close()
	data, err := io.ReadAll(reply.Body)
	return Response{Status: reply.StatusCode, Body: data, Location: reply.Header.Get("Location")}, err
}

func (s *Suite) collectRace(results <-chan ConcurrentResult) error {
	for result := range results {
		if result.Err != nil {
			return result.Err
		}
		if result.Kind == "archive" {
			s.raceArchive = result.Response
			continue
		}
		s.raceTask = result.Response
	}
	return nil
}

func (s *Suite) raceArchivedOnce(project string) error {
	if s.raceArchive.Status != http.StatusOK {
		return fmt.Errorf("Archivaktion: %d %s", s.raceArchive.Status, s.raceArchive.Body)
	}
	return s.archiveOnce(s.activeOrg, project)
}

func (s *Suite) raceOutcome(title string) error {
	if title != s.raceTitle {
		return fmt.Errorf("Rennauftrag %q statt %q", title, s.raceTitle)
	}
	if s.raceTask.Status == http.StatusCreated {
		return s.raceCreated(title)
	}
	if s.raceTask.Status == http.StatusConflict {
		return s.raceRejected(title)
	}
	return fmt.Errorf("Aufgabenrennen: HTTP %d %s", s.raceTask.Status, s.raceTask.Body)
}

func (s *Suite) raceCreated(title string) error {
	var task Task
	if err := json.Unmarshal(s.raceTask.Body, &task); err != nil {
		return err
	}
	if task.ID == "" || task.Title != title || task.ProjectID != s.projects[s.key(s.activeOrg, s.activeProject)] {
		return fmt.Errorf("Teilaufgabe: %+v", task)
	}
	if err := s.raceTaskList(title, task.ID, 1); err != nil {
		return err
	}
	return s.raceActivity(title, task.ID, 2)
}

func (s *Suite) raceRejected(title string) error {
	if !strings.Contains(string(s.raceTask.Body), "project_archived") {
		return fmt.Errorf("Archivhinweis fehlt: %s", s.raceTask.Body)
	}
	if err := s.raceTaskList(title, "", 0); err != nil {
		return err
	}
	return s.raceActivity(title, "", 0)
}

func (s *Suite) raceTaskList(title, taskID string, want int) error {
	if err := s.request("GET", s.taskPath(s.activeOrg, s.activeProject), ""); err != nil {
		return err
	}
	if err := s.status(http.StatusOK); err != nil {
		return err
	}
	var list TaskList
	if err := json.Unmarshal(s.response.Body, &list); err != nil {
		return err
	}
	count := 0
	for _, task := range list.Tasks {
		if task.Title == title && (taskID == "" || task.ID == taskID) {
			count++
		}
	}
	if count != want {
		return fmt.Errorf("Teilzustand Aufgaben %d statt %d: %s", count, want, s.response.Body)
	}
	return nil
}

func (s *Suite) raceActivity(title, taskID string, want int) error {
	list, err := s.readActivity(s.activeOrg)
	if err != nil {
		return err
	}
	count, created, assigned := 0, false, false
	for _, event := range list.Events {
		if event.ObjectTitle != title {
			continue
		}
		if taskID != "" && event.TaskID != taskID {
			return fmt.Errorf("Falscher Aktivitätsbezug: %+v", event)
		}
		count++
		created = created || event.Kind == "created"
		assigned = assigned || event.Kind == "assigned"
	}
	if count != want || want == 2 && (!created || !assigned) {
		return fmt.Errorf("Teilzustand Aktivität %d statt %d (created=%t assigned=%t): %s", count, want, created, assigned, s.response.Body)
	}
	return nil
}
