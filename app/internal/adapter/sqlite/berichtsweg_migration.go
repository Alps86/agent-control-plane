package sqlite

// BerichtswegMigration speichert genau einen Vorgesetzten je Agent und Organisation.
func BerichtswegMigration() Migration {
	return NewMigration(18, `CREATE TABLE agent_reporting_lines (
		organization_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		parent_agent_id TEXT NOT NULL,
		PRIMARY KEY (organization_id, agent_id),
		FOREIGN KEY (organization_id, agent_id) REFERENCES agents (organization_id, id),
		FOREIGN KEY (organization_id, parent_agent_id) REFERENCES agents (organization_id, id),
		CHECK (agent_id <> parent_agent_id)
	)`, `CREATE INDEX agent_reporting_parent_idx
		ON agent_reporting_lines (organization_id, parent_agent_id)`)
}
