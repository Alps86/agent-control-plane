package projektort

import portort "agentcontrolplane/app/internal/port/projektort"

var ErrIntegrity = portort.ErrIntegrity

// Locator leitet den Projektort ausschließlich aus dem Datenbankpfad und IDs ab.
type Locator struct {
	databasePath string
}

// StartProbe prüft die Übergabe des freigegebenen Orts an die Adaptergrenze.
type StartProbe struct {
	locator *Locator
}
