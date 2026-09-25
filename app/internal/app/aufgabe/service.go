package aufgabe

import (
	"context"
	"errors"
	"strings"

	appagent "agentcontrolplane/app/internal/app/agent"
	appprojekt "agentcontrolplane/app/internal/app/projekt"
	domainagent "agentcontrolplane/app/internal/domain/agent"
	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	portaktivitaet "agentcontrolplane/app/internal/port/aktivitaet"
	portaufgabe "agentcontrolplane/app/internal/port/aufgabe"
	"github.com/google/uuid"
)

// NewService bindet Speicher und die bestehenden Projekt- und Agentengrenzen.
func NewService(store portaufgabe.Store, projects *appprojekt.Service, agents *appagent.Service, recorder portaktivitaet.Recorder, transactor portaktivitaet.Transactor) *Service {
	return &Service{store: store, projects: projects, agents: agents, recorder: recorder, transactor: transactor}
}

// Create legt eine Aufgabe nach allen Prüfungen ohne Teilmutation an.
func (s *Service) Create(ctx context.Context, organizationID string, input CreateInput) (domainaufgabe.Task, error) {
	if s == nil || s.store == nil || s.projects == nil || s.agents == nil || s.recorder == nil || s.transactor == nil {
		return domainaufgabe.Task{}, ErrAccessDenied
	}

	task, assigneeName, err := s.prepare(ctx, organizationID, input)
	if err != nil {
		return domainaufgabe.Task{}, err
	}

	if err := s.createWithEvents(ctx, task, assigneeName, input.Source); err != nil {
		return domainaufgabe.Task{}, err
	}

	return task, nil
}

func (s *Service) createWithEvents(ctx context.Context, task domainaufgabe.Task, assigneeName, source string) error {
	return s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.store.CreateTask(txCtx, task); err != nil {
			return err
		}

		return s.recorder.RecordTaskCreation(txCtx, task, assigneeName, source)
	})
}

func (s *Service) prepare(ctx context.Context, organizationID string, input CreateInput) (domainaufgabe.Task, string, error) {
	task, err := domainaufgabe.NewTask(uuid.NewString(), organizationID, strings.TrimSpace(input.ProjectID),
		input.Title, input.Description, input.Priority, input.AssigneeID)
	if err != nil {
		return domainaufgabe.Task{}, "", err
	}

	if err := s.checkProject(ctx, organizationID, task.ProjectID); err != nil {
		return domainaufgabe.Task{}, "", err
	}

	assigneeName, err := s.checkAssignee(ctx, organizationID, task.AssigneeID)
	if err != nil {
		return domainaufgabe.Task{}, "", err
	}

	return task, assigneeName, nil
}

// List liest Aufgaben nur aus einem Projekt derselben Organisation.
func (s *Service) List(ctx context.Context, organizationID, projectID string) ([]domainaufgabe.Task, error) {
	if s == nil || s.store == nil || s.projects == nil {
		return nil, ErrAccessDenied
	}

	if err := s.checkProject(ctx, organizationID, projectID); err != nil {
		return nil, err
	}

	return s.store.ListTasks(ctx, organizationID, projectID)
}

// Find liest eine Aufgabe nur innerhalb ihrer Organisation.
func (s *Service) Find(ctx context.Context, organizationID, taskID string) (domainaufgabe.Task, error) {
	if s == nil || s.store == nil || s.projects == nil {
		return domainaufgabe.Task{}, ErrAccessDenied
	}

	if _, err := s.projects.List(ctx, organizationID); err != nil {
		return domainaufgabe.Task{}, err
	}

	return s.store.FindTask(ctx, organizationID, taskID)
}

func (s *Service) checkProject(ctx context.Context, organizationID, projectID string) error {
	_, err := s.projects.Find(ctx, organizationID, projectID)
	if errors.Is(err, appprojekt.ErrNotFound) {
		return ErrInvalidProject
	}

	return err
}

func (s *Service) checkAssignee(ctx context.Context, organizationID, assigneeID string) (string, error) {
	profile, err := s.agents.Get(ctx, organizationID, assigneeID)
	if errors.Is(err, appagent.ErrNotFound) {
		return "", ErrInvalidAssignee
	}

	if err != nil {
		return "", err
	}

	if profile.Status != domainagent.StatusActive {
		return "", ErrAssigneePaused
	}

	return profile.Name, nil
}
