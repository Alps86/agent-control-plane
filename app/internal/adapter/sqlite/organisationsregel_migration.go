package sqlite

// OrganizationRulesMigration speichert versionierte Fachkanten je Organisation.
func OrganizationRulesMigration() Migration {
	return NewMigration(12, `CREATE TABLE organization_rules (
		organization_id TEXT PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
		revision INTEGER NOT NULL DEFAULT 0 CHECK (revision >= 0),
		rules_json TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(rules_json) AND json_type(rules_json) = 'array')
	)`)
}
