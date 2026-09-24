package openrouterverbindung

import (
	"context"
	"errors"

	"agentcontrolplane/app/internal/port/credentials"
	"agentcontrolplane/app/internal/port/modellschluessel"
)

func (service *Service) Binding(ctx context.Context) (modellschluessel.Binding, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	entry, err := service.load(ctx)
	if errors.Is(err, credentials.ErrNotFound) {
		return modellschluessel.Binding{}, modellschluessel.ErrMissing
	}

	if err != nil {
		return modellschluessel.Binding{}, ErrStorage
	}

	return modellschluessel.Binding{Reference: Reference, Generation: entry.Generation}, nil
}

func (service *Service) Resolve(ctx context.Context, binding modellschluessel.Binding) (string, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	entry, err := service.load(ctx)
	if errors.Is(err, credentials.ErrNotFound) {
		return "", modellschluessel.ErrMissing
	}

	if err != nil {
		return "", ErrStorage
	}

	return entry.resolve(binding)
}

func (entry record) resolve(binding modellschluessel.Binding) (string, error) {
	if binding.Reference != Reference || binding.Generation == "" || binding.Generation != entry.Generation {
		return "", modellschluessel.ErrRevoked
	}

	if entry.Status == "nicht einsatzbereit" {
		return "", modellschluessel.ErrRevoked
	}

	if entry.Status != "einsatzbereit" {
		return "", modellschluessel.ErrNotReady
	}

	return entry.Key, nil
}
