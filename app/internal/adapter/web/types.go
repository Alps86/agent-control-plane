package web

import "agentcontrolplane/app/internal/port/system"

// Server stellt öffentliche HTTP-Routen bereit.
type Server struct {
	probe system.Probe
}
