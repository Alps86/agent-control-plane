package codexcli

// Probe bewertet nur die konfigurierte CLI-Referenz und den belegten
// Durchsetzungsstand. Er startet keinen Prozess und liest keine Anmeldung.
type Probe struct {
	cliReference string
}
