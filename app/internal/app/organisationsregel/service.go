package organisationsregel

import (
	"context"
	"errors"

	domainregel "agentcontrolplane/app/internal/domain/organisationsregel"
	"agentcontrolplane/app/internal/domain/rechte"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	portregel "agentcontrolplane/app/internal/port/organisationsregel"
)

// NewService bindet Identität, Organisationsmitgliedschaft und Regelspeicher.
func NewService(organizations portorganisation.Store, identity portorganisation.Identity, store portregel.Store) *Service {
	return &Service{organizations: organizations, identity: identity, store: store}
}

// Read liefert nur die Regeln einer zugeordneten Organisation.
func (s *Service) Read(ctx context.Context, organizationID string) (Snapshot, error) {
	actorID, err := s.authorized(ctx, organizationID)
	if err != nil {
		return Snapshot{}, err
	}

	policy, err := s.store.Read(ctx, organizationID, actorID)
	return s.snapshot(organizationID, policy), err
}

// Preview prüft die Kante und nennt nur ihren betroffenen Fachablauf.
func (s *Service) Preview(ctx context.Context, organizationID string, input Change) (Preview, error) {
	current, err := s.Read(ctx, organizationID)
	if err != nil {
		return Preview{}, err
	}

	if input.Revision != current.Revision {
		return Preview{}, ErrRevisionConflict
	}

	operation, err := input.operation()
	if err != nil {
		return Preview{}, err
	}
	rule := input.rule()
	if err := rule.Validate(); err != nil {
		return Preview{}, err
	}
	if operation == "revoke" && !s.policy(current).Allows(rule) {
		return Preview{}, ErrRuleNotFound
	}

	return Preview{Operation: operation, Area: rule.Area, AffectedFlows: []string{s.flow(rule)}, Rule: rule, Revision: current.Revision}, nil
}

// Save speichert nur die validierte Kante mit atomarem Revisionsvergleich.
func (s *Service) Save(ctx context.Context, organizationID string, input Change) (Snapshot, error) {
	actorID, err := s.authorized(ctx, organizationID)
	if err != nil {
		return Snapshot{}, err
	}

	operation, err := input.operation()
	if err != nil {
		return Snapshot{}, err
	}
	rule := input.rule()
	if err := rule.Validate(); err != nil {
		return Snapshot{}, err
	}

	policy, err := s.store.SaveRule(ctx, organizationID, actorID, input.Revision, operation, rule)
	return s.snapshot(organizationID, policy), err
}

func (s *Service) authorized(ctx context.Context, organizationID string) (string, error) {
	if s == nil || s.organizations == nil || s.identity == nil || s.store == nil {
		return "", ErrAccessDenied
	}

	actors := s.identity.Actors()
	if len(actors) != 1 || !actors[0].Valid() || actors[0].Kind != rechte.Operator {
		return "", ErrAccessDenied
	}

	_, err := s.organizations.Get(ctx, organizationID, actors[0].ID)
	if errors.Is(err, portorganisation.ErrNotFound) {
		return "", ErrNotFound
	}

	return actors[0].ID, err
}

func (s *Service) snapshot(organizationID string, policy domainregel.Policy) Snapshot {
	rules := append([]Rule{}, policy.Rules...)
	return Snapshot{OrganizationID: organizationID, Revision: policy.Revision, Areas: []string{"status", "assignment", "delegation"}, Rules: rules}
}

func (s *Service) flow(rule Rule) string {
	if rule.Area == "status" {
		return "Aufgabenstatus: " + rule.From + " → " + rule.To
	}

	if rule.Area == "assignment" {
		return "Zuweisung: " + rule.From + " → " + rule.To
	}

	return "Delegation: " + rule.From + " → " + rule.To
}

func (c Change) rule() Rule {
	approver := c.Approver
	if c.Operation == "revoke" && approver == "" {
		approver = "betreiber"
	}
	return Rule{Area: c.Area, From: c.From, To: c.To, Approver: approver}
}

func (c Change) operation() (string, error) {
	if c.Operation == "" || c.Operation == "allow" {
		return "allow", nil
	}
	if c.Operation == "revoke" {
		return "revoke", nil
	}
	return "", ErrInvalidOperation
}

func (s *Service) policy(snapshot Snapshot) domainregel.Policy {
	return domainregel.Policy{Revision: snapshot.Revision, Rules: snapshot.Rules}
}
