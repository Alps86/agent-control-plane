package modellpruefung

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (h *Handler) validBind() bool {
	host, port, err := net.SplitHostPort(h.bindAddress)
	if err != nil || !h.loopback(host) {
		return false
	}

	value, err := strconv.Atoi(port)
	return err == nil && value >= 0 && value <= 65535
}

func (h *Handler) allowed(r *http.Request) bool {
	if !h.validHost(r) || !h.localPeer(r.RemoteAddr) {
		return false
	}

	return h.sameOrigin(r)
}

func (h *Handler) validHost(r *http.Request) bool {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || !h.loopback(host) {
		return false
	}

	_, bindPort, _ := net.SplitHostPort(h.bindAddress)
	if bindPort != "0" && port != bindPort {
		return false
	}

	return h.localListener(r, port, bindPort == "0")
}

func (h *Handler) localListener(r *http.Request, port string, required bool) bool {
	listener, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return !required
	}

	host, actual, err := net.SplitHostPort(listener.String())
	return err == nil && h.loopback(host) && actual == port
}

func (h *Handler) localPeer(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	return err == nil && h.loopback(host)
}

func (h *Handler) loopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func (h *Handler) sameOrigin(r *http.Request) bool {
	values := r.Header.Values("Origin")
	if len(values) != 1 {
		return false
	}

	parsed, err := url.Parse(values[0])
	if err != nil || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Opaque != "" {
		return false
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return parsed.Scheme == scheme && strings.EqualFold(parsed.Host, r.Host)
}
