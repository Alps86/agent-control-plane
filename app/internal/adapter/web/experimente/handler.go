package experimente

import (
	"bytes"
	"net/http"
	"strings"
)

// NewHandler baut den öffentlichen Leseweg mit der UI-Bridge.
func NewHandler(renderer Renderer) http.Handler {
	return &Handler{renderer: renderer, loader: inventoryLoader{}}
}

// ServeHTTP rendert eine Vollseite oder das passende HTMX-Fragment.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	value, err := h.loader.load()
	if err != nil || h.renderer == nil {
		http.Error(w, "inventory unavailable", http.StatusInternalServerError)
		return
	}

	h.render(w, r, value)
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, value inventory) {
	name := "experimente/page"
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		name = "experimente/content"
	}

	var output bytes.Buffer
	if err := h.renderer.Render(&output, name, value.viewMap()); err != nil {
		http.Error(w, "inventory rendering failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output.Bytes())
}
