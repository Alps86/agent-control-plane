package openrouter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"agentcontrolplane/app/internal/app/openrouterverbindung"
)

func NewProbe(client HTTPDoer) *Client {
	return NewProbeAt(client, baseURL)
}

func NewProbeAt(client HTTPDoer, base string) *Client {
	parsed, err := url.Parse(base)
	if err == nil {
		parsed = &url.URL{Scheme: parsed.Scheme, Host: parsed.Host, Path: "/api/v1/key"}
	}

	probe := &Client{http: client}
	if parsed != nil {
		probe.endpoint = parsed.String()
	}

	if probe.http == nil {
		probe.http = &http.Client{Timeout: 5 * time.Second, CheckRedirect: probe.rejectRedirect}
	}

	return probe
}

func (probe *Client) rejectRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

func (probe *Client) Probe(ctx context.Context, key string) (Result, error) {
	if strings.TrimSpace(key) == "" {
		return Result{}, ErrMissingKey
	}

	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(bounded, http.MethodGet, probe.endpoint, nil)
	if err != nil {
		return Result{}, ErrUnavailable
	}

	request.Header.Set("Authorization", "Bearer "+key)
	response, err := probe.http.Do(request)
	return probe.read(ctx, response, err)
}

func (probe *Client) Check(ctx context.Context, key string) openrouterverbindung.CheckResult {
	result, err := probe.Probe(ctx, key)
	if err == nil && result.Valid {
		return openrouterverbindung.CheckReady
	}

	if errors.Is(err, ErrInvalidKey) || errors.Is(err, ErrMissingKey) {
		return openrouterverbindung.CheckInvalid
	}

	return openrouterverbindung.CheckUnavailable
}

func (probe *Client) read(ctx context.Context, response *http.Response, err error) (Result, error) {
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if ctx.Err() != nil {
		return Result{}, ctx.Err()
	}

	if err != nil || response == nil || response.Body == nil {
		return Result{}, ErrUnavailable
	}

	if response.StatusCode == http.StatusUnauthorized {
		return Result{}, ErrInvalidKey
	}

	if response.StatusCode != http.StatusOK {
		return Result{}, ErrUnavailable
	}

	return probe.parse(response.Body)
}

func (probe *Client) parse(body io.Reader) (Result, error) {
	content, err := io.ReadAll(io.LimitReader(body, maxResponseBytes+1))
	if err != nil || len(content) > maxResponseBytes {
		return Result{}, ErrInvalidResponse
	}

	var envelope keyEnvelope
	if json.Unmarshal(content, &envelope) != nil || envelope.Data == nil {
		return Result{}, ErrInvalidResponse
	}

	return Result{Valid: true}, nil
}
