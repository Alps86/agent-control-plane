package sqlite

// ProjectArchiveMigration schützt neue Arbeit am tatsächlichen SQLite-Schreibpunkt.
func ProjectArchiveMigration() Migration {
	return NewMigration(16,
		`CREATE TABLE project_archives (
			organization_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			archived_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (organization_id, project_id),
			FOREIGN KEY (organization_id, project_id)
				REFERENCES projects (organization_id, id)
		)`,
		`CREATE TRIGGER project_archive_blocks_task
			BEFORE INSERT ON tasks
			WHEN EXISTS (SELECT 1 FROM project_archives a
				WHERE a.organization_id = NEW.organization_id
				AND a.project_id = NEW.project_id)
			BEGIN SELECT RAISE(ABORT, 'project_archived'); END`,
		`CREATE TRIGGER project_archive_blocks_run
			BEFORE INSERT ON lauf_reservations
			WHEN EXISTS (SELECT 1 FROM tasks t
				JOIN project_archives a ON a.organization_id = t.organization_id
					AND a.project_id = t.project_id
				WHERE t.organization_id = NEW.organization_id AND t.id = NEW.task_id)
			BEGIN SELECT RAISE(ABORT, 'project_archived'); END`)
}
