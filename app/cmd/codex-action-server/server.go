package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"time"

	portprofil "agentcontrolplane/app/internal/port/codexprofil"
)

const actionSocket = "/run/acp/action.sock"

// Run bedient nur die fest benannte Go-Artefaktaktion über MCP stdio.
func (s *ActionServer) Run() error {
	decoder := json.NewDecoder(s.input)
	encoder := json.NewEncoder(s.output)
	for {
		var request rpcRequest
		err := decoder.Decode(&request)
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		if len(request.ID) == 0 {
			continue
		}

		if err := encoder.Encode(s.respond(request)); err != nil {
			return err
		}
	}
}

func (s *ActionServer) respond(request rpcRequest) rpcResponse {
	response := rpcResponse{JSONRPC: "2.0", ID: request.ID}
	if request.JSONRPC != "2.0" {
		response.Error = &rpcError{Code: -32600, Message: "invalid request"}
		return response
	}

	if request.Method == "initialize" {
		response.Result = map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "fach", "version": "1"}}
		return response
	}

	if request.Method == "tools/list" {
		response.Result = s.tools()
		return response
	}

	if request.Method == "tools/call" {
		response.Result = s.call(request.Params)
		return response
	}

	response.Error = &rpcError{Code: -32601, Message: "method not found"}
	return response
}

func (s *ActionServer) tools() any {
	return map[string]any{"tools": []any{map[string]any{
		"name": "artifact.markdown.save", "description": "Speichert ein fest benanntes privates Markdown-Artefakt.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{
			"markdown": map[string]string{"type": "string"}}, "required": []string{"markdown"}, "additionalProperties": false},
	}}}
}

func (s *ActionServer) call(raw json.RawMessage) any {
	call, err := s.decodeCall(raw)
	if err != nil {
		return s.toolResult(false, "action_denied")
	}

	if call.Name != "artifact.markdown.save" {
		return s.toolResult(false, "action_denied")
	}

	markdown, err := s.decodeMarkdown(call.Arguments)
	if err != nil {
		return s.toolResult(false, "action_denied")
	}

	result, err := s.forward(markdown)
	if err != nil {
		return s.toolResult(false, "action_denied")
	}

	if !result.Allowed {
		return s.toolResult(false, "action_denied")
	}

	return s.toolResult(true, result.Code)
}

func (s *ActionServer) decodeCall(raw json.RawMessage) (toolCall, error) {
	var call toolCall
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	return call, decoder.Decode(&call)
}

func (s *ActionServer) decodeMarkdown(raw json.RawMessage) (string, error) {
	var input markdownInput
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return "", err
	}

	if len(input.Markdown) == 0 || len(input.Markdown) > 65536 {
		return "", errors.New("invalid markdown")
	}

	return input.Markdown, nil
}

func (s *ActionServer) forward(markdown string) (portprofil.ActionResponse, error) {
	var result portprofil.ActionResponse
	if os.Getenv("ACP_CODEX_ACTION_SOCKET") != actionSocket {
		return result, errors.New("action socket unavailable")
	}

	connection, err := net.DialTimeout("unix", actionSocket, 5*time.Second)
	if err != nil {
		return result, err
	}

	defer connection.Close()
	connection.SetDeadline(time.Now().Add(10 * time.Second))
	request := portprofil.ActionRequest{ActionID: "artifact.markdown.save", Markdown: markdown}
	if err := json.NewEncoder(connection).Encode(request); err != nil {
		return result, err
	}

	err = json.NewDecoder(connection).Decode(&result)
	return result, err
}

func (s *ActionServer) toolResult(allowed bool, code string) any {
	return map[string]any{"content": []any{map[string]string{"type": "text", "text": code}}, "isError": !allowed}
}
