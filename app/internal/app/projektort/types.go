package projektort

import (
	"errors"

	domainort "agentcontrolplane/app/internal/domain/projektort"
	portort "agentcontrolplane/app/internal/port/projektort"
)

const (
	CodeMissing            = "project_location_missing"
	CodeInvalid            = "project_location_invalid"
	CodeReady              = "project_location_ready"
	CodeAdapterUnavailable = "adapter_unavailable"
)

var ErrKindInvalid = domainort.ErrKindInvalid
var ErrNotFound = portort.ErrNotFound
var ErrAccessDenied = errors.New("access_denied")

// Status ist der öffentliche, pfadfreie Fehler- oder freigegebene Ortszustand.
type Status struct {
	Configured bool    `json:"configured"`
	Kind       *string `json:"kind"`
	Path       *string `json:"path"`
	StartReady bool    `json:"start_ready"`
	Code       string  `json:"code"`
	Reason     string  `json:"reason,omitempty"`
}

// Preparation ist nur eine Startvorbereitung, kein gestarteter Lauf.
type Preparation struct {
	Ready  bool    `json:"ready"`
	Code   string  `json:"code"`
	Path   *string `json:"path"`
	Reason string  `json:"reason,omitempty"`
}

// Service bindet Projektzugriff, Persistenz, Locator und Adapterprobe.
type Service struct {
	store    portort.Store
	projects portort.ProjectReader
	locator  portort.Locator
	adapter  portort.StartProbe
}
