package codexabo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func (a *Adapter) makeRequest(ctx context.Context, in []*schema.Message, opts ...model.Option) (*http.Request, error) {
	if a.Auth == nil || a.Model == "" {
		return nil, errors.New("subscription model or credential port absent")
	}

	cfg := model.GetCommonOptions(nil, opts...)
	body, err := a.requestBody(in, cfg)
	if err != nil {
		return nil, err
	}

	return a.authorizedRequest(ctx, body)
}

func (a *Adapter) authorizedRequest(ctx context.Context, body []byte) (*http.Request, error) {
	token, account, err := a.credential(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	a.setHeaders(req, token, account)
	return req, nil
}

func (a *Adapter) credential(ctx context.Context) (string, string, error) {
	token, account, err := a.Auth.Access(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return "", "", ctx.Err()
		}

		return "", "", errors.New("subscription credential unavailable")
	}

	if token == "" {
		return "", "", errors.New("empty subscription token")
	}

	return token, account, nil
}

func (a *Adapter) setHeaders(req *http.Request, token, account string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("originator", "agent-control-plane")
	req.Header.Set("User-Agent", "agent-control-plane/story-08")
	if account != "" {
		req.Header.Set("ChatGPT-Account-ID", account)
	}
}

func (a *Adapter) requestBody(in []*schema.Message, cfg *model.Options) ([]byte, error) {
	name, tools := a.selection(cfg)
	builder := &wireBuilder{}
	input, instructions, err := builder.messages(in)
	if err != nil {
		return nil, err
	}

	defs, err := builder.tools(tools)
	if err != nil {
		return nil, err
	}

	p := map[string]any{"model": name, "instructions": strings.Join(instructions, "\n"), "input": input, "tools": defs, "tool_choice": "auto", "parallel_tool_calls": false, "store": false, "stream": true, "include": []string{}}
	return json.Marshal(p)
}

func (a *Adapter) selection(cfg *model.Options) (string, []*schema.ToolInfo) {
	name := a.Model
	if cfg.Model != nil && *cfg.Model != "" {
		name = *cfg.Model
	}

	tools := a.tools
	if cfg.Tools != nil {
		tools = cfg.Tools
	}

	return name, tools
}

func (a *Adapter) send(req *http.Request) (*http.Response, error) {
	guarded := a.guardedClient()
	resp, err := guarded.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return nil, req.Context().Err()
		}

		if a.Observer != nil {
			a.Observer.RecordModelResponse(Observation{Status: "transport_error"})
		}
		return nil, errors.New("subscription transport failed")
	}

	if resp.StatusCode == http.StatusOK {
		return resp, nil
	}

	return nil, a.reject(resp)
}

func (a *Adapter) guardedClient() *http.Client {
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}

	guarded := *client
	guarded.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &guarded
}

func (a *Adapter) reject(resp *http.Response) error {
	defer resp.Body.Close()
	id := a.safeID(resp.Header.Get("x-request-id"))
	if a.Observer != nil {
		a.Observer.RecordModelResponse(Observation{RequestID: id, Status: "http_rejected"})
	}

	return &ProviderError{Kind: a.statusKind(resp.StatusCode), Status: resp.StatusCode, RequestID: id}
}
