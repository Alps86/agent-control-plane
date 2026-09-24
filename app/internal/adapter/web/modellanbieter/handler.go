package modellanbieter

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"agentcontrolplane/app/internal/app/modellverbindung"
)

func New(flow modellverbindung.DeviceFlow, bindAddress string) (*Handler, error) {
	handler := &Handler{flow: flow}
	handler.trustedBind = handler.validBind(bindAddress)
	if !handler.trustedBind {
		return nil, ErrUntrustedBind
	}

	return handler, nil
}

func (h *Handler) validBind(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil || !h.validPort(port) {
		return false
	}

	host = strings.ToLower(host)
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func (h *Handler) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /settings/modelle/codex/device/start", h.start)
	mux.HandleFunc("GET /settings/modelle/codex/device/status", h.status)
	mux.HandleFunc("POST /settings/modelle/codex/device/cancel", h.cancel)
	return mux
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	if !h.trustedBind || !h.allowedMutation(r) {
		h.forbid(w)
		return
	}

	status, err := h.flow.Start(r.Context())
	h.respond(w, status, err)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	if !h.trustedBind || !h.authorizedHost(r.Host) || !h.localPeer(r.RemoteAddr) {
		h.forbid(w)
		return
	}

	status, err := h.flow.Status(r.Context())
	h.respond(w, status, err)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	if !h.trustedBind || !h.allowedMutation(r) {
		h.forbid(w)
		return
	}

	status, err := h.flow.Cancel(r.Context())
	h.respond(w, status, err)
}

func (h *Handler) allowedMutation(r *http.Request) bool {
	if !h.authorizedHost(r.Host) || !h.localPeer(r.RemoteAddr) {
		return false
	}

	return h.sameOrigin(r)
}

func (h *Handler) localPeer(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func (h *Handler) authorizedHost(authority string) bool {
	host, port, ok := h.localHost(authority)
	if !ok || host == "" || (port != "" && !h.validPort(port)) {
		return false
	}

	return true
}

func (h *Handler) localHost(authority string) (string, string, bool) {
	host, port := authority, ""
	if strings.Contains(authority, ":") {
		var err error
		host, port, err = net.SplitHostPort(authority)
		if err != nil || port == "" {
			return "", "", false
		}
	}

	host = strings.ToLower(host)
	return host, port, host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func (h *Handler) validPort(raw string) bool {
	port, err := strconv.Atoi(raw)
	return err == nil && port > 0 && port <= 65535
}

func (h *Handler) sameOrigin(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) != 1 {
		return false
	}

	origin, err := url.Parse(origins[0])
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return err == nil && origin.Scheme == scheme && strings.EqualFold(origin.Host, r.Host) && origin.User == nil && origin.Path == "" && origin.RawQuery == "" && origin.Fragment == ""
}

func (h *Handler) forbid(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(modellverbindung.Status{State: "unavailable", Reason: "request_not_allowed"})
}

func (h *Handler) respond(w http.ResponseWriter, status modellverbindung.Status, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(modellverbindung.Status{State: "unavailable", Reason: "transport_unavailable"})
		return
	}

	_ = json.NewEncoder(w).Encode(status)
}
