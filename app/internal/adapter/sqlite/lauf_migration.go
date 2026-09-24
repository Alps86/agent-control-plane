package sqlite

// RunMigration ergänzt die dauerhafte Laufreservierung als Fachschema.
func RunMigration(version int) Migration {
	return NewMigration(version,
		`CREATE TABLE lauf_reservations (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('Reserviert', 'Fehlgeschlagen')),
			failure_reason TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE UNIQUE INDEX lauf_one_active_task
			ON lauf_reservations (organization_id, task_id)
			WHERE status = 'Reserviert'`,
		`CREATE INDEX lauf_task_history
			ON lauf_reservations (organization_id, task_id)`,
	)
}
