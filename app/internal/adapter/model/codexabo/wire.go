package codexabo

import (
	"encoding/json"
	"errors"

	"github.com/cloudwego/eino/schema"
)

func (b *wireBuilder) messages(in []*schema.Message) ([]map[string]any, []string, error) {
	out := make([]map[string]any, 0, len(in))
	var instructions []string
	for _, m := range in {
		if m == nil {
			continue
		}

		if err := b.appendMessage(m, &out, &instructions); err != nil {
			return nil, nil, err
		}
	}

	return out, instructions, nil
}

func (b *wireBuilder) appendMessage(m *schema.Message, out *[]map[string]any, instructions *[]string) error {
	if b.appendTextMessage(m, out, instructions) {
		return nil
	}

	if b.appendActionMessage(m, out) {
		return nil
	}

	return errors.New("unsupported message role")
}

func (b *wireBuilder) appendTextMessage(m *schema.Message, out *[]map[string]any, instructions *[]string) bool {
	if m.Role == schema.System {
		*instructions = append(*instructions, m.Content)
		return true
	}

	if m.Role == schema.User {
		*out = append(*out, map[string]any{"role": "user", "content": m.Content})
		return true
	}

	return false
}

func (b *wireBuilder) appendActionMessage(m *schema.Message, out *[]map[string]any) bool {
	if m.Role == schema.Assistant {
		*out = b.appendAssistant(*out, m)
		return true
	}

	if m.Role == schema.Tool {
		*out = append(*out, map[string]any{"type": "function_call_output", "call_id": m.ToolCallID, "output": m.Content})
		return true
	}

	return false
}

func (b *wireBuilder) appendAssistant(out []map[string]any, m *schema.Message) []map[string]any {
	if m.Content != "" {
		out = append(out, map[string]any{"role": "assistant", "content": m.Content})
	}

	for _, c := range m.ToolCalls {
		out = append(out, map[string]any{"type": "function_call", "call_id": c.ID, "name": c.Function.Name, "arguments": c.Function.Arguments})
	}

	return out
}

func (b *wireBuilder) tools(tools []*schema.ToolInfo) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		if t == nil || t.Name == "" {
			return nil, errors.New("invalid tool definition")
		}

		params, err := b.parameters(t)
		if err != nil {
			return nil, err
		}

		out = append(out, map[string]any{"type": "function", "name": t.Name, "description": t.Desc, "parameters": params})
	}

	return out, nil
}

func (b *wireBuilder) parameters(t *schema.ToolInfo) (map[string]any, error) {
	params := map[string]any{"type": "object", "properties": map[string]any{}}
	if t.ParamsOneOf == nil {
		return params, nil
	}

	sc, err := t.ParamsOneOf.ToJSONSchema()
	if err != nil {
		return nil, errors.New("invalid tool parameters")
	}

	encoded, err := json.Marshal(sc)
	if err != nil {
		return nil, errors.New("invalid tool parameters")
	}

	if json.Unmarshal(encoded, &params) != nil {
		return nil, errors.New("invalid tool parameters")
	}

	return params, nil
}
