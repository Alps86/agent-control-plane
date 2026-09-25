package aktivitaet

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	appaktivitaet "agentcontrolplane/app/internal/app/aktivitaet"
	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
)

func (h *Handler) queryFilter(r *http.Request) (domainaktivitaet.Filter, bool, error) {
	values := r.URL.Query()
	filter := domainaktivitaet.Filter{}
	allowed := map[string]bool{"agent": true, "action": true, "from": true, "to": true, "object": true, "limit": true, "offset": true}
	for key := range values {
		if !allowed[key] || len(values[key]) != 1 || values.Get(key) != "" && strings.TrimSpace(values.Get(key)) == "" {
			return filter, false, appaktivitaet.ErrInvalidFilter
		}
	}

	filter.AgentID, filter.Action = values.Get("agent"), values.Get("action")
	filter.From, filter.To, filter.ObjectID = values.Get("from"), values.Get("to"), values.Get("object")
	if err := h.validateTimes(filter); err != nil {
		return filter, false, err
	}

	return h.queryPage(values, filter)
}

func (h *Handler) validateTimes(filter domainaktivitaet.Filter) error {
	var from, to time.Time
	var err error
	if filter.From != "" {
		from, err = time.Parse(time.RFC3339, filter.From)
		if err != nil {
			return appaktivitaet.ErrInvalidFilter
		}
	}

	if filter.To != "" {
		to, err = time.Parse(time.RFC3339, filter.To)
	}

	if err != nil || !from.IsZero() && !to.IsZero() && from.After(to) {
		return appaktivitaet.ErrInvalidFilter
	}

	return nil
}

func (h *Handler) queryPage(values url.Values, filter domainaktivitaet.Filter) (domainaktivitaet.Filter, bool, error) {
	var err error
	if values.Get("limit") != "" {
		filter.Limit, err = strconv.Atoi(values.Get("limit"))
		if err != nil || filter.Limit < 1 || filter.Limit > 200 {
			return filter, false, appaktivitaet.ErrInvalidFilter
		}
	}

	if values.Get("offset") != "" {
		filter.Offset, err = strconv.Atoi(values.Get("offset"))
		if err != nil || filter.Offset < 0 {
			return filter, false, appaktivitaet.ErrInvalidFilter
		}
	}

	return filter, h.queryActive(values), nil
}

func (h *Handler) queryActive(values url.Values) bool {
	for _, entries := range values {
		if entries[0] != "" {
			return true
		}
	}

	return false
}
