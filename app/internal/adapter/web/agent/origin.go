package agent

import (
	"net"
	"net/http"
	"strconv"
)

// localWrite binds every mutation to the trusted local bind, peer, and origin.
func (h *Handler) localWrite(r *http.Request) bool {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || !h.localName(host) {
		return false
	}

	if !h.localListener(r, port) {
		return false
	}

	peerHost, _, err := net.SplitHostPort(r.RemoteAddr)
	return err == nil && h.loopbackIP(peerHost) && h.sameOrigin(r)
}

func (h *Handler) localListener(r *http.Request, requestPort string) bool {
	listener, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return false
	}

	listenerHost, actualPort, err := net.SplitHostPort(listener.String())
	return err == nil && requestPort == actualPort && h.localBind(listenerHost, actualPort)
}

func (h *Handler) localBind(listenerHost, actualPort string) bool {
	bindHost, bindPort, err := net.SplitHostPort(h.bindAddress)
	if err != nil || !h.loopbackIP(bindHost) || !h.loopbackIP(listenerHost) {
		return false
	}

	port, err := strconv.Atoi(bindPort)
	if err != nil || port < 0 || port > 65535 {
		return false
	}

	return net.ParseIP(bindHost).Equal(net.ParseIP(listenerHost)) && (port == 0 || bindPort == actualPort)
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
