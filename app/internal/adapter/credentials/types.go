package credentials

import "sync"

// Store speichert verschlüsselte Credentials in einer lokalen Datei.
type Store struct {
	secretPath string
	keyPath    string
	key        []byte
	mu         sync.Mutex
}
