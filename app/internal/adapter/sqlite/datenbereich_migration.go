package sqlite

// DataScopeMigration folgt auf das Projektschema der Version 6.
func DataScopeMigration() Migration {
	return NewMigration(7, `CREATE TABLE agent_project_scopes (
		agent_id TEXT NOT NULL REFERENCES agents(id),
		organization_id TEXT NOT NULL REFERENCES organizations(id),
		project_id TEXT NOT NULL REFERENCES projects(id),
		can_read INTEGER NOT NULL CHECK (can_read IN (0, 1)),
		can_write INTEGER NOT NULL CHECK (can_write IN (0, 1)),
		PRIMARY KEY (agent_id, project_id),
		CHECK (can_read = 1 OR can_write = 1)
	)`, `CREATE INDEX agent_project_scopes_project_idx ON agent_project_scopes(project_id, organization_id)`,
		`CREATE TABLE agent_data_denials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			action TEXT NOT NULL CHECK (action IN ('read_project', 'write_project_description')),
			occurred_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		)`, `CREATE INDEX agent_data_denials_lookup_idx ON agent_data_denials(organization_id, agent_id, id)`)
}
