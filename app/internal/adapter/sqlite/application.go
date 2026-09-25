package sqlite

import "context"

// OpenApplication öffnet den Anwendungsbestand mit allen aktuellen Migrationen.
func OpenApplication(ctx context.Context, path string) (*Database, error) {
	steps := []Migration{
		RunMigration(2),
		OrganizationMigration(),
		GoalMigration(),
		AgentMigration(),
		ProjectMigration(),
		DataScopeMigration(),
		CodexProfileMigration(),
		TaskMigration(),
		GoalTreeMigration(),
		ModellfreigabeMigration(),
		OrganizationRulesMigration(),
	}

	return OpenWithMigrations(ctx, path, steps...)
}
