package modellpruefung

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func (h *Handler) stream(w http.ResponseWriter, ctx context.Context, cancel context.CancelFunc, agent *react.Agent, run *probeRun, transport *boundedTransport) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.jsonError(w, http.StatusServiceUnavailable, "stream_unavailable")
		return
	}

	h.streamHeaders(w)
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(45 * time.Second))
	h.writeEvent(w, flusher, "start", sseEvent{State: "running"})
	reader, err := agent.Stream(ctx, h.prompt())
	if err != nil {
		h.finish(w, flusher, run, transport, h.errorKind(ctx, err))
		return
	}

	defer reader.Close()
	state, err := h.consume(w, flusher, ctx, cancel, reader)
	h.finish(w, flusher, run, transport, h.resultKind(ctx, err, state))
}

func (h *Handler) streamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

func (h *Handler) prompt() []*schema.Message {
	return []*schema.Message{{Role: schema.System, Content: "Call the registered verbindungsstatus_lesen function exactly once with an empty object. Then answer briefly. Do not request any other action."}, {Role: schema.User, Content: "Prüfe den technischen Verbindungsstatus."}}
}

func (h *Handler) consume(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, cancel context.CancelFunc, reader *schema.StreamReader[*schema.Message]) (streamState, error) {
	state := streamState{}
	for {
		message, err := reader.Recv()
		if err == io.EOF {
			return state, nil
		}

		if err != nil {
			return state, err
		}

		if ctx.Err() != nil {
			return state, ctx.Err()
		}

		if err := h.consumeChunk(w, flusher, cancel, &state, message); err != nil {
			return state, err
		}
	}
}

func (h *Handler) consumeChunk(w http.ResponseWriter, flusher http.Flusher, cancel context.CancelFunc, state *streamState, message *schema.Message) error {
	if message == nil || message.Content == "" {
		return nil
	}

	state.chunks++
	state.bytes += len(message.Content)
	if state.chunks > 64 || state.bytes > 4096 {
		cancel()
		return errors.New("output limit")
	}

	if !h.writeEvent(w, flusher, "chunk", sseEvent{Sequence: state.chunks}) {
		cancel()
		return context.Canceled
	}

	return nil
}

func (h *Handler) resultKind(ctx context.Context, err error, state streamState) string {
	if err == nil {
		return ""
	}

	if state.chunks > 64 || state.bytes > 4096 {
		return "output_limit"
	}

	return h.errorKind(ctx, err)
}

func (h *Handler) errorKind(ctx context.Context, err error) string {
	if ctx.Err() != nil {
		return "cancelled"
	}

	var provider *codexabo.ProviderError
	if errors.As(err, &provider) {
		return provider.Kind
	}

	return "execution_error"
}
