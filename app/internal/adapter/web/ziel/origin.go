package ziel

import (
	"net"
	"net/http"
)

func (h *Handler) localWrite(r *http.Request) bool {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || !h.localName(host) {
		return false
	}

	if !h.localBind(port) || !h.localRemote(r.RemoteAddr) {
		return false
	}

	listener, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return false
	}

	listenerHost, actualPort, err := net.SplitHostPort(listener.String())
	return err == nil && h.localName(listenerHost) && port == actualPort && h.sameOrigin(r)
}

func (h *Handler) localBind(port string) bool {
	bindHost, bindPort, err := net.SplitHostPort(h.bindAddress)
	return err == nil && h.localName(bindHost) && port == bindPort
}

func (h *Handler) localRemote(address string) bool {
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
