package sqlite

// CodexProfileMigration ergänzt ausschließlich betreiberseitige CLI-Freigaben.
func CodexProfileMigration() Migration {
	return NewMigration(8, `CREATE TABLE codex_profiles (
		agent_id TEXT PRIMARY KEY REFERENCES agents(id) ON DELETE CASCADE,
		workspace_enabled INTEGER NOT NULL DEFAULT 0 CHECK (workspace_enabled IN (0, 1)),
		write_enabled INTEGER NOT NULL DEFAULT 0 CHECK (write_enabled IN (0, 1)),
		CHECK (write_enabled = 0 OR workspace_enabled = 1)
	)`)
}
