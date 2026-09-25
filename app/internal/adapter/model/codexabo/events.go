package codexabo

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func (p *eventParser) send(m *schema.Message, err error) bool {
	select {
	case <-p.ctx.Done():
		return false
	case p.out <- streamEvent{m, err}:
		return true
	}
}

func (p *eventParser) flush() {
	if len(p.data) == 0 || p.failed || p.completed {
		return
	}

	var e wireEvent
	if json.Unmarshal([]byte(strings.Join(p.data, "\n")), &e) != nil {
		p.failed = true
		p.observe(Observation{RequestID: p.requestID, Model: p.headerModel, Status: "invalid_stream"})
		p.send(nil, errors.New("invalid provider stream event"))
		return
	}

	typ := e.Type
	if typ == "" {
		typ = p.event
	}

	p.dispatch(typ, e)
}

func (p *eventParser) dispatch(typ string, e wireEvent) {
	if typ == "response.output_text.delta" {
		p.textDelta(e.Delta)
		return
	}

	if typ == "response.output_item.done" {
		p.toolItem(e.Item)
		return
	}

	if typ == "response.completed" {
		p.complete(e.Response)
		return
	}

	if typ == "response.failed" || typ == "response.incomplete" || typ == "error" {
		p.failure(e, typ)
	}
}

func (p *eventParser) textDelta(delta string) {
	if delta == "" {
		return
	}

	p.delivered = true
	p.send(&schema.Message{Role: schema.Assistant, Content: delta}, nil)
}

func (p *eventParser) toolItem(item wireItem) {
	if item.Type != "function_call" {
		return
	}

	p.pending = append(p.pending, schema.ToolCall{ID: item.CallID, Type: "function", Function: schema.FunctionCall{Name: item.Name, Arguments: item.Arguments}})
}

func (p *eventParser) failure(e wireEvent, status string) {
	p.failed = true
	code := e.Error.Code
	if code == "" {
		code = e.Response.Error.Code
	}

	reason := e.Response.IncompleteDetails.Reason
	if reason == "" {
		reason = e.Error.Type
	}

	k, r := p.providerReason(code, reason)
	p.observe(Observation{RequestID: p.requestID, Model: p.headerModel, Status: status})
	p.send(nil, &ProviderError{Kind: k, Reason: r, RequestID: p.requestID})
}

func (p *eventParser) complete(r wireResponse) {
	if r.Status != "" && r.Status != "completed" {
		p.failure(wireEvent{Response: r}, r.Status)
		return
	}

	p.completed = true
	p.completedText(r.Output)
	p.completedTools(r.Output)
	p.completionMetadata(r)
}

func (p *eventParser) completionMetadata(r wireResponse) {
	actual := p.safeID(r.Model)
	if actual == "" {
		actual = p.headerModel
	}

	usage := p.usage(r.Usage)
	extra := map[string]any{"transport": "codex_subscription_responses"}
	if p.requestID != "" {
		extra["provider_request_id"] = p.requestID
	}

	if actual != "" {
		extra["provider_model"] = actual
	}

	p.observe(Observation{RequestID: p.requestID, Model: actual, Status: "completed", Usage: usage})
	p.send(&schema.Message{Role: schema.Assistant, ResponseMeta: &schema.ResponseMeta{FinishReason: "stop", Usage: p.einoUsage(usage)}, Extra: extra}, nil)
}

func (p *eventParser) completedText(items []wireItem) {
	if p.delivered {
		return
	}

	for _, item := range items {
		if item.Type != "message" {
			continue
		}

		for _, part := range item.Content {
			if part.Type == "output_text" && part.Text != "" {
				p.send(&schema.Message{Role: schema.Assistant, Content: part.Text}, nil)
			}
		}
	}
}

func (p *eventParser) completedTools(items []wireItem) {
	if len(p.pending) == 0 {
		for _, item := range items {
			p.toolItem(item)
		}
	}

	if len(p.pending) > 0 {
		p.send(&schema.Message{Role: schema.Assistant, ToolCalls: p.pending}, nil)
	}
}
