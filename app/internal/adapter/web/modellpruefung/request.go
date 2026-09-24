package modellpruefung

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
)

func (h *Handler) emptyBody(r *http.Request) bool {
	if r.URL.RawQuery != "" || !h.jsonContentType(r) {
		return false
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 65))
	if err != nil || len(body) > 64 {
		return false
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	var fields map[string]json.RawMessage
	if decoder.Decode(&fields) != nil || fields == nil || len(fields) != 0 {
		return false
	}

	var trailing any
	return decoder.Decode(&trailing) == io.EOF
}

func (h *Handler) jsonContentType(r *http.Request) bool {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && media == "application/json"
}
