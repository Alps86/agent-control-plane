package projektort

import (
	"context"
	"errors"

	appprojekt "agentcontrolplane/app/internal/app/projekt"
	domainort "agentcontrolplane/app/internal/domain/projektort"
	portort "agentcontrolplane/app/internal/port/projektort"
)

// NewService bindet die vier schmalen Projektortgrenzen.
func NewService(store portort.Store, projects portort.ProjectReader, locator portort.Locator, adapter portort.StartProbe) *Service {
	return &Service{store: store, projects: projects, locator: locator, adapter: adapter}
}

// Get liefert nur einen gültigen appverwalteten Pfad oder einen korrigierbaren Zustand.
func (s *Service) Get(ctx context.Context, organizationID, projectID string) (Status, error) {
	if err := s.authorize(ctx, organizationID, projectID); err != nil {
		return Status{}, err
	}

	binding, err := s.store.Get(ctx, organizationID, projectID)
	if errors.Is(err, portort.ErrMissing) {
		return s.missing(), nil
	}

	if err != nil {
		return Status{}, err
	}

	return s.resolve(ctx, binding), nil
}

// Bind aktiviert ausschließlich den serverseitig abgeleiteten privaten Projektort.
func (s *Service) Bind(ctx context.Context, organizationID, projectID, kind string) (Status, error) {
	if err := s.authorize(ctx, organizationID, projectID); err != nil {
		return Status{}, err
	}

	binding := domainort.Binding{OrganizationID: organizationID, ProjectID: projectID, Kind: kind}
	if err := binding.Validate(); err != nil {
		return Status{}, err
	}

	if _, err := s.locator.Ensure(ctx, organizationID, projectID); err != nil {
		return s.invalid(), nil
	}

	if err := s.store.Save(ctx, binding); err != nil {
		return Status{}, err
	}

	return s.Get(ctx, organizationID, projectID)
}

// Prepare prüft die aktuelle Integrität und übergibt erst dann den Ort an den Adapter.
func (s *Service) Prepare(ctx context.Context, organizationID, projectID string) (Preparation, error) {
	status, err := s.Get(ctx, organizationID, projectID)
	if err != nil {
		return Preparation{}, err
	}

	if !status.StartReady || status.Path == nil {
		return Preparation{Code: status.Code, Path: nil, Reason: status.Reason}, nil
	}

	if s.adapter == nil {
		return Preparation{Code: CodeAdapterUnavailable, Reason: "Die Adapterprüfung ist nicht verfügbar."}, nil
	}

	if err := s.adapter.Probe(ctx, organizationID, projectID, *status.Path); err != nil {
		return s.probeFailure(err), nil
	}

	return Preparation{Ready: true, Code: CodeReady, Path: status.Path}, nil
}

func (s *Service) probeFailure(err error) Preparation {
	if errors.Is(err, portort.ErrIntegrity) {
		return Preparation{Code: CodeInvalid, Reason: s.invalid().Reason}
	}

	return Preparation{Code: CodeAdapterUnavailable, Reason: "Die Adapterprüfung ist nicht verfügbar."}
}

func (s *Service) authorize(ctx context.Context, organizationID, projectID string) error {
	if s == nil || s.store == nil || s.projects == nil || s.locator == nil {
		return ErrAccessDenied
	}

	_, err := s.projects.Find(ctx, organizationID, projectID)
	if errors.Is(err, appprojekt.ErrNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, appprojekt.ErrAccessDenied) {
		return ErrAccessDenied
	}

	return err
}

func (s *Service) resolve(ctx context.Context, binding domainort.Binding) Status {
	if !binding.Valid() {
		return s.invalidBinding(binding)
	}

	path, err := s.locator.Resolve(ctx, binding.OrganizationID, binding.ProjectID)
	if err != nil || path == "" {
		return s.invalidBinding(binding)
	}

	kind := binding.Kind
	return Status{Configured: true, Kind: &kind, Path: &path, StartReady: true, Code: CodeReady}
}

func (s *Service) missing() Status {
	return Status{Code: CodeMissing, Reason: "Aktivieren Sie den privaten Projektort vor dem Start."}
}

func (s *Service) invalid() Status {
	return Status{Code: CodeInvalid, Reason: "Prüfen und reparieren Sie den privaten Projektordner vor dem Start."}
}

func (s *Service) invalidBinding(binding domainort.Binding) Status {
	status := s.invalid()
	status.Configured = true
	status.Kind = &binding.Kind
	return status
}
