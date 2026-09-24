package sqlite

// GoalMigration ergänzt Stammziele nach dem Organisationsschema.
func GoalMigration() Migration {
	return NewMigration(4, `CREATE TABLE goals (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL REFERENCES organizations(id),
		name TEXT NOT NULL,
		parent_goal_id TEXT REFERENCES goals(id),
		CHECK (length(trim(name)) > 0)
	)`, `CREATE INDEX goals_organization_name ON goals (organization_id, name, id)`)
}
