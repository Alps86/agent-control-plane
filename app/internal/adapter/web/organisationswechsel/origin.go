package organisationswechsel

import (
	"net"
	"net/http"
)

func (h *Handler) localWrite(r *http.Request) bool {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || !h.localName(host) || !h.loopbackAddr(h.bindAddress) {
		return false
	}

	listener, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok || !h.loopbackAddr(listener.String()) || !h.loopbackAddr(r.RemoteAddr) {
		return false
	}

	_, actualPort, err := net.SplitHostPort(listener.String())
	return err == nil && port == actualPort && h.sameOrigin(r)
}

func (h *Handler) loopbackAddr(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (h *Handler) localName(host string) bool {
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}

func (h *Handler) sameOrigin(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) == 0 {
		return true
	}

	if len(origins) != 1 {
		return false
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return origins[0] == scheme+"://"+r.Host
}
