package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	sqlite3 "modernc.org/sqlite"
)

// NewOrganizationStore bindet einen Organisationsspeicher an die geöffnete Datenbank.
func NewOrganizationStore(database *Database) *OrganizationStore {
	return &OrganizationStore{database: database}
}

// Create speichert Betreiberzuordnung und Standardregel in einer Transaktion.
func (s *OrganizationStore) Create(ctx context.Context, actorID string, organization domainorganisation.Organization) error {
	return s.database.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.insertOrganization(txCtx, actorID, organization)
	})
}

func (s *OrganizationStore) insertOrganization(ctx context.Context, actorID string, organization domainorganisation.Organization) error {
	policy, err := json.Marshal(organization.WorkflowPolicy)
	if err != nil {
		return fmt.Errorf("Organisationsregel kodieren: %w", err)
	}

	_, err = s.database.executor(ctx).ExecContext(ctx, `INSERT INTO organizations
		(id, operator_id, name, name_key, description, workflow_policy)
		VALUES (?, ?, ?, ?, ?, ?)`, organization.ID, actorID, organization.Name,
		s.nameKey(organization.Name), organization.Description, string(policy))
	return s.organizationWriteError(err)
}

func (s *OrganizationStore) nameKey(name string) string {
	var key strings.Builder
	for _, character := range name {
		minimum := character
		for folded := unicode.SimpleFold(character); folded != character; folded = unicode.SimpleFold(folded) {
			if folded < minimum {
				minimum = folded
			}
		}

		key.WriteRune(minimum)
	}

	return key.String()
}

func (s *OrganizationStore) organizationWriteError(err error) error {
	if err == nil {
		return nil
	}

	var sqliteErr *sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == 2067 {
		return portorganisation.ErrNameConflict
	}

	return fmt.Errorf("Organisation speichern: %w", err)
}

// List liest ausschließlich Organisationen des gegebenen Betreibers.
func (s *OrganizationStore) List(ctx context.Context, actorID string) ([]domainorganisation.Organization, error) {
	rows, err := s.database.db.QueryContext(ctx, `SELECT id, name, description, workflow_policy
		FROM organizations WHERE operator_id = ? ORDER BY name_key, id`, actorID)
	if err != nil {
		return nil, fmt.Errorf("Organisationen lesen: %w", err)
	}

	defer rows.Close()
	return s.collectOrganizations(rows)
}

func (s *OrganizationStore) collectOrganizations(rows *sql.Rows) ([]domainorganisation.Organization, error) {
	organizations := []domainorganisation.Organization{}
	for rows.Next() {
		organization, err := s.scanOrganization(rows)
		if err != nil {
			return nil, err
		}

		organizations = append(organizations, organization)
	}

	return organizations, rows.Err()
}

// Get liest eine Organisation nur mit passender Betreiberzuordnung.
func (s *OrganizationStore) Get(ctx context.Context, id, actorID string) (domainorganisation.Organization, error) {
	row := s.database.executor(ctx).QueryRowContext(ctx, `SELECT id, name, description, workflow_policy
		FROM organizations WHERE id = ? AND operator_id = ?`, id, actorID)
	return s.scanOrganization(row)
}

func (s *OrganizationStore) scanOrganization(row interface{ Scan(...any) error }) (domainorganisation.Organization, error) {
	var organization domainorganisation.Organization
	var policy string
	err := row.Scan(&organization.ID, &organization.Name, &organization.Description, &policy)
	if errors.Is(err, sql.ErrNoRows) {
		return domainorganisation.Organization{}, portorganisation.ErrNotFound
	}

	if err != nil {
		return domainorganisation.Organization{}, fmt.Errorf("Organisation lesen: %w", err)
	}

	return s.decodeOrganization(organization, policy)
}

func (s *OrganizationStore) decodeOrganization(organization domainorganisation.Organization, policy string) (domainorganisation.Organization, error) {
	if err := json.Unmarshal([]byte(policy), &organization.WorkflowPolicy); err != nil {
		return domainorganisation.Organization{}, fmt.Errorf("Organisationsregel lesen: %w", err)
	}

	return organization, nil
}
