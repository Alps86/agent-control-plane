package sqlite

// AgentMigration hält das Agentenprofil nach Organisation und Ziel bereit.
func AgentMigration() Migration {
	return NewMigration(5, `CREATE TABLE agents (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL REFERENCES organizations(id),
		name TEXT NOT NULL,
		name_key TEXT NOT NULL,
		role TEXT NOT NULL,
		instructions TEXT NOT NULL,
		execution_kind TEXT NOT NULL CHECK (execution_kind IN ('eino', 'codex_cli')),
		template_id TEXT NOT NULL,
		capabilities TEXT NOT NULL CHECK (json_valid(capabilities)),
		UNIQUE (organization_id, name_key),
		CHECK (length(trim(name)) > 0)
	)`, `CREATE INDEX agents_organization_idx ON agents(organization_id)`)
}
