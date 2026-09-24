package uipreview

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func (s *Suite) loopbackOnly() error {
	connection, err := net.DialTimeout("tcp4", "127.0.0.1:"+strconv.Itoa(s.port), time.Second)
	if err != nil {
		return err
	}

	return connection.Close()
}

func (s *Suite) reachable() error {
	result, err := s.fetch("/?view=organization", false)
	if err != nil || result.Status != http.StatusOK {
		return fmt.Errorf("loopback preview: %v, HTTP %d", err, result.Status)
	}

	return nil
}

func (s *Suite) nonLoopbackDenied() error {
	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}

	for _, address := range interfaces {
		ip, _, err := net.ParseCIDR(address.String())
		if err == nil && ip.To4() != nil && !ip.IsLoopback() {
			return s.denyAddress(ip.String())
		}
	}

	return fmt.Errorf("no non-loopback IPv4 interface available for access check")
}

func (s *Suite) denyAddress(ip string) error {
	connection, err := net.DialTimeout("tcp4", net.JoinHostPort(ip, strconv.Itoa(s.port)), time.Second)
	if err == nil {
		connection.Close()
		return fmt.Errorf("preview reachable via non-loopback address %s", ip)
	}

	return nil
}

func (s *Suite) invalidPort() error {
	s.portErrors = nil
	for _, port := range []string{"invalid", strconv.Itoa(s.port)} {
		command := exec.Command(s.binary)
		command.Env = append(os.Environ(), "ACP_PREVIEW_PORT="+port)
		output, err := command.CombinedOutput()
		if err == nil {
			return fmt.Errorf("port %q unexpectedly started", port)
		}

		s.portErrors = append(s.portErrors, string(output))
	}

	return nil
}

func (s *Suite) startError() error {
	if len(s.portErrors) != 2 || !strings.Contains(s.portErrors[0], "ACP_PREVIEW_PORT") || !strings.Contains(s.portErrors[1], "listen") {
		return fmt.Errorf("missing useful port errors: %v", s.portErrors)
	}

	return nil
}

func (s *Suite) noSecondServer() error {
	return s.reachable()
}

func (s *Suite) assets() error {
	page, err := s.browserCommand(map[string]any{"assets": true})
	if err != nil {
		return err
	}

	if len(page.Assets) < 2 {
		return fmt.Errorf("browser saw only %d assets", len(page.Assets))
	}

	for _, asset := range page.Assets {
		if asset.Status != 200 {
			return fmt.Errorf("asset %s: HTTP %d", asset.Path, asset.Status)
		}
	}

	return nil
}

func (s *Suite) openHelp() error {
	page, err := s.browserCommand(map[string]any{"help": true})
	s.page = page
	return err
}

func (s *Suite) helpFragment() error {
	result, err := s.fetch("/fragments/preview-help.html", false)
	if err != nil {
		return err
	}

	if result.Status != 200 || strings.Contains(result.Body, "<html") || s.page.Help == "" {
		return fmt.Errorf("help fragment: HTTP %d, browser text %q", result.Status, s.page.Help)
	}

	return nil
}

func (s *Suite) stopServer() {
	if s.server != nil && s.server.Process != nil {
		s.server.Process.Kill()
		s.server.Wait()
	}
}
