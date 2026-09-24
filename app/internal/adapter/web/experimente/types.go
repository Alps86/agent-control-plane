package experimente

import (
	"io"
	"net/http"
)

// Renderer ist die öffentliche, anbieterneutrale Grenze zur UI-Bridge.
type Renderer interface {
	Render(io.Writer, string, map[string]any) error
}

// Handler stellt das freigegebene Inventar ausschließlich lesend bereit.
type Handler struct {
	renderer Renderer
	loader   inventoryLoader
}

type inventoryLoader struct{}

type inventory struct {
	SourceAsOf   string  `json:"sourceAsOf"`
	SourceSHA256 string  `json:"sourceSha256"`
	Groups       []group `json:"groups"`
}

type group struct {
	Key     string  `json:"key"`
	Heading string  `json:"heading"`
	Entries []entry `json:"entries"`
}

type entry struct {
	SourceKey               string         `json:"sourceKey"`
	Name                    string         `json:"name"`
	SourceLinks             []sourceLink   `json:"sourceLinks"`
	PaperclipMaturity       string         `json:"paperclipMaturity"`
	Benefit                 string         `json:"benefit"`
	ACPDependencies         string         `json:"acpDependencies"`
	DecisionParts           []decisionPart `json:"decisionParts"`
	ACPImplementationStatus string         `json:"acpImplementationStatus"`
	RelatedSourceKey        string         `json:"relatedSourceKey,omitempty"`
}

type sourceLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type decisionPart struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

var _ http.Handler = (*Handler)(nil)
