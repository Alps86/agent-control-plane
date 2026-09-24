package bridge

import (
	"io/fs"
	"net/http"

	"agentcontrolplane/ui/internal/render"
)

// Bridge is the sole public Go boundary for embedded UI resources.
type Bridge struct {
	source    fs.FS
	renderer  *render.Renderer
	assets    http.Handler
	fragments http.Handler
}
