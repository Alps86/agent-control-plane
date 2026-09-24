package openrouterverbindung

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) registerRemoteSteps(sc *godog.ScenarioContext) {
	sc.Step(`^ein fremder TCP-Client "(GET|POST|DELETE)" an "([^"]+)" mit gefälschtem lokalem Host und passendem Origin sendet$`, s.remoteRequest)
	sc.Step(`^ich einen lokalen Settings-Server mit Bindadresse "([^"]+)" über "(GET|POST|DELETE)" an "([^"]+)" aufrufe$`, s.unsafeBindRequest)
	sc.Step(`^antwortet die öffentliche Grenze mit HTTP 403 ohne Schlüssel oder Verbindungsstatus$`, s.remoteDenied)
	sc.Step(`^die Verbindung bleibt über den lokalen Settings-Zugriff unverändert$`, s.localUnchanged)
}

func (s *Suite) unsafeBindRequest(bindAddress, method, path string) error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}

	handler, err := s.newHandler(bindAddress)
	if err != nil {
		listener.Close()
		return err
	}

	s.remoteServer = httptest.NewUnstartedServer(handler)
	s.remoteServer.Listener.Close()
	s.remoteServer.Listener = listener
	s.remoteServer.Start()
	return s.sendUnsafe(method, path)
}

func (s *Suite) sendUnsafe(method, path string) error {
	request, err := http.NewRequest(method, s.remoteServer.URL+path, s.remotePayload(method, path))
	if err != nil {
		return err
	}

	request.Header.Set("Origin", s.remoteServer.URL)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.remoteServer.Client().Do(request)
	if err != nil {
		return err
	}

	return s.readRemoteResponse(response)
}

func (s *Suite) readRemoteResponse(response *http.Response) error {
	defer response.Body.Close()
	s.remoteStatus = response.StatusCode
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	s.remoteBody = body
	return err
}

func (s *Suite) remoteRequest(method, path string) error {
	ip, err := s.nonLoopbackIP()
	if err != nil {
		return err
	}

	if err := s.startRemoteServer(); err != nil {
		return err
	}

	return s.sendRemote(method, path, ip)
}

func (s *Suite) nonLoopbackIP() (net.IP, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, address := range addresses {
		cidr, ok := address.(*net.IPNet)
		if ok && cidr.IP.To4() != nil && !cidr.IP.IsLoopback() && cidr.IP.IsGlobalUnicast() {
			return cidr.IP, nil
		}
	}

	return nil, fmt.Errorf("keine echte nicht lokale IPv4-Adresse für Remote-Peer-Abnahme verfügbar")
}

func (s *Suite) startRemoteServer() error {
	listener, err := net.Listen("tcp4", "0.0.0.0:0")
	if err != nil {
		return err
	}

	s.remotePeerAddr = make(chan string, 1)
	s.remoteServer = httptest.NewUnstartedServer(http.HandlerFunc(s.observeRemote))
	s.remoteServer.Listener.Close()
	s.remoteServer.Listener = listener
	s.remoteServer.Start()
	return nil
}

func (s *Suite) observeRemote(w http.ResponseWriter, r *http.Request) {
	s.remotePeerAddr <- r.RemoteAddr
	s.productHandler.ServeHTTP(w, r)
}

func (s *Suite) sendRemote(method, path string, ip net.IP) error {
	port := strconv.Itoa(s.remoteServer.Listener.Addr().(*net.TCPAddr).Port)
	host := net.JoinHostPort("localhost", port)
	url := "http://" + net.JoinHostPort(ip.String(), port) + path
	request, err := http.NewRequest(method, url, s.remotePayload(method, path))
	if err != nil {
		return err
	}

	request.Host = host
	request.Header.Set("Origin", "http://"+host)
	request.Header.Set("Content-Type", "application/json")
	return s.performRemote(request)
}

func (s *Suite) remotePayload(method, path string) io.Reader {
	if method == http.MethodPost && path == apiPath {
		return bytes.NewBufferString(`{"key":"` + replacementKey + `"}`)
	}

	return nil
}

func (s *Suite) performRemote(request *http.Request) error {
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if err := s.readRemoteResponse(response); err != nil {
		return err
	}

	return s.verifyRemotePeer()
}

func (s *Suite) verifyRemotePeer() error {
	select {
	case remote := <-s.remotePeerAddr:
		host, _, err := net.SplitHostPort(remote)
		if err != nil || net.ParseIP(host) == nil || net.ParseIP(host).IsLoopback() {
			return fmt.Errorf("kein echter nicht lokaler Remote-Peer: %q", remote)
		}
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("HTTP-Handler sah keinen Remote-Peer")
	}
}

func (s *Suite) remoteDenied() error {
	if s.remoteStatus != http.StatusForbidden {
		return fmt.Errorf("fremder Client erhielt HTTP %d: %s", s.remoteStatus, s.remoteBody)
	}

	body := string(s.remoteBody)
	if strings.Contains(body, s.currentKey) || strings.Contains(body, "openrouter-central") || strings.Contains(body, "nicht geprüft") {
		return fmt.Errorf("Settings-Daten an fremden Client ausgegeben")
	}

	return nil
}

func (s *Suite) localUnchanged() error {
	if err := s.getSettings(); err != nil {
		return err
	}

	if err := s.expectConnection(); err != nil {
		return err
	}

	return s.expectStatusText("nicht geprüft")
}
