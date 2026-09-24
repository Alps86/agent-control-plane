package sqlite

// OrganizationMigration hält die nächste Fachmigration nach ARC-04 bereit.
func OrganizationMigration() Migration {
	return NewMigration(3, `CREATE TABLE organizations (
		id TEXT PRIMARY KEY,
		operator_id TEXT NOT NULL,
		name TEXT NOT NULL,
		name_key TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		workflow_policy TEXT NOT NULL,
		CHECK (length(trim(name)) > 0),
		CHECK (length(operator_id) > 0),
		CHECK (json_valid(workflow_policy))
	)`)
}
