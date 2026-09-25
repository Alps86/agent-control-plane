package ziel

import (
	"context"

	domainziel "agentcontrolplane/app/internal/domain/ziel"
	portziel "agentcontrolplane/app/internal/port/ziel"
	"github.com/google/uuid"
)

// NewService bindet Zielstore und Organisationszugriff an den Anwendungsfall.
func NewService(store portziel.GoalStore, organizations portziel.OrganizationReader) *Service {
	return &Service{store: store, organizations: organizations}
}

// Create legt ein neues Stammziel in einer zugänglichen Organisation an.
func (s *Service) Create(ctx context.Context, organizationID, name string) (domainziel.Goal, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return domainziel.Goal{}, err
	}

	goal, err := (domainziel.Goal{}).NewRootGoal(uuid.NewString(), organizationID, name)
	if err != nil {
		return domainziel.Goal{}, err
	}

	if err := s.store.CreateGoal(ctx, goal); err != nil {
		return domainziel.Goal{}, err
	}

	return goal, nil
}

// CreateChild legt ein Teilziel unter einem Ziel derselben Organisation an.
func (s *Service) CreateChild(ctx context.Context, organizationID, parentGoalID, name string) (domainziel.Goal, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return domainziel.Goal{}, err
	}

	goal, err := (domainziel.Goal{}).NewChildGoal(uuid.NewString(), organizationID, parentGoalID, name)
	if err != nil {
		return domainziel.Goal{}, err
	}

	if err := s.store.CreateChildGoal(ctx, goal); err != nil {
		return domainziel.Goal{}, err
	}

	return goal, nil
}

// SetStatus ändert nur den ausdrücklich benannten Zielstatus.
func (s *Service) SetStatus(ctx context.Context, organizationID, goalID, status string) (domainziel.Goal, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return domainziel.Goal{}, err
	}

	if _, err := (domainziel.Goal{}).WithStatus(status); err != nil {
		return domainziel.Goal{}, err
	}

	return s.store.SetGoalStatus(ctx, organizationID, goalID, status)
}

// List zeigt nur Ziele einer zugänglichen Organisation.
func (s *Service) List(ctx context.Context, organizationID string) ([]domainziel.Goal, error) {
	if err := s.checkOrganization(ctx, organizationID); err != nil {
		return nil, err
	}

	return s.store.ListGoals(ctx, organizationID)
}

// ProjectPaths liefert jeden Projektpfad vom Stammziel zum verknüpften Ziel.
func (s *Service) ProjectPaths(ctx context.Context, organizationID string) ([]domainziel.ProjectPath, error) {
	goals, err := s.List(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	links, err := s.store.ListGoalProjectLinks(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	return s.projectPaths(goals, links)
}

func (s *Service) projectPaths(goals []domainziel.Goal, links []domainziel.ProjectLink) ([]domainziel.ProjectPath, error) {
	lookup := make(map[string]domainziel.Goal, len(goals))
	for _, goal := range goals {
		lookup[goal.ID] = goal
	}

	paths := make([]domainziel.ProjectPath, 0, len(links))
	for _, link := range links {
		path, err := s.projectPath(link, lookup)
		if err != nil {
			return nil, err
		}

		paths = append(paths, path)
	}

	return paths, nil
}

func (s *Service) projectPath(link domainziel.ProjectLink, goals map[string]domainziel.Goal) (domainziel.ProjectPath, error) {
	path := domainziel.ProjectPath{ProjectID: link.ProjectID, ProjectName: link.ProjectName, GoalIDs: []string{}, GoalNames: []string{}}
	seen := map[string]bool{}
	for id := link.GoalID; id != ""; {
		goal, ok := goals[id]
		if !ok || seen[id] {
			return domainziel.ProjectPath{}, ErrInvalidParent
		}

		seen[id] = true
		path.GoalIDs = append([]string{id}, path.GoalIDs...)
		path.GoalNames = append([]string{goal.Name}, path.GoalNames...)
		id = s.parentID(goal)
	}

	return path, nil
}

func (s *Service) parentID(goal domainziel.Goal) string {
	if goal.ParentGoalID == nil {
		return ""
	}

	return *goal.ParentGoalID
}

func (s *Service) checkOrganization(ctx context.Context, organizationID string) error {
	if s == nil || s.store == nil || s.organizations == nil {
		return ErrAccessDenied
	}

	_, err := s.organizations.Get(ctx, organizationID)
	return err
}
