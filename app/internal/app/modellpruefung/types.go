package modellpruefung

import (
	"context"
	"sync"

	"agentcontrolplane/app/internal/app/modellverbindung"
)

// StatusAktion ist die einzige in diesem Prüfpfad registrierbare Fachaktion.
const StatusAktion = "Projektstatus lesen"

// VerbindungsstatusAktion ist die freigegebene reine Leseaktion der HTTP-Probe.
const VerbindungsstatusAktion = "Verbindungsstatus lesen"

// VerbindungsstatusQuelle liefert den öffentlichen Zustand ohne Zugangsdaten.
type VerbindungsstatusQuelle interface {
	Status(context.Context) (modellverbindung.Status, error)
}

// Verbindungszustand enthält nur den redigierten öffentlichen Zustand.
type Verbindungszustand struct {
	Zustand string `json:"zustand"`
}

// Verbindungspruefung begrenzt die Probe auf eine benannte Leseaktion.
type Verbindungspruefung struct {
	quelle VerbindungsstatusQuelle
}

// StatusQuelle liefert den fachlichen Projektstatus nach der Rechteprüfung.
type StatusQuelle interface {
	Status(context.Context, string) (Status, error)
}

// Status ist das anbieterneutrale Ergebnis eines Statuszugriffs.
type Status struct {
	Projekt      string
	Wert         string
	Pruefkennung string
}

// Aufruf dokumentiert die fachliche Entscheidung ohne Modellgeheimnisse.
type Aufruf struct {
	Agent       string
	Aktion      string
	Projekt     string
	Berechtigt  bool
	Ausgefuehrt bool
	Grund       string
}

// Pruefung validiert Fachaktion, Parameter und Recht vor jedem Zugriff.
type Pruefung struct {
	quelle  StatusQuelle
	mu      sync.Mutex
	rechte  map[string]map[string]bool
	aufrufe []Aufruf
}

// Transportbeleg enthält nur redigierte, anbieterneutrale Nachweismerkmale.
type Transportbeleg struct {
	Anbieter       string
	Modell         string
	Anfragekennung string
	Antwort        bool
	Stream         bool
	Toolrunde      bool
	Folgeaufruf    bool
	Abbruch        bool
	Limitfehler    bool
}

// Nachweisstatus ist die öffentliche Bewertung des technischen Gates.
type Nachweisstatus struct {
	Offen         bool
	Agentenlaeufe int
	Modellbelege  int
}

// Nachweis bewertet Modelltransport getrennt von Codex-Agentenläufen.
type Nachweis struct {
	mu            sync.Mutex
	agentenlaeufe int
	modellbelege  int
	vollstaendig  bool
}
