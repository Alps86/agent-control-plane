package codexauth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (d *Direct) postJSON(ctx context.Context, path string, value any) (response, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return response{}, errProtocol
	}

	return d.post(ctx, path, "application/json", bytes.NewReader(data))
}

func (d *Direct) postForm(ctx context.Context, path string, form url.Values) (response, error) {
	return d.post(ctx, path, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
}

func (d *Direct) post(ctx context.Context, path, contentType string, body io.Reader) (response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, d.issuer+path, body)
	if err != nil {
		return response{}, errProtocol
	}

	request.Header.Set("Content-Type", contentType)
	reply, err := d.http.Do(request)
	if err != nil {
		return response{}, errProtocol
	}

	defer reply.Body.Close()
	return d.readResponse(reply)
}

func (d *Direct) readResponse(reply *http.Response) (response, error) {
	body, err := io.ReadAll(io.LimitReader(reply.Body, 65537))
	if err != nil || len(body) > 65536 {
		return response{}, errProtocol
	}

	return response{status: reply.StatusCode, body: body}, nil
}
