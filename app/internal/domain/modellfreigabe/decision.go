package modellfreigabe

// Error erlaubt, eine verweigerte Entscheidung unmittelbar zurückzugeben.
func (d Decision) Error() string { return d.Reason }

// Decide prüft die drei unabhängigen Voraussetzungen ohne implizite Freigabe.
func (s Status) Decide() Decision {
	if !s.Connected || s.ConnectionStatus != "einsatzbereit" {
		return Decision{Reason: ReasonConnection}
	}
	if !s.OrganizationGranted {
		return Decision{Reason: ReasonOrganization}
	}
	if !s.AgentGranted {
		return Decision{Reason: ReasonAgent}
	}
	return Decision{Allowed: true}
}
