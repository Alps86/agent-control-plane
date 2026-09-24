package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
)

// New parses the templates in the supplied build filesystem.
func New(source fs.FS) (*Renderer, error) {
	renderer := &Renderer{}
	if err := renderer.load(source); err != nil {
		return nil, err
	}

	return renderer, nil
}

func (r *Renderer) load(source fs.FS) error {
	if err := r.collect(source); err != nil {
		return fmt.Errorf("walk UI templates: %w", err)
	}

	if len(r.files) == 0 {
		return fmt.Errorf("no UI templates found")
	}

	parsed, err := template.ParseFS(source, r.files...)
	if err != nil {
		return fmt.Errorf("parse UI templates: %w", err)
	}

	r.templates = parsed
	return nil
}

func (r *Renderer) collect(source fs.FS) error {
	return fs.WalkDir(source, "templates", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && strings.HasSuffix(path, ".html") {
			r.files = append(r.files, path)
		}

		return nil
	})
}

// Render writes a complete named template only after execution succeeds.
func (r *Renderer) Render(w io.Writer, templateName string, data map[string]any) error {
	if strings.HasSuffix(templateName, ".html") || r.templates.Lookup(templateName) == nil {
		return fmt.Errorf("unknown UI template %q", templateName)
	}

	var output bytes.Buffer
	if err := r.templates.ExecuteTemplate(&output, templateName, data); err != nil {
		return fmt.Errorf("render UI template %q: %w", templateName, err)
	}

	if _, err := io.Copy(w, &output); err != nil {
		return fmt.Errorf("write UI template %q: %w", templateName, err)
	}

	return nil
}
