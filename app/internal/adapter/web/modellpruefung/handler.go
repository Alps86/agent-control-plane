package modellpruefung

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"agentcontrolplane/app/internal/port/credentials"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// NewHandler constructs the local probe with server-owned connection and model selection.
func NewHandler(access credentials.AccessResolver, status StatusAction, modelID, bindAddress string, client *http.Client) (*Handler, error) {
	h := &Handler{access: access, status: status, modelID: modelID, bindAddress: bindAddress, client: client, slot: make(chan struct{}, 1)}
	if !h.validBind() {
		return nil, errors.New("probe requires a loopback listener")
	}

	return h, nil
}

func (h *Handler) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/settings/modelle/codex/probe", h.probe)
	return mux
}

func (h *Handler) probe(w http.ResponseWriter, r *http.Request) {
	if !h.preflight(w, r) {
		return
	}

	if !h.acquire() {
		h.jsonError(w, http.StatusConflict, "probe_busy")
		return
	}

	defer h.release()
	h.run(w, r)
}

func (h *Handler) preflight(w http.ResponseWriter, r *http.Request) bool {
	if !h.allowed(r) {
		h.jsonError(w, http.StatusForbidden, "request_not_allowed")
		return false
	}

	if !h.emptyBody(r) {
		h.jsonError(w, http.StatusBadRequest, "empty_object_required")
		return false
	}

	if h.modelID == "" || h.access == nil || h.status == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "model_not_ready")
		return false
	}

	return true
}

func (h *Handler) acquire() bool {
	select {
	case h.slot <- struct{}{}:
		return true
	default:
		return false
	}
}
func (h *Handler) release() { <-h.slot }

func (h *Handler) run(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()
	run := &probeRun{status: h.status}
	transport := h.transport()
	chat, err := codexabo.New(h.modelID, h.access, &http.Client{Transport: transport, Timeout: 45 * time.Second}, run)
	if err != nil {
		h.jsonError(w, http.StatusServiceUnavailable, "model_not_ready")
		return
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{ToolCallingModel: chat, ToolsConfig: compose.ToolsNodeConfig{Tools: run.tools()}, StreamToolCallChecker: run.checkToolCalls})
	if err != nil {
		h.jsonError(w, http.StatusServiceUnavailable, "model_not_ready")
		return
	}

	h.stream(w, ctx, cancel, agent, run, transport)
}

func (h *Handler) transport() *boundedTransport {
	base := http.RoundTripper(&http.Transport{Proxy: nil})
	if h.client != nil && h.client.Transport != nil {
		base = h.client.Transport
	}

	return &boundedTransport{base: base}
}

func (h *Handler) jsonError(w http.ResponseWriter, status int, kind string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(sseEvent{State: "unavailable", Kind: kind})
}
