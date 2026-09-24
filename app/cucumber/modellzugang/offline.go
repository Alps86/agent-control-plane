package modellzugang

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"agentcontrolplane/app/internal/app/modellpruefung"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// NewSuite erstellt den isolierten öffentlichen Godog-Prüfpfad.
func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t, nachweis: modellpruefung.NewNachweis()}
}

func (s *Suite) prepareLocal(mode string) error {
	s.mode = mode
	s.requests = nil
	s.fixture = &statusFixture{status: modellpruefung.Status{Wert: "in Arbeit", Pruefkennung: "MS01-WEBSITE-4711"}}
	s.pruefung = modellpruefung.NewPruefung(s.fixture)
	s.observations = &observationCollector{}
	s.local = true
	transport := &localTransport{suite: s}
	adapter, err := codexabo.New("codex-test-model", &credential{token: "offline-token"}, &http.Client{Transport: transport}, s.observations)
	if err != nil {
		return err
	}

	s.rawAdapter = adapter
	return s.buildLocalAgent(adapter)
}

func (s *Suite) buildLocalAgent(adapter model.ToolCallingChatModel) error {
	toolNode := &fachTool{pruefung: s.pruefung, agent: "Mira"}
	agent, err := react.NewAgent(context.Background(), &react.AgentConfig{ToolCallingModel: adapter,
		ToolsConfig: compose.ToolsNodeConfig{Tools: []tool.BaseTool{toolNode}}})
	s.agent = agent
	return err
}

func (s *Suite) serve(w http.ResponseWriter, r *http.Request) {
	count := s.recordRequest(r)
	if s.mode == "limit" {
		w.Header().Set("x-request-id", "offline-limit-1")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	s.serveStream(w, count)
}

func (s *Suite) recordRequest(r *http.Request) int {
	body, _ := io.ReadAll(r.Body)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, string(body))
	return len(s.requests)
}

func (s *Suite) serveStream(w http.ResponseWriter, count int) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("x-request-id", fmt.Sprintf("offline-request-%d", count))
	if s.mode == "sse_limit" {
		s.writeSSELimit(w)
		return
	}

	if s.mode == "usage" {
		s.writeAnswer(w)
		return
	}

	if count == 1 {
		s.writeToolCall(w)
		return
	}

	s.writeAnswer(w)
}

func (s *Suite) writeSSELimit(w http.ResponseWriter) {
	fmt.Fprint(w, "event: response.failed\n")
	fmt.Fprint(w, "data: {\"type\":\"response.failed\",\"error\":{\"code\":\"rate_limit_exceeded\",\"type\":\"usage_limit\"}}\n\n")
}

func (s *Suite) writeToolCall(w http.ResponseWriter) {
	fmt.Fprint(w, "event: response.output_item.done\n")
	fmt.Fprint(w, "data: {\"type\":\"response.output_item.done\",\"item\":{\"type\":\"function_call\",\"call_id\":\"status-1\",\"name\":\"projektstatus_lesen\",\"arguments\":\"{\\\"projekt\\\":\\\"Website\\\"}\"}}\n\n")
	fmt.Fprint(w, "event: response.completed\n")
	fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"model\":\"codex-test-model\"}}\n\n")
}

func (s *Suite) writeAnswer(w http.ResponseWriter) {
	fmt.Fprint(w, "event: response.output_text.delta\n")
	fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Website: in Arbeit, MS01-WEBSITE-4711\"}\n\n")
	fmt.Fprint(w, "event: response.completed\n")
	fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"model\":\"codex-test-model\"}}\n\n")
}

func (s *Suite) runLocal() error {
	stream, err := s.agent.Stream(context.Background(), []*schema.Message{schema.UserMessage("Status von Website?")})
	if err != nil {
		s.modelErr = err
		return nil
	}

	defer stream.Close()
	s.consume(stream)
	return nil
}

func (s *Suite) requestBodies() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.requests...)
}

func (s *Suite) hasToolResult() bool {
	for _, request := range s.requestBodies() {
		if strings.Contains(request, "function_call_output") && strings.Contains(request, "MS01-WEBSITE-4711") {
			return true
		}
	}

	return false
}
