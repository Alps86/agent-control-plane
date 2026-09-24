package modellzugang

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"agentcontrolplane/app/internal/adapter/model/codexabo"
	"agentcontrolplane/app/internal/app/modellpruefung"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func (s *Suite) prepareLive() error {
	s.live = true
	s.fixture = &statusFixture{status: modellpruefung.Status{Wert: "in Arbeit", Pruefkennung: "MS01-WEBSITE-4711"}}
	s.pruefung = modellpruefung.NewPruefung(s.fixture)
	s.trace = &requestTrace{base: http.DefaultTransport, expectedAccount: os.Getenv("MS01_LIVE_ACCOUNT_ID")}
	s.observations = &observationCollector{}
	s.tool = &fachTool{pruefung: s.pruefung, agent: "Mira", entered: make(chan struct{}, 1)}
	adapter, err := codexabo.New(os.Getenv("MS01_LIVE_MODEL"), &credential{
		token: os.Getenv("MS01_LIVE_TOKEN"), account: os.Getenv("MS01_LIVE_ACCOUNT_ID")},
		&http.Client{Transport: s.trace, Timeout: 90 * time.Second}, s.observations)
	if err != nil {
		return err
	}

	agent, err := react.NewAgent(context.Background(), &react.AgentConfig{ToolCallingModel: adapter,
		ToolsConfig: compose.ToolsNodeConfig{Tools: []tool.BaseTool{s.tool}}})
	s.agent = agent
	return err
}

func (s *Suite) runLive(prompt string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	input := append(append([]*schema.Message(nil), s.messages...), schema.UserMessage(prompt))
	stream, err := s.agent.Stream(ctx, input)
	if err != nil {
		s.modelErr = err
		return nil
	}

	defer stream.Close()
	s.answer, s.chunks = "", 0
	s.consume(stream)
	s.messages = append(input, schema.AssistantMessage(s.answer, nil))
	return nil
}

func (s *Suite) consume(stream *schema.StreamReader[*schema.Message]) {
	for {
		message, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			s.modelErr = err
			return
		}

		s.acceptMessage(message)
	}
}

func (s *Suite) acceptMessage(message *schema.Message) {
	if message == nil {
		return
	}

	if message.Content != "" {
		s.answer += message.Content
		s.chunks++
		s.notifyFirstChunk()
	}

	if message.ResponseMeta != nil && message.ResponseMeta.Usage == nil {
		s.unknownUsage = true
	}
}

func (s *Suite) notifyFirstChunk() {
	if s.firstChunk == nil {
		return
	}

	select {
	case s.firstChunk <- struct{}{}:
	default:
	}
}

func (t *requestTrace) RoundTrip(req *http.Request) (*http.Response, error) {
	observation := t.inspect(req)
	response, err := t.base.RoundTrip(req)
	if response != nil {
		observation.status = response.StatusCode
		observation.requestID = response.Header.Get("x-request-id")
		response.Body = &trackedBody{ReadCloser: response.Body, trace: t}
	}

	t.mu.Lock()
	t.requests = append(t.requests, observation)
	t.mu.Unlock()
	return response, err
}

func (t *requestTrace) inspect(req *http.Request) requestObservation {
	observation := requestObservation{host: req.URL.Hostname(), path: req.URL.Path}
	observation.accountMatched = t.expectedAccount != "" && req.Header.Get("ChatGPT-Account-ID") == t.expectedAccount
	observation.hasBearer = strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ")
	if req.Body == nil {
		return observation
	}

	body, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewReader(body))
	observation.hasToolOutput = bytes.Contains(body, []byte("function_call_output"))
	observation.hasMarker = bytes.Contains(body, []byte("MS01-WEBSITE-4711"))
	return observation
}

func (t *requestTrace) Requests() []requestObservation {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]requestObservation(nil), t.requests...)
}

func (b *trackedBody) Close() error {
	err := b.ReadCloser.Close()
	b.trace.mu.Lock()
	b.trace.bodyClosed = true
	b.trace.mu.Unlock()
	return err
}

func (t *requestTrace) BodyClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bodyClosed
}
