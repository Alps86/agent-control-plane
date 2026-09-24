package modellpruefung

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	appmodellpruefung "agentcontrolplane/app/internal/app/modellpruefung"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func (p *probeRun) tools() []tool.BaseTool { return []tool.BaseTool{p} }

func (p *probeRun) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: toolName, Desc: "Lies den technischen Verbindungsstatus der bereits eingerichteten Codex-Verbindung. Nutze diese Aktion genau einmal.", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{})}, nil
}

func (p *probeRun) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	if err := p.authorizeTool(ctx, arguments); err != nil {
		return "", err
	}

	result, err := p.status.Lies(ctx, appmodellpruefung.VerbindungsstatusAktion)
	if err != nil {
		return "", errors.New("connection status unavailable")
	}

	state := p.safeState(result.Zustand)
	p.recordTool(state)
	encoded, _ := json.Marshal(map[string]string{"zustand": state})
	return string(encoded), nil
}

func (p *probeRun) authorizeTool(ctx context.Context, arguments string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if !p.emptyArguments(arguments) {
		return errors.New("invalid tool arguments")
	}

	if !p.reserveTool() {
		return errors.New("tool already requested")
	}

	return nil
}

func (p *probeRun) reserveTool() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.toolRequested {
		return false
	}

	p.toolRequested = true
	return true
}

func (p *probeRun) emptyArguments(raw string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &fields) != nil {
		return false
	}

	return fields != nil && len(fields) == 0
}

func (p *probeRun) safeState(state string) string {
	if state == "connected" || state == "pending" || state == "idle" || state == "expired" || state == "denied" || state == "cancelled" {
		return state
	}

	return "unavailable"
}

func (p *probeRun) recordTool(state string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.toolCalled = true
	p.toolState = state
}

func (p *probeRun) toolSnapshot() (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.toolCalled, p.toolState
}

func (p *probeRun) checkToolCalls(ctx context.Context, reader *schema.StreamReader[*schema.Message]) (bool, error) {
	defer reader.Close()
	seen := false
	for {
		message, err := reader.Recv()
		if err == io.EOF {
			return seen, nil
		}

		if err != nil {
			return false, err
		}

		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if message != nil && len(message.ToolCalls) > 0 {
			seen = true
		}
	}
}
