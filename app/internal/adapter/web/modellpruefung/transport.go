package modellpruefung

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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

	return b.dispatch(r, count)
}

func (b *boundedTransport) dispatch(r *http.Request, count int) (*http.Response, error) {
	resp, err := b.base.RoundTrip(r)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}

	resp.Body = &auditBody{inner: resp.Body, transport: b, index: count}
	return resp, nil
}

func (b *boundedTransport) count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.requests
}

func (b *boundedTransport) streamedBeforeCompletion() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.chunkBeforeCompleted
}

func (b *boundedTransport) markStreamed() {
	b.mu.Lock()
	b.chunkBeforeCompleted = true
	b.mu.Unlock()
}

func (b *boundedTransport) signalProgress() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.progressSent {
		return
	}

	b.progressSent = true
	select {
	case b.progress <- struct{}{}:
	default:
	}
}

func (a *auditBody) Read(buffer []byte) (int, error) {
	n, err := a.inner.Read(buffer)
	for _, value := range buffer[:n] {
		a.acceptByte(value)
	}

	return n, err
}

func (a *auditBody) Close() error { return a.inner.Close() }

func (a *auditBody) acceptByte(value byte) {
	if value == '\n' {
		a.acceptLine()
		return
	}

	if len(a.line) < 4096 {
		a.line = append(a.line, value)
		return
	}

	a.overflowed = true
}

func (a *auditBody) acceptLine() {
	line := strings.TrimSuffix(string(a.line), "\r")
	a.line = a.line[:0]
	if a.overflowed {
		a.eventType = ""
		a.eventValid = false
		a.completedValid = false
		a.overflowed = false
		return
	}

	if line == "" {
		a.flush()
		return
	}

	a.acceptField(line)
}

func (a *auditBody) acceptField(line string) {
	if strings.HasPrefix(line, "event:") {
		a.eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
	}

	if strings.HasPrefix(line, "data:") {
		a.acceptData(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
	}
}

func (a *auditBody) acceptData(data string) {
	a.dataLines++
	if a.dataLines > 1 {
		a.eventValid = false
		return
	}

	a.parseData(data)
}

func (a *auditBody) parseData(data string) {
	var event auditEvent
	if json.Unmarshal([]byte(data), &event) != nil {
		return
	}

	a.eventValid = true
	a.completedValid = event.Response.Status == "completed"
	if event.Type != "" {
		a.eventType = event.Type
	}

	if a.eventType == "response.output_text.delta" && event.Delta != "" {
		a.deltaPending = true
	}
}

func (a *auditBody) flush() {
	if a.eventValid && a.eventType == "response.output_text.delta" && a.deltaPending {
		a.deltaSeen = true
		if a.index == 2 {
			a.transport.signalProgress()
		}
	}

	if a.index == 2 && a.eventValid && a.completedValid && a.eventType == "response.completed" && a.deltaSeen {
		a.transport.markStreamed()
	}

	a.eventType = ""
	a.eventValid = false
	a.completedValid = false
	a.deltaPending = false
	a.dataLines = 0
}
