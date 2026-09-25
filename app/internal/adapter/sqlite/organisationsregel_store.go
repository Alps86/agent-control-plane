package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	domainregel "agentcontrolplane/app/internal/domain/organisationsregel"
	portregel "agentcontrolplane/app/internal/port/organisationsregel"
)

// Read liest die versionierten Regeln ausschließlich unter Betreiberzuordnung.
func (d *Database) Read(ctx context.Context, organizationID, operatorID string) (domainregel.Policy, error) {
	if err := d.ruleOrganization(ctx, organizationID, operatorID); err != nil {
		return domainregel.Policy{}, err
	}

	return d.readRules(ctx, organizationID)
}

func (d *Database) ruleOrganization(ctx context.Context, organizationID, operatorID string) error {
	var present int
	err := d.executor(ctx).QueryRowContext(ctx, `SELECT 1 FROM organizations
		WHERE id = ? AND operator_id = ?`, organizationID, operatorID).Scan(&present)
	if errors.Is(err, sql.ErrNoRows) {
		return portregel.ErrNotFound
	}

	return err
}

func (d *Database) readRules(ctx context.Context, organizationID string) (domainregel.Policy, error) {
	var policy domainregel.Policy
	var encoded string
	err := d.executor(ctx).QueryRowContext(ctx, `SELECT revision, rules_json
		FROM organization_rules WHERE organization_id = ?`, organizationID).Scan(&policy.Revision, &encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return domainregel.Policy{Rules: []domainregel.Rule{}}, nil
	}

	if err != nil {
		return domainregel.Policy{}, fmt.Errorf("Organisationsregeln lesen: %w", err)
	}

	return d.decodeRules(policy, encoded)
}

func (d *Database) decodeRules(policy domainregel.Policy, encoded string) (domainregel.Policy, error) {
	if err := json.Unmarshal([]byte(encoded), &policy.Rules); err != nil {
		return domainregel.Policy{}, fmt.Errorf("Organisationsregeln dekodieren: %w", err)
	}

	return policy, nil
}

// SaveRule ersetzt eine Kante atomar und lehnt veraltete Revisionen ab.
func (d *Database) SaveRule(ctx context.Context, organizationID, operatorID string, revision int64, operation string, rule domainregel.Rule) (domainregel.Policy, error) {
	var saved domainregel.Policy
	err := d.WithinTransaction(ctx, func(txCtx context.Context) error {
		var saveErr error
		saved, saveErr = d.saveRuleIn(txCtx, organizationID, operatorID, revision, operation, rule)
		return saveErr
	})
	if err != nil {
		return domainregel.Policy{}, err
	}

	return saved, err
}

func (d *Database) saveRuleIn(ctx context.Context, organizationID, operatorID string, revision int64, operation string, rule domainregel.Rule) (domainregel.Policy, error) {
	current, err := d.Read(ctx, organizationID, operatorID)
	if err != nil {
		return domainregel.Policy{}, err
	}

	if revision != current.Revision || revision < 0 {
		return domainregel.Policy{}, portregel.ErrRevisionConflict
	}

	next, err := d.nextPolicy(current, operation, rule)
	if err != nil {
		return domainregel.Policy{}, err
	}

	next.Revision++
	if err := d.writeRules(ctx, organizationID, revision, next); err != nil {
		return domainregel.Policy{}, err
	}

	return next, d.projectRulePolicy(ctx, organizationID, operatorID, next)
}

func (d *Database) nextPolicy(current domainregel.Policy, operation string, rule domainregel.Rule) (domainregel.Policy, error) {
	if operation == "allow" {
		return current.WithRule(rule), nil
	}

	if operation == "revoke" {
		return current.WithoutRule(rule)
	}

	return domainregel.Policy{}, domainregel.ErrInvalidOperation
}

func (d *Database) projectRulePolicy(ctx context.Context, organizationID, operatorID string, policy domainregel.Policy) error {
	work := []string{}
	delegation := []string{}
	for _, rule := range policy.Rules {
		edge := rule.From + " → " + rule.To + " (" + rule.Approver + ")"
		if rule.Area == "status" {
			work = append(work, edge)
		}

		if rule.Area == "delegation" {
			delegation = append(delegation, edge)
		}
	}

	return d.updateRuleProjection(ctx, organizationID, operatorID, work, delegation)
}

func (d *Database) updateRuleProjection(ctx context.Context, organizationID, operatorID string, work, delegation []string) error {
	projection, err := json.Marshal(map[string][]string{
		"work_transitions": work, "delegation_transitions": delegation,
	})
	if err != nil {
		return fmt.Errorf("Organisationsansicht kodieren: %w", err)
	}

	result, err := d.executor(ctx).ExecContext(ctx, `UPDATE organizations SET workflow_policy = ?
		WHERE id = ? AND operator_id = ?`, string(projection), organizationID, operatorID)
	return d.ruleWriteResult(result, err)
}

func (d *Database) writeRules(ctx context.Context, organizationID string, revision int64, policy domainregel.Policy) error {
	encoded, err := json.Marshal(policy.Rules)
	if err != nil {
		return fmt.Errorf("Organisationsregeln kodieren: %w", err)
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO organization_rules
		(organization_id, revision, rules_json) VALUES (?, ?, ?)
		ON CONFLICT(organization_id) DO UPDATE SET revision = excluded.revision,
		rules_json = excluded.rules_json WHERE organization_rules.revision = ?`,
		organizationID, policy.Revision, string(encoded), revision)
	return d.ruleWriteResult(result, err)
}

func (d *Database) ruleWriteResult(result sql.Result, err error) error {
	if err != nil {
		return fmt.Errorf("Organisationsregeln speichern: %w", err)
	}

	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if changed == 0 {
		return portregel.ErrRevisionConflict
	}

	return nil
}
