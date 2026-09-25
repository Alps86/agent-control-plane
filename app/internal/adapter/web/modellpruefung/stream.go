package modellpruefung

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
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
	reader, err := h.awaitAgent(w, flusher, ctx, cancel, agent, transport)
	if err != nil {
		h.finish(w, flusher, run, transport, h.errorKind(ctx, err), streamState{})
		return
	}

	defer reader.Close()
	state, err := h.consume(w, flusher, ctx, cancel, reader, run, transport)
	h.finish(w, flusher, run, transport, h.resultKind(ctx, err, state), state)
}

func (h *Handler) awaitAgent(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, cancel context.CancelFunc, agent *react.Agent, transport *boundedTransport) (*schema.StreamReader[*schema.Message], error) {
	starts := h.startAgent(ctx, agent)
	for {
		if err := h.reportIfReady(w, flusher, cancel, transport.progress); err != nil {
			return nil, err
		}

		select {
		case result := <-starts:
			return result.reader, result.err
		case <-transport.progress:
			if _, err := h.reportProgress(w, flusher, cancel); err != nil {
				return nil, err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (h *Handler) reportIfReady(w http.ResponseWriter, flusher http.Flusher, cancel context.CancelFunc, progress <-chan struct{}) error {
	if h.progressReady(progress) {
		_, err := h.reportProgress(w, flusher, cancel)
		return err
	}

	return nil
}

func (h *Handler) startAgent(ctx context.Context, agent *react.Agent) <-chan streamStart {
	starts := make(chan streamStart)
	go func() {
		reader, err := agent.Stream(ctx, h.prompt())
		select {
		case starts <- streamStart{reader: reader, err: err}:
		case <-ctx.Done():
			if reader != nil {
				reader.Close()
			}
		}
	}()
	return starts
}

func (h *Handler) streamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

func (h *Handler) prompt() []*schema.Message {
	return []*schema.Message{{Role: schema.System, Content: "Call the registered verbindungsstatus_lesen function exactly once with an empty object. Then include the public marker MS01-PROBE-OK in the final answer. Do not request any other action."}, {Role: schema.User, Content: "Prüfe den technischen Verbindungsstatus."}}
}

func (h *Handler) consume(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, cancel context.CancelFunc, reader *schema.StreamReader[*schema.Message], run *probeRun, transport *boundedTransport) (streamState, error) {
	state := streamState{}
	messages := h.readMessages(ctx, reader)
	for {
		delivery := h.nextDelivery(ctx, messages, transport.progress)
		done, err := h.processDelivery(w, flusher, ctx, cancel, &state, delivery, run)
		if done {
			return state, err
		}
	}
}

func (h *Handler) readMessages(ctx context.Context, reader *schema.StreamReader[*schema.Message]) <-chan streamDelivery {
	results := make(chan streamDelivery, 1)
	go func() {
		defer close(results)
		for {
			message, err := reader.Recv()
			select {
			case results <- streamDelivery{message: message, err: err}:
			case <-ctx.Done():
				return
			}

			if err != nil {
				return
			}
		}
	}()
	return results
}

func (h *Handler) nextDelivery(ctx context.Context, messages <-chan streamDelivery, progress <-chan struct{}) streamDelivery {
	if ctx.Err() != nil {
		return streamDelivery{err: ctx.Err()}
	}

	if h.progressReady(progress) {
		return streamDelivery{progress: true}
	}

	return h.waitDelivery(ctx, messages, progress)
}

func (h *Handler) progressReady(progress <-chan struct{}) bool {
	select {
	case <-progress:
		return true
	default:
		return false
	}
}

func (h *Handler) waitDelivery(ctx context.Context, messages <-chan streamDelivery, progress <-chan struct{}) streamDelivery {
	select {
	case <-progress:
		return streamDelivery{progress: true}
	case result, ok := <-messages:
		if !ok {
			return streamDelivery{err: io.EOF}
		}

		return result
	case <-ctx.Done():
		return streamDelivery{err: ctx.Err()}
	}
}

func (h *Handler) processDelivery(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, cancel context.CancelFunc, state *streamState, delivery streamDelivery, run *probeRun) (bool, error) {
	if delivery.progress {
		return h.reportProgress(w, flusher, cancel)
	}

	if delivery.err == io.EOF {
		return true, nil
	}

	if delivery.err != nil {
		return true, delivery.err
	}

	err := h.nextChunk(w, flusher, ctx, cancel, state, delivery.message, run)
	return err != nil, err
}

func (h *Handler) reportProgress(w http.ResponseWriter, flusher http.Flusher, cancel context.CancelFunc) (bool, error) {
	if h.writeEvent(w, flusher, "provider_progress", sseEvent{State: "provider_delta_received", Requests: 2}) {
		return false, nil
	}

	cancel()
	return true, context.Canceled
}

func (h *Handler) nextChunk(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, cancel context.CancelFunc, state *streamState, message *schema.Message, run *probeRun) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return h.consumeChunk(w, flusher, cancel, state, message, run)
}

func (h *Handler) consumeChunk(w http.ResponseWriter, flusher http.Flusher, cancel context.CancelFunc, state *streamState, message *schema.Message, run *probeRun) error {
	if message == nil || message.Content == "" {
		return nil
	}

	state.chunks++
	state.bytes += len(message.Content)
	if state.chunks > 64 || state.bytes > 4096 {
		cancel()
		return errors.New("output limit")
	}
	state.captureFinal(message, run)

	if !h.writeEvent(w, flusher, "chunk", sseEvent{Sequence: state.chunks}) {
		cancel()
		return context.Canceled
	}

	return nil
}

func (s *streamState) captureFinal(message *schema.Message, run *probeRun) {
	called, _ := run.toolSnapshot()
	if called && message.Role == schema.Assistant {
		s.finalText += message.Content
	}
}

func (h *Handler) matchesExpected(state streamState) bool {
	return strings.Contains(state.finalText, expectedAnswer)
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
