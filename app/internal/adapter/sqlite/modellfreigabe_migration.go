package sqlite

// ModellfreigabeMigration speichert ausschließlich explizite OpenRouter-Rechte.
func ModellfreigabeMigration() Migration {
	return NewMigration(11, `CREATE TABLE openrouter_organization_grants (
		organization_id TEXT PRIMARY KEY REFERENCES organizations(id),
		reference TEXT NOT NULL,
		generation TEXT NOT NULL,
		granted_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`, `CREATE TABLE openrouter_agent_grants (
		agent_id TEXT PRIMARY KEY REFERENCES agents(id),
		organization_id TEXT NOT NULL REFERENCES organizations(id),
		reference TEXT NOT NULL,
		generation TEXT NOT NULL,
		granted_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`, `CREATE INDEX openrouter_agent_grants_organization_idx
		ON openrouter_agent_grants(organization_id)`)
}
