package projekte

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) registerStartupSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ich den Server mit derselben SQLite-Datenbank und APP_ADDR "([^"]*)" neu starte$`, s.startRejectedAddress)
	sc.Step(`^beendet sich der Server vor dem Listener mit einem konkreten APP_ADDR-Loopback-Fehler$`, s.rejectedWithAddressError)
	sc.Step(`^die abgewiesene Bindadresse nimmt keine HTTP-Verbindung an$`, s.rejectedAddressHasNoListener)
	sc.Step(`^ich den Server beende und mit derselben SQLite-Datenbank auf der Loopback-Adresse erneut starte$`, s.restartLoopback)
}

func (s *Suite) startRejectedAddress(raw string) error {
	s.stopServer()
	s.rejectedAddr = strings.ReplaceAll(raw, "<Port>", s.listenerPort())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.binary)
	cmd.Env = append(os.Environ(), "APP_ADDR="+s.rejectedAddr, "APP_DB_PATH="+s.database)
	output, err := cmd.CombinedOutput()
	s.rejectedLog, s.rejectedErr = string(output), err
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("Server beendete sich bei APP_ADDR %q nicht vor dem Listener", s.rejectedAddr)
	}

	return nil
}

func (s *Suite) rejectedWithAddressError() error {
	if s.rejectedErr == nil || !strings.Contains(s.rejectedLog, "APP_ADDR must be a local loopback address") {
		return fmt.Errorf("kein konkreter APP_ADDR-Loopback-Fehler bei %q: %v: %s", s.rejectedAddr, s.rejectedErr, s.rejectedLog)
	}

	return nil
}

func (s *Suite) rejectedAddressHasNoListener() error {
	if err := s.rejectedWithAddressError(); err != nil {
		return err
	}

	addresses := []string{s.rejectedProbeAddress()}
	if addresses[0] != s.address {
		addresses = append(addresses, s.address)
	}

	for _, address := range addresses {
		connection, err := net.DialTimeout("tcp", address, 150*time.Millisecond)
		if err == nil {
			connection.Close()
			return fmt.Errorf("abgewiesene APP_ADDR %q ließ Listener auf %s zurück", s.rejectedAddr, address)
		}
	}

	return nil
}

func (s *Suite) rejectedProbeAddress() string {
	host, port, err := net.SplitHostPort(s.rejectedAddr)
	if err != nil || net.ParseIP(host) == nil {
		return s.address
	}

	if host == "0.0.0.0" {
		host = "127.0.0.1"
	}

	if host == "::" {
		host = "::1"
	}

	return net.JoinHostPort(host, port)
}

func (s *Suite) restartLoopback() error {
	s.stopServer()
	s.bindHost = "127.0.0.1"
	return s.startServer()
}
