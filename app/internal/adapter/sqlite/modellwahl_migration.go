package sqlite

// ModelChoiceMigration adds one credential-free route per existing agent.
// Register only after versions 9 through 13 are in the server chain.
func ModelChoiceMigration() Migration {
	return NewMigration(14, `CREATE TABLE agent_model_selections (
		agent_id TEXT PRIMARY KEY REFERENCES agents(id) ON DELETE CASCADE,
		provider_id TEXT NOT NULL,
		connection_id TEXT NOT NULL,
		model_id TEXT NOT NULL,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (length(trim(provider_id)) > 0),
		CHECK (length(trim(connection_id)) > 0),
		CHECK (length(trim(model_id)) > 0)
	)`)
}
