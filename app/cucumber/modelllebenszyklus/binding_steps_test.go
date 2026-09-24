package modelllebenszyklus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	credentialadapter "agentcontrolplane/app/internal/adapter/credentials"
	modelopenrouter "agentcontrolplane/app/internal/adapter/model/openrouter"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	"agentcontrolplane/app/internal/port/modellschluessel"
	"github.com/cucumber/godog"
)

func (s *Suite) registerBinding(sc *godog.ScenarioContext) {
	sc.Step(`^eine geprüfte OpenRouter-Verbindung ist serverseitig gebunden$`, s.boundRouter)
	sc.Step(`^ich ihren Schlüssel bewusst ersetze und erneut prüfe$`, s.rotateBoundRouter)
	sc.Step(`^bleibt dieselbe Bindung mit dem neuen Schlüssel auflösbar$`, s.sameBinding)
	sc.Step(`^ich den geschützten Speicher und den Verbindungsdienst neu öffne$`, s.reopenRouter)
	sc.Step(`^ich die Verbindung trenne und mit einem neuen Schlüssel erneut einrichte und prüfe$`, s.reconnectRouter)
	sc.Step(`^ist die alte Bindung endgültig nicht mehr auflösbar$`, s.oldBindingRevoked)
	sc.Step(`^die neue Bindung löst nur den neuen Schlüssel auf$`, s.newBinding)
	sc.Step(`^eine vor der Bindungseinführung geschützte und geprüfte OpenRouter-Verbindung besteht$`, s.legacyRouter)
	sc.Step(`^ich den Verbindungsdienst mit demselben geschützten Speicher neu öffne$`, s.reopenRouter)
	sc.Step(`^kann ich die bestehende Verbindung binden und ihren Schlüssel auflösen$`, s.legacyBinding)
	sc.Step(`^dieselbe Bindung bleibt nach erneutem Öffnen des Dienstes gültig$`, s.reopenedLegacyBinding)
}

func (s *Suite) boundRouter() error {
	if err := s.setup(); err != nil {
		return err
	}
	if _, err := s.routerService.Save(context.Background(), boundKey); err != nil {
		return err
	}
	if _, err := s.routerService.Check(context.Background()); err != nil {
		return err
	}
	binding, err := s.routerService.Binding(context.Background())
	s.binding = binding
	return err
}

func (s *Suite) rotateBoundRouter() error {
	if _, err := s.routerService.Save(context.Background(), newKey); err != nil {
		return err
	}
	_, err := s.routerService.Check(context.Background())
	return err
}

func (s *Suite) sameBinding() error {
	current, err := s.routerService.Binding(context.Background())
	if err != nil {
		return err
	}
	key, err := s.routerService.Resolve(context.Background(), s.binding)
	if current != s.binding || err != nil || key != newKey {
		return fmt.Errorf("Rotation oder Neustart änderte Bindung")
	}
	return nil
}

func (s *Suite) reopenRouter() error {
	store, err := credentialadapter.NewStore(s.secretPath, s.keyPath)
	if err != nil {
		return err
	}
	s.store = store
	probe := modelopenrouter.NewProbeAt(nil, s.provider.URL+"/api/v1/key")
	s.routerService = openrouterverbindung.NewService(store, probe)
	return nil
}

func (s *Suite) reconnectRouter() error {
	if _, err := s.routerService.Disconnect(context.Background()); err != nil {
		return err
	}
	if _, err := s.routerService.Save(context.Background(), newKey); err != nil {
		return err
	}
	_, err := s.routerService.Check(context.Background())
	return err
}

func (s *Suite) oldBindingRevoked() error {
	key, err := s.routerService.Resolve(context.Background(), s.binding)
	if key != "" || !errors.Is(err, modellschluessel.ErrRevoked) {
		return fmt.Errorf("alte Bindung nach Neuverbindung nicht widerrufen: %v", err)
	}
	return nil
}

func (s *Suite) newBinding() error {
	current, err := s.routerService.Binding(context.Background())
	if err != nil {
		return err
	}
	key, err := s.routerService.Resolve(context.Background(), current)
	if current == s.binding || err != nil || key != newKey {
		return fmt.Errorf("neue Bindung nicht auf neuen Schlüssel begrenzt")
	}
	return nil
}

func (s *Suite) legacyRouter() error {
	if err := s.setup(); err != nil {
		return err
	}
	fixture, err := json.Marshal(map[string]string{"key": newKey, "status": "einsatzbereit"})
	if err != nil {
		return err
	}
	return s.store.Save(context.Background(), openrouterverbindung.Reference, fixture)
}

func (s *Suite) legacyBinding() error {
	binding, err := s.routerService.Binding(context.Background())
	if err != nil {
		return err
	}
	s.binding = binding
	key, err := s.routerService.Resolve(context.Background(), binding)
	if err != nil || key != newKey {
		return fmt.Errorf("geschützte Bestandsverbindung nicht auflösbar: %v", err)
	}
	return nil
}

func (s *Suite) reopenedLegacyBinding() error {
	if err := s.reopenRouter(); err != nil {
		return err
	}
	key, err := s.routerService.Resolve(context.Background(), s.binding)
	if err != nil || key != newKey {
		return fmt.Errorf("migrierte Bindung nach Neustart verloren: %v", err)
	}
	return nil
}
