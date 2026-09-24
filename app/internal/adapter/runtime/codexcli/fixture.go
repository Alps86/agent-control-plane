package codexcli

import (
	"encoding/json"
	"io"
	"net/http"
)

func (f *fixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/v1/responses" {
		http.Error(w, "synthetic endpoint only", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "invalid synthetic request", http.StatusBadRequest)
		return
	}

	f.respond(w, body)
}

func (f *fixture) respond(w http.ResponseWriter, body []byte) {
	var request map[string]any
	if json.Unmarshal(body, &request) != nil {
		http.Error(w, "invalid synthetic request", http.StatusBadRequest)
		return
	}

	f.captureTools(request)
	item := f.nextItem(request)
	f.writeEvents(w, item)
}

func (f *fixture) captureTools(request map[string]any) {
	tools, _ := request["tools"].([]any)
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, raw := range tools {
		tool, _ := raw.(map[string]any)
		name, _ := tool["name"].(string)
		f.tools = append(f.tools, name)
		f.captureNested(tool)
	}
}

func (f *fixture) captureNested(namespace map[string]any) {
	tools, _ := namespace["tools"].([]any)
	for _, raw := range tools {
		tool, _ := raw.(map[string]any)
		name, _ := tool["name"].(string)
		f.tools = append(f.tools, name)
	}
}

func (f *fixture) names() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.tools...)
}

func (f *fixture) nextItem(request map[string]any) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requestCount++
	if f.requestCount > 1 || f.hasToolOutput(request) {
		return f.answer()
	}

	return f.call()
}

func (f *fixture) hasToolOutput(request map[string]any) bool {
	inputs, _ := request["input"].([]any)
	for _, raw := range inputs {
		item, _ := raw.(map[string]any)
		if item["type"] == "function_call_output" {
			return true
		}
	}

	return false
}

func (f *fixture) answer() map[string]any {
	return map[string]any{"id": "msg_synthetic", "type": "message", "role": "assistant", "status": "completed",
		"content": []any{map[string]any{"type": "output_text", "text": "SYNTHETIC_OK", "annotations": []any{}}}}
}

func (f *fixture) call() map[string]any {
	name, namespace, args := f.callSpec()
	encoded, _ := json.Marshal(args)
	item := map[string]any{"id": "fc_synthetic", "type": "function_call", "status": "completed",
		"call_id": "call_synthetic", "name": name, "arguments": string(encoded)}
	if namespace != "" {
		item["namespace"] = namespace
	}

	return item
}

func (f *fixture) callSpec() (string, string, map[string]any) {
	args := map[string]any{"markdown": "SYNTHETIC_ARTIFACT"}
	if f.scenario == "shell_denied" {
		return "exec_command", "", map[string]any{"cmd": "touch /work/forged-shell"}
	}

	if f.scenario == "http_denied" {
		return "http_request", "", map[string]any{"url": "https://example.invalid/"}
	}

	f.attackArguments(args)
	return "artifact_markdown_save", "mcp__fach", args
}

func (f *fixture) attackArguments(args map[string]any) {
	if f.scenario == "foreign_org" {
		args["organization_id"] = "foreign-organization"
	}

	if f.scenario == "foreign_agent" {
		args["agent_id"] = "foreign-agent"
	}

	if f.scenario == "traversal" {
		args["path"] = "../outside"
	}

	if f.scenario == "symlink" {
		args["path"] = "/work/outside-link"
	}
}

func (f *fixture) writeEvents(w http.ResponseWriter, item map[string]any) {
	response := map[string]any{"id": "resp_synthetic", "object": "response", "created_at": 0,
		"status": "completed", "model": "synthetic-model", "output": []any{item},
		"usage": map[string]any{"input_tokens": 1, "output_tokens": 1, "total_tokens": 2}}
	w.Header().Set("Content-Type", "text/event-stream")
	f.event(w, "response.created", map[string]any{"response": f.started(response)})
	f.event(w, "response.output_item.added", map[string]any{"output_index": 0, "item": f.started(item)})
	f.event(w, "response.output_item.done", map[string]any{"output_index": 0, "item": item})
	f.event(w, "response.completed", map[string]any{"response": response})
}

func (f *fixture) started(item map[string]any) map[string]any {
	copy := make(map[string]any, len(item))
	for key, value := range item {
		copy[key] = value
	}

	copy["status"] = "in_progress"
	if copy["type"] == "function_call" {
		copy["arguments"] = ""
	}

	if copy["type"] == "message" {
		copy["content"] = []any{}
	}

	if copy["object"] == "response" {
		copy["output"] = []any{}
	}

	return copy
}

func (f *fixture) event(w http.ResponseWriter, name string, data map[string]any) {
	data["type"] = name
	encoded, _ := json.Marshal(data)
	_, _ = w.Write([]byte("event: " + name + "\ndata: "))
	_, _ = w.Write(encoded)
	_, _ = w.Write([]byte("\n\n"))
}
