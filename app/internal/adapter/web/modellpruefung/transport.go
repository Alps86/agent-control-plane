package modellpruefung

import (
	"errors"
	"net/http"
)

func (b *boundedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Scheme != "https" || r.URL.Host != "chatgpt.com" || r.URL.Path != "/backend-api/codex/responses" {
		return nil, errors.New("unexpected provider destination")
	}

	b.mu.Lock()
	b.requests++
	count := b.requests
	b.mu.Unlock()
	if count > 2 {
		return nil, errors.New("probe request limit reached")
	}

	return b.base.RoundTrip(r)
}

func (b *boundedTransport) count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.requests
}
