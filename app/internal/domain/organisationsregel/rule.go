package organisationsregel

import "strings"

// Validate verwirft unbekannte Kanten, fehlende Freigabe und freie Prozesssprache.
func (r Rule) Validate() error {
	if r.expression() {
		return ErrWorkflowDSL
	}

	if r.Approver == "" {
		return ErrApproverRequired
	}

	if r.Approver != "betreiber" {
		return ErrRightsExpansion
	}

	if !r.allowed() {
		return ErrInvalidTransition
	}

	return nil
}

func (r Rule) expression() bool {
	for _, value := range []string{r.Area, r.From, r.To, r.Approver} {
		if strings.ContainsAny(value, " =>;:{}()|&!\n\t") {
			return true
		}
	}

	return false
}

func (r Rule) allowed() bool {
	if r.Area == "status" {
		return r.statusAllowed()
	}

	if r.Area == "assignment" {
		return r.assignmentAllowed()
	}

	if r.Area == "delegation" {
		return r.delegationAllowed()
	}

	return false
}

func (r Rule) statusAllowed() bool {
	allowed := map[string]bool{
		"todo:in_progress": true, "todo:cancelled": true,
		"in_progress:blocked": true, "in_progress:in_review": true,
		"in_progress:done": true, "in_progress:cancelled": true,
		"blocked:in_progress": true, "blocked:cancelled": true,
		"in_review:in_progress": true, "in_review:done": true,
	}
	return allowed[r.From+":"+r.To]
}

func (r Rule) assignmentAllowed() bool {
	return r.From == "unassigned" && r.To == "assigned" ||
		r.From == "assigned" && r.To == "unassigned"
}

func (r Rule) delegationAllowed() bool {
	return r.From == "requested" && (r.To == "approved" || r.To == "rejected") ||
		r.From == "approved" && r.To == "completed"
}

// WithRule ersetzt genau eine Kante und belässt andere Regelbereiche unverändert.
func (p Policy) WithRule(rule Rule) Policy {
	next := make([]Rule, 0, len(p.Rules)+1)
	for _, existing := range p.Rules {
		if existing.Area == rule.Area && existing.From == rule.From && existing.To == rule.To {
			continue
		}

		next = append(next, existing)
	}

	p.Rules = append(next, rule)
	return p
}

// Allows nennt nur die konfigurierte Freigabe; aktuelle Rechte bleiben gesondert zu prüfen.
func (p Policy) Allows(rule Rule) bool {
	for _, existing := range p.Rules {
		if existing == rule {
			return true
		}
	}

	return false
}

// WithoutRule entfernt nur eine vorhandene, genau benannte Fachkante.
func (p Policy) WithoutRule(rule Rule) (Policy, error) {
	if !p.Allows(rule) {
		return Policy{}, ErrRuleNotFound
	}
	next := make([]Rule, 0, len(p.Rules)-1)
	for _, existing := range p.Rules {
		if existing != rule {
			next = append(next, existing)
		}
	}
	p.Rules = next
	return p, nil
}
