package datenbereiche

import (
	"context"

	"agentcontrolplane/app/internal/adapter/sqlite"
	appdatenbereich "agentcontrolplane/app/internal/app/datenbereich"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	"agentcontrolplane/app/internal/domain/rechte"
)

func (i FixedIdentity) Actors() []rechte.Actor {
	return []rechte.Actor{i.Actor}
}

func (s *Suite) openStore() error {
	if s.dataStore != nil {
		return nil
	}

	store, err := sqlite.OpenWithMigrations(context.Background(), s.database,
		sqlite.RunMigration(2), sqlite.OrganizationMigration(), sqlite.GoalMigration(),
		sqlite.AgentMigration(), sqlite.ProjectMigration(), sqlite.DataScopeMigration(), sqlite.CodexProfileMigration())
	if err != nil {
		return err
	}

	s.dataStore = store
	return nil
}

func (s *Suite) openDataFacade() error {
	if err := s.openStore(); err != nil {
		return err
	}
	operator := apporganisation.NewLocalIdentity()
	s.dataService = appdatenbereich.NewService(s.dataStore, operator, FixedIdentity{Actor: s.agentActor})
	return nil
}

func (s *Suite) closeDataFacade() {
	s.dataService = nil
	if s.dataStore != nil {
		_ = s.dataStore.Close()
		s.dataStore = nil
	}
}

func (s *Suite) requestScope(method, path, body, contentType string) error {
	if s.dataService == nil {
		if err := s.openDataFacade(); err != nil {
			return err
		}
	}

	return s.request(method, path, body, contentType)
}
