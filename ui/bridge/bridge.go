package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"agentcontrolplane/ui/internal/render"
)

// New opens the built, embedded UI contract.
func New() (*Bridge, error) {
	source, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded UI: %w", err)
	}

	renderer, err := render.New(source)
	if err != nil {
		return nil, fmt.Errorf("parse embedded UI: %w", err)
	}

	assets, err := fs.Sub(source, "assets")
	if err != nil {
		return nil, fmt.Errorf("open UI assets: %w", err)
	}

	fragments, err := fs.Sub(source, "fragments")
	if err != nil {
		return nil, fmt.Errorf("open UI fragments: %w", err)
	}

	return &Bridge{source, renderer, http.FileServer(http.FS(assets)), http.FileServer(http.FS(fragments))}, nil
}

// Render renders a complete page or HTMX content from the same display map.
func (b *Bridge) Render(w io.Writer, templateName string, data map[string]any) error {
	return b.renderer.Render(w, templateName, data)
}

// Assets serves only public assets and fragments from the embedded build.
func (b *Bridge) Assets() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", b.assets))
	mux.Handle("/fragments/", http.StripPrefix("/fragments/", b.fragments))
	return mux
}

// LoadFixture reads synthetic preview data by its simple fixture name.
func (b *Bridge) LoadFixture(name string) (map[string]any, error) {
	base := strings.TrimSuffix(name, ".json")
	valid, err := regexp.MatchString(`^[a-z][a-z0-9-]*$`, base)
	if err != nil || !valid {
		return nil, errors.New("invalid fixture name")
	}

	content, err := fs.ReadFile(b.source, "fixtures/"+base+".json")
	if err != nil {
		return nil, fmt.Errorf("read fixture %q: %w", base, err)
	}

	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("decode fixture %q: %w", base, err)
	}

	return data, nil
}
