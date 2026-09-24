package codexprofil

import (
	"context"
	"net/http"

	runtimecli "agentcontrolplane/app/internal/adapter/runtime/codexcli"
	appagent "agentcontrolplane/app/internal/app/agent"
	appcodexprofil "agentcontrolplane/app/internal/app/codexprofil"
	"agentcontrolplane/ui/bridge"
)

// Handler stellt das geschützte Codex-Arbeitsprofil bereit.
type Handler struct {
	profiles    *appcodexprofil.Service
	agents      *appagent.Service
	bridge      *bridge.Bridge
	bindAddress string
	runner      Runner
	mux         *http.ServeMux
}

// Runner prüft ein festes serverseitiges CLI-Profil ohne Clientparameter.
type Runner interface {
	Run(context.Context, string, string) (runtimecli.Result, error)
}

type saveRequest struct {
	WorkspaceEnabled *bool `json:"workspace_enabled"`
	WriteEnabled     *bool `json:"write_enabled"`
}

type errorResponse struct {
	Error       string            `json:"error,omitempty"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}
