package sqlite

// ProjectMigration ergänzt Projekte nach der Agentmigration.
func ProjectMigration() Migration {
	return NewMigration(6, `CREATE TABLE projects (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL REFERENCES organizations(id),
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		goal_id TEXT NOT NULL REFERENCES goals(id),
		CHECK (length(trim(name)) > 0),
		CHECK (length(trim(goal_id)) > 0)
	)`, `CREATE INDEX projects_organization_name ON projects (organization_id, name, id)`)
}
