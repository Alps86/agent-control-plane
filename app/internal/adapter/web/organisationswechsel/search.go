package organisationswechsel

import (
	"net/http"
	"strings"

	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
)

func (h *Handler) search(r *http.Request) ([]domainorganisation.Organization, error) {
	organizations, err := h.organizations.List(r.Context())
	if err != nil {
		return nil, err
	}

	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	items := make([]domainorganisation.Organization, 0, len(organizations))
	for _, organization := range organizations {
		if query == "" || strings.Contains(strings.ToLower(organization.Name), query) {
			items = append(items, organization)
		}
	}

	return items, nil
}

func (h *Handler) selectOrganization(w http.ResponseWriter, r *http.Request, id string) {
	http.SetCookie(w, &http.Cookie{
		Name: "acp_organisation", Value: id, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil,
	})
}

func (h *Handler) selectedOrganization(r *http.Request) string {
	cookie, err := r.Cookie("acp_organisation")
	if err != nil || cookie.Value == "" {
		return ""
	}

	organization, err := h.organizations.Get(r.Context(), cookie.Value)
	if err != nil {
		return ""
	}

	return organization.ID
}
