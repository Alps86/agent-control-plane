package rechte

import apprechte "agentcontrolplane/app/internal/app/rechte"

// Handler stellt den öffentlichen Beispiel-Leseweg bereit.
type Handler struct {
	reader *apprechte.Reader
}

// resourceResponse enthält nur die freigegebenen öffentlichen Felder.
type resourceResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
}
