package codexprofil

import (
	"net"
	"net/http"
	"strconv"
)

func (h *Handler) localWrite(r *http.Request) bool {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || !h.localName(host) || !h.localListener(r, port) {
		return false
	}

	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	return err == nil && h.loopbackIP(peer) && h.sameOrigin(r)
}

func (h *Handler) localListener(r *http.Request, port string) bool {
	listener, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return false
	}

	host, actualPort, err := net.SplitHostPort(listener.String())
	return err == nil && port == actualPort && h.localBind(host, actualPort)
}

func (h *Handler) localBind(host, port string) bool {
	bindHost, bindPort, err := net.SplitHostPort(h.bindAddress)
	if err != nil || !h.loopbackIP(bindHost) || !h.loopbackIP(host) {
		return false
	}

	parsedPort, err := strconv.Atoi(bindPort)
	if err != nil || parsedPort < 0 || parsedPort > 65535 {
		return false
	}

	return net.ParseIP(bindHost).Equal(net.ParseIP(host)) && (parsedPort == 0 || bindPort == port)
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
