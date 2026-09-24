package modellzugang

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

func (s *Suite) openCancelStream(req *http.Request) *http.Response {
	s.recordRequest(req)
	chunk := "event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"Teilantwort\"}\n\n"
	s.cancelBody = &cancelBody{data: []byte(chunk), done: make(chan struct{})}
	header := make(http.Header)
	header.Set("Content-Type", "text/event-stream")
	header.Set("x-request-id", "offline-cancel-1")
	return &http.Response{StatusCode: 200, Header: header, Body: s.cancelBody, Request: req}
}

func (s *Suite) runRawLocal() error {
	stream, err := s.rawAdapter.Stream(context.Background(), []*schema.Message{schema.UserMessage("Prüfaufruf")})
	if err != nil {
		s.modelErr = err
		return nil
	}

	defer stream.Close()
	s.consume(stream)
	return nil
}

func (s *Suite) runCancelLocal() error {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := s.rawAdapter.Stream(ctx, []*schema.Message{schema.UserMessage("Prüfaufruf")})
	if err != nil {
		cancel()
		return err
	}

	message, err := stream.Recv()
	if err == nil && message != nil {
		s.answer, s.chunks = message.Content, 1
	}

	cancel()
	stream.Close()
	s.cancelled = true
	return s.waitBodyClose()
}

func (s *Suite) waitBodyClose() error {
	select {
	case <-s.cancelBody.done:
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("lokaler SSE-Body blieb offen")
	}
}

func (s *Suite) offlineObservations() error {
	entries := s.observations.Entries()
	if len(entries) != 2 || entries[0].RequestID != "offline-request-1" ||
		entries[1].RequestID != "offline-request-2" {
		return fmt.Errorf("redigierte Request-ID-Reihenfolge fehlt")
	}

	return nil
}

func (s *Suite) offlineToolOutput() error {
	requests := s.requestBodies()
	if len(requests) != 2 || !strings.Contains(requests[1], "function_call_output") ||
		!strings.Contains(requests[1], "MS01-WEBSITE-4711") {
		return fmt.Errorf("zweiter Request enthält kein Fachtoolergebnis")
	}

	return nil
}
