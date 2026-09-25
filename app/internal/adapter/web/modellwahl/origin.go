package modellwahl

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

func (h *Handler) localListener(r *http.Request, requestPort string) bool {
	listener, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return false
	}

	host, port, err := net.SplitHostPort(listener.String())
	return err == nil && requestPort == port && h.localBind(host, port)
}

func (h *Handler) localBind(listenerHost, actualPort string) bool {
	host, port, err := net.SplitHostPort(h.bindAddress)
	if err != nil || !h.loopbackIP(host) || !h.loopbackIP(listenerHost) {
		return false
	}

	number, err := strconv.Atoi(port)
	return err == nil && number >= 0 && number <= 65535 && net.ParseIP(host).Equal(net.ParseIP(listenerHost)) && (number == 0 || port == actualPort)
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
