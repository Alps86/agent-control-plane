package modellpruefung

import (
	"encoding/json"
	"fmt"
	"net/http"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
)

func (p *probeRun) RecordModelResponse(observation codexabo.Observation) {
	complete := observation.Usage != nil && observation.Usage.InputTokens != nil && observation.Usage.OutputTokens != nil && observation.Usage.TotalTokens != nil
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observations = append(p.observations, probeObservation{
		requestID:          p.requestRef(observation.RequestID),
		status:             p.safeStatus(observation.Status),
		usageComplete:      complete,
		modelMatchesConfig: observation.Model != "" && observation.Model == p.modelID,
	})
}

func (p *probeRun) requestRef(value string) string {
	if value != "" {
		return fmt.Sprintf("redacted-request-%d", len(p.observations)+1)
	}

	return ""
}

func (p *probeRun) safeStatus(value string) string {
	if value == "completed" || value == "http_rejected" || value == "transport_error" || value == "canceled" || value == "stream_error" || value == "stream_truncated" || value == "invalid_stream" || value == "response.failed" || value == "response.incomplete" || value == "error" {
		return value
	}

	return "provider_error"
}

func (p *probeRun) snapshots() []probeObservation {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]probeObservation(nil), p.observations...)
}

func (h *Handler) finish(w http.ResponseWriter, flusher http.Flusher, run *probeRun, transport *boundedTransport, kind string, stream streamState) {
	called, state := run.toolSnapshot()
	h.reportTool(w, flusher, called, state)
	h.reportProvider(w, flusher, run)
	kind = h.completionKind(kind, called, transport.count())
	if kind != "" {
		h.writeEvent(w, flusher, "error", sseEvent{Kind: kind, Requests: transport.count(), AnswerMatchesExpected: h.boolRef(false), ChunkBeforeCompleted: h.boolRef(false)})
		return
	}

	h.writeEvent(w, flusher, "done", sseEvent{State: "completed", Tool: called, Requests: transport.count(), AnswerMatchesExpected: h.boolRef(h.matchesExpected(stream)), ChunkBeforeCompleted: h.boolRef(transport.streamedBeforeCompletion())})
}

func (h *Handler) boolRef(value bool) *bool { return &value }

func (h *Handler) reportTool(w http.ResponseWriter, flusher http.Flusher, called bool, state string) {
	if called {
		h.writeEvent(w, flusher, "tool", sseEvent{Tool: true, ToolState: state})
	}
}

func (h *Handler) reportProvider(w http.ResponseWriter, flusher http.Flusher, run *probeRun) {
	for _, observation := range run.snapshots() {
		h.writeEvent(w, flusher, "provider", sseEvent{
			RequestID:                  observation.requestID,
			Model:                      h.modelID,
			State:                      observation.status,
			UsageComplete:              observation.usageComplete,
			ProviderModelMatchesConfig: &observation.modelMatchesConfig,
		})
	}
}

func (h *Handler) completionKind(kind string, called bool, requests int) string {
	if kind == "" && !called {
		kind = "tool_not_called"
	}

	if kind == "" && requests != 2 {
		kind = "incomplete_round"
	}

	return kind
}

func (h *Handler) writeEvent(w http.ResponseWriter, flusher http.Flusher, name string, event sseEvent) bool {
	encoded, err := json.Marshal(event)
	if err != nil {
		return false
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, encoded)
	if err != nil {
		return false
	}

	flusher.Flush()
	return true
}
