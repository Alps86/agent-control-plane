package openrouter

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (h *Handler) localRequest(r *http.Request) bool {
	if !h.loopbackHost(r.Host) {
		return false
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}

	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func (h *Handler) numericLoopbackBind(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}

	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return false
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (h *Handler) originAllowed(r *http.Request) bool {
	if !h.loopbackHost(r.Host) {
		return false
	}

	if len(r.Header.Values("Origin")) != 1 {
		return false
	}

	origin, err := url.Parse(r.Header.Get("Origin"))
	if err != nil || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" || origin.Opaque != "" {
		return false
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return origin.Scheme == scheme && strings.EqualFold(origin.Host, r.Host)
}

func (h *Handler) loopbackHost(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}

	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return false
	}

	ip := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || ip != nil && ip.IsLoopback()
}
