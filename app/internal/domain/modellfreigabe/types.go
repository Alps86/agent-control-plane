package modellfreigabe

// Decision ist die schlüsselfreie Entscheidung an einer Modellzugangsgrenze.
type Decision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// Status zeigt nur Referenz und wirksame Freigaben.
type Status struct {
	Reference           string `json:"reference"`
	Connected           bool   `json:"connected"`
	ConnectionStatus    string `json:"connection_status"`
	OrganizationGranted bool   `json:"organization_granted"`
	AgentGranted        bool   `json:"agent_granted"`
	Allowed             bool   `json:"allowed"`
	Reason              string `json:"reason,omitempty"`
}

const (
	ReasonConnection   = "OpenRouter-Verbindung nicht einsatzbereit"
	ReasonOrganization = "Organisationsfreigabe fehlt"
	ReasonAgent        = "Agentenfreigabe fehlt"
	ReasonReference    = "Verbindungsreferenz ungültig"
)
