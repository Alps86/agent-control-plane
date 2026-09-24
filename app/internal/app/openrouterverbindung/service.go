package openrouterverbindung

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"agentcontrolplane/app/internal/port/credentials"
)

func NewService(store credentials.Store, probe Probe) *Service {
	return &Service{store: store, probe: probe}
}

func (service *Service) Save(ctx context.Context, key string) (View, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	key = strings.TrimSpace(key)
	if key == "" {
		return View{}, ErrMissingKey
	}

	entry := record{Key: key, Status: "nicht geprüft"}
	if service.persist(ctx, entry) != nil {
		return View{}, ErrStorage
	}

	return entry.view(), nil
}

func (service *Service) Status(ctx context.Context) (View, error) {
	entry, err := service.load(ctx)
	if errors.Is(err, credentials.ErrNotFound) {
		return View{Status: "nicht eingerichtet"}, nil
	}

	if err != nil {
		return View{}, ErrStorage
	}

	return entry.view(), nil
}

func (service *Service) Check(ctx context.Context) (View, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	entry, err := service.load(ctx)
	if errors.Is(err, credentials.ErrNotFound) {
		return View{Status: "nicht eingerichtet"}, nil
	}

	if err != nil {
		return View{}, ErrStorage
	}

	entry.Status = service.checkedStatus(ctx, entry.Key)
	if service.persist(ctx, entry) != nil {
		return View{}, ErrStorage
	}

	return entry.view(), nil
}

func (service *Service) Disconnect(ctx context.Context) (View, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	err := service.store.Delete(ctx, Reference)
	if err != nil && !errors.Is(err, credentials.ErrNotFound) {
		return View{}, ErrStorage
	}

	return View{Status: "nicht eingerichtet"}, nil
}

func (service *Service) checkedStatus(ctx context.Context, key string) string {
	if service.probe == nil {
		return "Prüfung derzeit nicht möglich"
	}

	result := service.probe.Check(ctx, key)
	if result == CheckReady {
		return "einsatzbereit"
	}

	if result == CheckInvalid {
		return "nicht einsatzbereit"
	}

	return "Prüfung derzeit nicht möglich"
}

func (service *Service) load(ctx context.Context) (record, error) {
	data, err := service.store.Load(ctx, Reference)
	if err != nil {
		return record{}, err
	}

	var entry record
	if json.Unmarshal(data, &entry) != nil || entry.Key == "" {
		return record{}, ErrStorage
	}

	return entry, nil
}

func (service *Service) persist(ctx context.Context, entry record) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return ErrStorage
	}

	return service.store.Save(ctx, Reference, data)
}

func (entry record) view() View {
	return View{Connected: true, Reference: Reference, Status: entry.Status, StatusDetail: entry.detail()}
}

func (entry record) detail() string {
	if entry.Status == "nicht einsatzbereit" {
		return "OpenRouter hat den Schlüssel abgelehnt. Ersetzen Sie ihn und prüfen Sie erneut."
	}

	if entry.Status == "Prüfung derzeit nicht möglich" {
		return "Die Verbindung konnte momentan nicht geprüft werden."
	}

	return ""
}
