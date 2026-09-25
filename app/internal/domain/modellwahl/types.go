package modellwahl

// Selection is an agent's explicit, secret-free model route.
type Selection struct {
	Provider            string `json:"provider_id"`
	ConnectionReference string `json:"connection_id"`
	Model               string `json:"model_id"`
}

// Capability records one claim and its evidence separately.
type Capability struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	CheckedAt string `json:"checked_at"`
}

// Model is a configured candidate, never an implicit account entitlement.
type Model struct {
	ID              string                `json:"id"`
	Label           string                `json:"label"`
	Source          string                `json:"source"`
	ObservedAt      string                `json:"observed_at"`
	CheckStatus     string                `json:"check_status"`
	AccountVerified bool                  `json:"account_verified"`
	RouteVerified   bool                  `json:"route_verified"`
	Capabilities    map[string]Capability `json:"capabilities"`
}

// Connection names a public reference without its credential.
type Connection struct {
	Reference string `json:"reference"`
	Label     string `json:"label"`
}

// Provider is a registered provider and its configured model candidates.
type Provider struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AuthType    string       `json:"auth_type"`
	Selectable  bool         `json:"selectable"`
	Reason      string       `json:"reason"`
	Connections []Connection `json:"connections"`
	Models      []Model      `json:"models"`
}

// Catalog contains no secrets or account availability claims.
type Catalog struct {
	Providers []Provider `json:"providers"`
}
