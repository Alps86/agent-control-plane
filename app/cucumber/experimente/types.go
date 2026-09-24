package experimente

import (
	"net/http"
	"net/http/httptest"
)

type Suite struct {
	server   *httptest.Server
	status   int
	header   http.Header
	body     string
	full     string
	rows     []SourceRow
	page     BrowserPage
	viewport int
}

type SourceRow struct {
	group string
	name  string
	links []string
	cells []string
}

type Inventory struct {
	rows []SourceRow
}

type BrowserPage struct {
	Title             string `json:"title"`
	Heading           string `json:"heading"`
	Groups            int    `json:"groups"`
	Entries           int    `json:"entries"`
	MobileWide        bool   `json:"mobileWide"`
	Source            string `json:"source"`
	FieldsComplete    bool   `json:"fieldsComplete"`
	StatusesComplete  bool   `json:"statusesComplete"`
	LegendComplete    bool   `json:"legendComplete"`
	DecisionsComplete bool   `json:"decisionsComplete"`
	NamedComplete     bool   `json:"namedComplete"`
	SourcesSafe       bool   `json:"sourcesSafe"`
	RelatedVisible    bool   `json:"relatedVisible"`
	LaterComplete     bool   `json:"laterComplete"`
	NoFalseClaim      bool   `json:"noFalseClaim"`
	HeadingsComplete  bool   `json:"headingsComplete"`
	AssetsLoaded      bool   `json:"assetsLoaded"`
	FragmentSame      bool   `json:"fragmentSame"`
	FragmentNoShell   bool   `json:"fragmentNoShell"`
}
