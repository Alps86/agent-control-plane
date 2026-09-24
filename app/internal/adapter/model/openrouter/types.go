package openrouter

import (
	"encoding/json"
	"errors"
	"net/http"
)

const baseURL = "https://openrouter.ai"
const maxResponseBytes = 16 * 1024

var ErrMissingKey = errors.New("openrouter: key missing")
var ErrInvalidKey = errors.New("openrouter: key invalid")
var ErrUnavailable = errors.New("openrouter: key check unavailable")
var ErrInvalidResponse = errors.New("openrouter: invalid key check response")

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	http     HTTPDoer
	endpoint string
}

type Result struct {
	Valid bool
}

type keyEnvelope struct {
	Data map[string]json.RawMessage `json:"data"`
}
