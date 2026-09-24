package projekt

import (
	"net"
	"net/http"
	"strconv"
)

// localWrite erlaubt Änderungen nur am konfigurierten Loopback-Listener.
func (h *Handler) localWrite(r *http.Request) bool {
	bindHost, bindPort, ok := h.localBind()
	if !ok || !h.requestHost(r.Host, bindPort) || !h.remoteLoopback(r.RemoteAddr) {
		return false
	}

	if !h.listenerLoopback(r, bindHost, bindPort) {
		return false
	}

	return h.sameOrigin(r)
}

func (h *Handler) localBind() (string, string, bool) {
	host, port, ok := h.localAddress(h.bindAddress)
	return host, port, ok && h.localName(host)
}

func (h *Handler) requestHost(address, bindPort string) bool {
	host, port, ok := h.localAddress(address)
	return ok && h.localName(host) && port == bindPort
}

func (h *Handler) remoteLoopback(address string) bool {
	host, _, err := net.SplitHostPort(address)
	return err == nil && h.loopbackIP(host)
}

func (h *Handler) listenerLoopback(r *http.Request, bindHost, bindPort string) bool {
	listener, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return false
	}

	host, port, ok := h.localAddress(listener.String())
	if !ok || !h.loopbackIP(host) || port != bindPort {
		return false
	}

	bindIP := net.ParseIP(bindHost)
	return bindIP == nil || bindIP.Equal(net.ParseIP(host))
}

func (h *Handler) localAddress(address string) (string, string, bool) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", "", false
	}

	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 || strconv.Itoa(value) != port {
		return "", "", false
	}

	return host, port, true
}

func (h *Handler) loopbackIP(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (h *Handler) localName(host string) bool {
	return host == "localhost" || h.loopbackIP(host)
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
