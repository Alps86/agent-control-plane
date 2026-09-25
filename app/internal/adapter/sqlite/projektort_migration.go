package sqlite

// ProjectLocationMigration ergänzt explizite private Projektort-Bindungen.
func ProjectLocationMigration() Migration {
	return NewMigration(13, `CREATE TABLE project_locations (
		project_id TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
		organization_id TEXT NOT NULL REFERENCES organizations(id),
		kind TEXT NOT NULL CHECK (kind = 'managed_directory')
	)`, `CREATE INDEX project_locations_organization_idx ON project_locations (organization_id, project_id)`)
}
