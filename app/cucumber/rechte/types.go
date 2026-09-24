package rechte

import (
	"net/http/httptest"

	domainrechte "agentcontrolplane/app/internal/domain/rechte"
)

// Suite hält nur die über HTTP beobachteten Ergebnisse eines Szenarios.
type Suite struct {
	directory *Directory
	server    *httptest.Server
	status    int
	body      []byte
	oldStatus int
	oldBody   []byte
}

// Directory ist die ausschließlich serverseitig konfigurierte Testquelle.
type Directory struct {
	actors        []domainrechte.Actor
	organizations map[string][]string
	resources     map[string]domainrechte.Resource
	assignments   map[string]map[string]bool
}
