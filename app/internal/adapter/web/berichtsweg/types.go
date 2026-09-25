package berichtsweg

import (
	"net/http"

	appberichtsweg "agentcontrolplane/app/internal/app/berichtsweg"
	"agentcontrolplane/ui/bridge"
)

// Handler exposes the organization chart to local operators.
type Handler struct {
	service     *appberichtsweg.Service
	bridge      *bridge.Bridge
	bindAddress string
	mux         *http.ServeMux
}

type assignRequest struct {
	ParentID string `json:"parent_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type chartNode struct {
	ID            string
	Name          string
	ExecutionKind string
	ParentID      string
	ParentName    string
	ProfileURL    string
	Children      []*chartNode
}
