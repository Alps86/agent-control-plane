package main

import (
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	modelopenrouter "agentcontrolplane/app/internal/adapter/model/openrouter"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
)

var errOpenRouterProbeStartup = errors.New("OpenRouter probe startup: invalid local endpoint")

func (b *Bootstrap) openRouterProbe() (openrouterverbindung.Probe, error) {
	raw := os.Getenv("APP_OPENROUTER_PROBE_URL")
	if raw == "" {
		return modelopenrouter.NewProbe(nil), nil
	}

	base, valid := b.openRouterProbeBase(raw)
	if !valid {
		return nil, errOpenRouterProbeStartup
	}

	return modelopenrouter.NewProbeAt(nil, base), nil
}

func (b *Bootstrap) openRouterProbeBase(raw string) (string, bool) {
	parsed, err := url.Parse(raw)
	if err != nil || !strings.EqualFold(parsed.Scheme, "http") || parsed.Opaque != "" {
		return "", false
	}

	if parsed.User != nil || strings.ContainsAny(raw, "?#") || parsed.RawPath != "" || parsed.Path != "" {
		return "", false
	}

	host, port := parsed.Hostname(), parsed.Port()
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() || port == "" || strings.Trim(port, "0123456789") != "" {
		return "", false
	}

	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 || parsed.Host != net.JoinHostPort(host, port) {
		return "", false
	}

	return "http://" + net.JoinHostPort(ip.String(), strconv.Itoa(number)), true
}
