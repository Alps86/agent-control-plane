package sqlite

// TaskMigration ergänzt Agentenstatus und organisationsgebundene Aufgaben.
func TaskMigration() Migration {
	return NewMigration(9,
		`ALTER TABLE agents ADD COLUMN status TEXT NOT NULL DEFAULT 'active'
			CHECK (status IN ('active', 'paused', 'ended'))`,
		`CREATE UNIQUE INDEX projects_org_id_unique ON projects (organization_id, id)`,
		`CREATE UNIQUE INDEX agents_org_id_unique ON agents (organization_id, id)`,
		`CREATE TABLE tasks (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL REFERENCES organizations(id),
			project_id TEXT NOT NULL,
			title TEXT NOT NULL CHECK (length(trim(title)) > 0),
			description TEXT NOT NULL,
			priority TEXT NOT NULL CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
			assignee_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'open' CHECK (status = 'open'),
			FOREIGN KEY (organization_id, project_id) REFERENCES projects (organization_id, id),
			FOREIGN KEY (organization_id, assignee_id) REFERENCES agents (organization_id, id)
		)`,
		`CREATE INDEX tasks_org_project_idx ON tasks (organization_id, project_id, title, id)`)
}
