package preview

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"agentcontrolplane/ui/bridge"
)

// New prepares the embedded preview without opening a network listener.
func New() (*Preview, error) {
	ui, err := bridge.New()
	if err != nil {
		return nil, err
	}

	return &Preview{bridge: ui}, nil
}

// Serve binds the preview exclusively to the IPv4 loopback interface.
func (p *Preview) Serve() error {
	port, err := p.port()
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", port))
	if err != nil {
		return fmt.Errorf("listen on preview loopback: %w", err)
	}

	fmt.Fprintf(os.Stderr, "UI preview listening on http://%s\n", listener.Addr())
	return http.Serve(listener, p)
}

func (p *Preview) port() (string, error) {
	value := os.Getenv("ACP_PREVIEW_PORT")
	if value == "" {
		return "4174", nil
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid ACP_PREVIEW_PORT %q: expected 1–65535", value)
	}

	return strconv.Itoa(port), nil
}

// ServeHTTP serves only pages, built assets, and the embedded help fragment.
func (p *Preview) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path == "/" {
		p.servePage(w, r)
		return
	}

	if p.staticAllowed(r.URL.Path) {
		p.bridge.Assets().ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func (p *Preview) staticAllowed(path string) bool {
	return path == "/assets/index.css" || path == "/assets/htmx.min.js" ||
		path == "/assets/preview.js" || path == "/fragments/preview-help.html"
}

func (p *Preview) servePage(w http.ResponseWriter, r *http.Request) {
	name, variant, ok := p.selection(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	data, err := p.bridge.LoadFixture(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	p.navigation(data, variant)
	p.render(w, r, data)
}

func (p *Preview) selection(r *http.Request) (string, string, bool) {
	view := r.URL.Query().Get("view")
	if view == "" {
		view = "organization"
	}

	variant := r.URL.Query().Get("fixture")
	if !p.validView(view) || !p.validVariant(variant) {
		return "", "", false
	}

	if variant == "" || variant == "default" {
		return view, "", true
	}

	return view + "-" + variant, variant, true
}

func (p *Preview) validView(view string) bool {
	return view == "organization" || view == "projects" ||
		view == "agents" || view == "tasks"
}

func (p *Preview) validVariant(variant string) bool {
	return variant == "" || variant == "default" || variant == "alternate" ||
		variant == "empty" || variant == "error"
}

func (p *Preview) navigation(data map[string]any, variant string) {
	items, ok := data["Navigation"].([]any)
	if !ok {
		return
	}

	for _, item := range items {
		link, ok := item.(map[string]any)
		if !ok {
			continue
		}

		p.navigationLink(link, variant)
	}
}

func (p *Preview) navigationLink(link map[string]any, variant string) {
	view, ok := link["Key"].(string)
	if !ok || !p.validView(view) {
		link["Href"] = "#"
		return
	}

	href := "/?view=" + view
	if variant != "" {
		href += "&fixture=" + variant
	}

	link["Href"] = href
}

func (p *Preview) render(w http.ResponseWriter, r *http.Request, data map[string]any) {
	templateName := "page"
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		templateName = "content"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := p.bridge.Render(w, templateName, data); err != nil {
		http.Error(w, "preview rendering failed", http.StatusInternalServerError)
	}
}
