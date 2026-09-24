package render

import "html/template"

// Renderer renders named templates from one immutable source tree.
type Renderer struct {
	templates *template.Template
	files     []string
}
