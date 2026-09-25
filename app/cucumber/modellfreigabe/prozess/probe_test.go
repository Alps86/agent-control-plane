package prozess

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

const syntheticKey = "sk-or-story24-PROCESS-TEST-ONLY-582741"
const settingsPath = "/api/settings/modellanbieter/openrouter"

type providerRequest struct {
	method, path, authorization string
	body                        []byte
}

type suite struct {
	t                                                               *testing.T
	binary, address, probeURL, dbPath, secretPath, keyPath, logPath string
	process                                                         *exec.Cmd
	processDone                                                     chan error
	provider, redirect, proxy                                       *http.Server
	providerListener, redirectListener, proxyListener               net.Listener
	mu                                                              sync.Mutex
	requests                                                        []providerRequest
	redirectRequests                                                []providerRequest
	proxyRequests                                                   []providerRequest
	responseBodies                                                  [][]byte
	lastStatus                                                      int
	lastBody                                                        []byte
	reference, orgID, agentID, orgName, agentName                   string
}

func (s *suite) initialize(sc *godog.ScenarioContext) {
	sc.Before(func(context.Context, *godog.Scenario) (context.Context, error) {
		s.reset()
		return context.Background(), nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		s.stop()
		return ctx, nil
	})
	sc.Step(`^ein lokaler synthetischer OpenRouter-Anbieter hört auf numerischem (IPv4|IPv6)-Loopback und antwortet auf GET "/api/v1/key" mit HTTP 200 und \{"data":\{\}\}$`, s.readyProvider)
	sc.Step(`^APP_OPENROUTER_PROBE_URL enthält seine HTTP-Adresse mit explizitem Port$`, s.providerURL)
	sc.Step(`^ein gebauter Server mit isolierter SQLite-, Credential- und Masterkey-Datei ist gestartet$`, s.startHealthy)
	sc.Step(`^ich einen synthetischen Schlüssel über die öffentliche Settings-API speichere und prüfe$`, s.saveAndCheck)
	sc.Step(`^empfängt der lokale Anbieter genau einen GET "/api/v1/key" mit diesem Schlüssel im Bearer-Header und ohne Requestbody$`, s.readyRequest)
	sc.Step(`^die öffentliche Settings-API meldet "einsatzbereit" und eine Verbindungsreferenz ohne Schlüssel$`, s.readySettings)
	sc.Step(`^ich über die öffentlichen APIs die Organisation "([^"]+)" und darin den Eino-Agenten "([^"]+)" anlege$`, s.createOrganizationAndAgent)
	sc.Step(`^ich die zentrale OpenRouter-Verbindung erst für "([^"]+)" und dann für "([^"]+)" öffentlich freigebe$`, s.grantBoth)
	sc.Step(`^meldet die öffentliche Freigabe-API beide Freigaben für dieselbe Verbindungsreferenz$`, s.grantsVisible)
	sc.Step(`^Settings-, Freigabe- und Prozessausgaben enthalten den synthetischen Schlüssel nicht$`, s.noSecret)
	sc.Step(`^APP_OPENROUTER_PROBE_URL ist "([^"]*)"$`, s.invalidURL)
	sc.Step(`^ich den gebauten Server mit isolierter SQLite-, Credential- und Masterkey-Datei starte$`, s.startInvalid)
	sc.Step(`^beendet er den Start mit einem Fehler vor der öffentlichen HTTP-Bereitschaft$`, s.startRejected)
	sc.Step(`^es wurde keine Anfrage an einen OpenRouter-Anbieter gesendet$`, s.noProviderRequest)
	sc.Step(`^die Prozessausgabe enthält keinen synthetischen Schlüssel$`, s.logNoSecret)
	sc.Step(`^ein lokaler synthetischer OpenRouter-Anbieter antwortet auf GET "/api/v1/key" mit HTTP (302|307) und einem Redirect zu einem zweiten lokalen Empfänger$`, s.redirectProvider)
	sc.Step(`^APP_OPENROUTER_PROBE_URL enthält die numerische IPv4-Loopback-HTTP-Adresse des ersten Anbieters mit explizitem Port$`, s.providerURL)
	sc.Step(`^empfängt nur der erste Anbieter genau einen GET "/api/v1/key" mit diesem Schlüssel im Bearer-Header$`, s.redirectRequest)
	sc.Step(`^der zweite Empfänger erhält keine Anfrage und keinen Schlüssel$`, s.redirectUntouched)
	sc.Step(`^die öffentliche Settings-API meldet nicht "einsatzbereit"$`, s.notReadySettings)
	sc.Step(`^Settings- und Prozessausgaben enthalten den synthetischen Schlüssel nicht$`, s.noSecret)
}

func (s *suite) reset() {
	s.address, s.probeURL, s.reference, s.orgID, s.agentID, s.orgName, s.agentName = "", "", "", "", "", "", ""
	s.lastBody, s.responseBodies = nil, nil
	s.mu.Lock()
	s.requests, s.redirectRequests, s.proxyRequests = nil, nil, nil
	s.mu.Unlock()
}

func (s *suite) build() error {
	s.binary = filepath.Join(s.t.TempDir(), "server")
	cmd := exec.Command("go", "build", "-o", s.binary, "./cmd/server")
	cmd.Dir = "../../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Serverbau: %w: %s", err, out)
	}
	return nil
}

func (s *suite) listen(network, address string, handler http.HandlerFunc) (*http.Server, net.Listener, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, nil, err
	}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	return server, listener, nil
}

func (s *suite) readyProvider(ipType string) error {
	network, address := "tcp4", "127.0.0.1:0"
	if ipType == "IPv6" {
		network, address = "tcp6", "[::1]:0"
	}
	var err error
	s.provider, s.providerListener, err = s.listen(network, address, func(w http.ResponseWriter, r *http.Request) {
		s.record(r, false)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{}}`)
	})
	if err != nil {
		return err
	}
	s.probeURL = "http://" + s.providerListener.Addr().String()
	return nil
}

func (s *suite) redirectProvider(status int) error {
	var err error
	s.redirect, s.redirectListener, err = s.listen("tcp4", "127.0.0.1:0", func(w http.ResponseWriter, r *http.Request) { s.record(r, true); w.WriteHeader(http.StatusNoContent) })
	if err != nil {
		return err
	}
	s.provider, s.providerListener, err = s.listen("tcp4", "127.0.0.1:0", func(w http.ResponseWriter, r *http.Request) {
		s.record(r, false)
		w.Header().Set("Location", "http://"+s.redirectListener.Addr().String()+"/capture")
		w.WriteHeader(status)
	})
	if err != nil {
		return err
	}
	s.probeURL = "http://" + s.providerListener.Addr().String()
	return nil
}

func (s *suite) record(r *http.Request, redirect bool) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1024))
	call := providerRequest{r.Method, r.URL.Path, r.Header.Get("Authorization"), body}
	s.mu.Lock()
	defer s.mu.Unlock()
	if redirect {
		s.redirectRequests = append(s.redirectRequests, call)
	} else {
		s.requests = append(s.requests, call)
	}
}

func (s *suite) providerURL() error {
	if s.probeURL == "" {
		return fmt.Errorf("Provideradresse fehlt")
	}
	return nil
}

func (s *suite) invalidURL(raw string) error { s.probeURL = raw; return nil }

func (s *suite) prepareProcess() error {
	dir := s.t.TempDir()
	s.dbPath = filepath.Join(dir, "data", "app.sqlite")
	s.secretPath = filepath.Join(dir, "credentials", "secrets.enc")
	s.keyPath = filepath.Join(dir, "master", "master.key")
	s.logPath = filepath.Join(dir, "server.log")
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.address = listener.Addr().String()
	_ = listener.Close()
	s.proxy, s.proxyListener, err = s.listen("tcp4", "127.0.0.1:0", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.proxyRequests = append(s.proxyRequests, providerRequest{method: r.Method, path: r.URL.String(), authorization: r.Header.Get("Authorization")})
		s.mu.Unlock()
		http.Error(w, "external access blocked", http.StatusBadGateway)
	})
	return err
}

func (s *suite) launch() error {
	if err := os.MkdirAll(filepath.Dir(s.dbPath), 0700); err != nil {
		return err
	}
	log, err := os.Create(s.logPath)
	if err != nil {
		return err
	}
	s.process = exec.Command(s.binary)
	s.process.Env = append(os.Environ(), "APP_ADDR="+s.address, "APP_DB_PATH="+s.dbPath,
		"APP_CREDENTIALS_PATH="+s.secretPath, "APP_CREDENTIAL_KEY_PATH="+s.keyPath,
		"APP_OPENROUTER_PROBE_URL="+s.probeURL, "HTTP_PROXY=http://"+s.proxyListener.Addr().String(),
		"HTTPS_PROXY=http://"+s.proxyListener.Addr().String(), "ALL_PROXY=http://"+s.proxyListener.Addr().String(),
		"NO_PROXY=127.0.0.1,::1,[::1]", "no_proxy=127.0.0.1,::1,[::1]")
	s.process.Stdout, s.process.Stderr = log, log
	if err := s.process.Start(); err != nil {
		log.Close()
		return err
	}
	s.processDone = make(chan error, 1)
	go func() { s.processDone <- s.process.Wait(); _ = log.Close() }()
	return nil
}

func (s *suite) startHealthy() error {
	if err := s.prepareProcess(); err != nil {
		return err
	}
	if err := s.launch(); err != nil {
		return err
	}
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-s.processDone:
			return fmt.Errorf("Server vor Bereitschaft beendet: %v: %s", err, s.logs())
		default:
		}
		status, _, err := s.request(http.MethodGet, "/health", nil)
		if err == nil && status == http.StatusOK {
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	return fmt.Errorf("Server nicht bereit: %s", s.logs())
}

func (s *suite) startInvalid() error {
	if err := s.prepareProcess(); err != nil {
		return err
	}
	return s.launch()
}

func (s *suite) startRejected() error {
	select {
	case err := <-s.processDone:
		if err == nil {
			return fmt.Errorf("ungültiger Start erfolgreich")
		}
	case <-time.After(6 * time.Second):
		return fmt.Errorf("Server blieb bei ungültiger Probeadresse aktiv")
	}
	if !bytes.Contains(s.logs(), []byte("OpenRouter probe startup: invalid local endpoint")) {
		return fmt.Errorf("Startfehler lag nicht an der Probeadresse: %s", s.logs())
	}
	client := &http.Client{Timeout: 200 * time.Millisecond}
	response, err := client.Get("http://" + s.address + "/health")
	if err == nil {
		response.Body.Close()
		return fmt.Errorf("HTTP-Bereitschaft trotz Startfehler: %d", response.StatusCode)
	}
	return nil
}

func (s *suite) stop() {
	if s.process != nil && s.process.ProcessState == nil {
		_ = s.process.Process.Kill()
		<-s.processDone
	}
	s.process = nil
	for _, server := range []*http.Server{s.provider, s.redirect, s.proxy} {
		if server != nil {
			_ = server.Close()
		}
	}
	s.provider, s.redirect, s.proxy = nil, nil, nil
}

func (s *suite) logs() []byte { data, _ := os.ReadFile(s.logPath); return data }

func (s *suite) request(method, path string, body []byte) (int, []byte, error) {
	req, err := http.NewRequest(method, "http://"+s.address+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	if method != http.MethodGet {
		req.Header.Set("Origin", "http://"+s.address)
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err == nil {
		s.responseBodies = append(s.responseBodies, bytes.Clone(data))
		s.lastStatus, s.lastBody = resp.StatusCode, data
	}
	return resp.StatusCode, data, err
}

func (s *suite) expect(method, path string, body []byte, want int) ([]byte, error) {
	status, data, err := s.request(method, path, body)
	if err != nil {
		return nil, err
	}
	if status != want {
		return nil, fmt.Errorf("%s %s: HTTP %d statt %d: %s", method, path, status, want, data)
	}
	return data, nil
}

func (s *suite) saveAndCheck() error {
	key, _ := json.Marshal(map[string]string{"key": syntheticKey})
	if _, err := s.expect(http.MethodPost, settingsPath, key, http.StatusOK); err != nil {
		return err
	}
	_, err := s.expect(http.MethodPost, settingsPath+"/pruefen", nil, http.StatusOK)
	return err
}

func (s *suite) calls() ([]providerRequest, []providerRequest, []providerRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]providerRequest(nil), s.requests...), append([]providerRequest(nil), s.redirectRequests...), append([]providerRequest(nil), s.proxyRequests...)
}

func (s *suite) oneKeyRequest() error {
	first, _, proxy := s.calls()
	if len(first) != 1 || len(proxy) != 0 {
		return fmt.Errorf("Anbieteranfragen: erster=%d Sperrproxy=%d", len(first), len(proxy))
	}
	r := first[0]
	if r.method != http.MethodGet || r.path != "/api/v1/key" || r.authorization != "Bearer "+syntheticKey || len(r.body) != 0 {
		return fmt.Errorf("unerwarteter Methoden-/Pfad-/Header-/Body-Vertrag: %s %s, Bearer=%v, Bodybytes=%d", r.method, r.path, r.authorization == "Bearer "+syntheticKey, len(r.body))
	}
	return nil
}

func (s *suite) readyRequest() error    { return s.oneKeyRequest() }
func (s *suite) redirectRequest() error { return s.oneKeyRequest() }

func (s *suite) settings() (struct {
	Connected         bool `json:"connected"`
	Reference, Status string
}, error) {
	var view struct {
		Connected         bool `json:"connected"`
		Reference, Status string
	}
	body, err := s.expect(http.MethodGet, settingsPath, nil, http.StatusOK)
	if err != nil {
		return view, err
	}
	err = json.Unmarshal(body, &view)
	return view, err
}

func (s *suite) readySettings() error {
	view, err := s.settings()
	if err != nil {
		return err
	}
	if !view.Connected || view.Status != "einsatzbereit" || view.Reference == "" {
		return fmt.Errorf("unzulässiger Settings-Status: %+v", view)
	}
	s.reference = view.Reference
	return nil
}

func (s *suite) notReadySettings() error {
	view, err := s.settings()
	if err != nil {
		return err
	}
	if view.Status == "einsatzbereit" {
		return fmt.Errorf("Redirect wurde als einsatzbereit gemeldet")
	}
	return nil
}

func (s *suite) createOrganizationAndAgent(org, agent string) error {
	payload, _ := json.Marshal(map[string]string{"name": org, "description": "Lokaler Probetest"})
	body, err := s.expect(http.MethodPost, "/api/organisationen", payload, http.StatusCreated)
	if err != nil {
		return err
	}
	var created struct{ ID, Name string }
	if err := json.Unmarshal(body, &created); err != nil {
		return err
	}
	if created.ID == "" || created.Name != org {
		return fmt.Errorf("Organisation nicht angelegt: %s", body)
	}
	s.orgID = created.ID
	s.orgName = org
	payload, _ = json.Marshal(map[string]string{"name": agent, "template_id": "recherche", "execution_kind": "eino"})
	body, err = s.expect(http.MethodPost, "/api/organisationen/"+url.PathEscape(s.orgID)+"/agenten", payload, http.StatusCreated)
	if err != nil {
		return err
	}
	created = struct{ ID, Name string }{}
	if err := json.Unmarshal(body, &created); err != nil {
		return err
	}
	if created.ID == "" || created.Name != agent {
		return fmt.Errorf("Eino-Agent nicht angelegt: %s", body)
	}
	s.agentID = created.ID
	s.agentName = agent
	return nil
}

func (s *suite) grantsPath() string {
	return "/api/organisationen/" + url.PathEscape(s.orgID) + "/modellfreigabe/openrouter"
}

func (s *suite) grantBoth(org, agent string) error {
	if org != s.orgName || agent != s.agentName || s.orgID == "" || s.agentID == "" {
		return fmt.Errorf("falsche Freigabeziele")
	}
	if _, err := s.expect(http.MethodPost, s.grantsPath()+"/organisation", nil, http.StatusOK); err != nil {
		return err
	}
	_, err := s.expect(http.MethodPost, s.grantsPath()+"/agenten/"+url.PathEscape(s.agentID), nil, http.StatusOK)
	return err
}

func (s *suite) grantsVisible() error {
	body, err := s.expect(http.MethodGet, s.grantsPath(), nil, http.StatusOK)
	if err != nil {
		return err
	}
	var overview struct {
		Status struct {
			Reference           string `json:"reference"`
			OrganizationGranted bool   `json:"organization_granted"`
		} `json:"status"`
		Agents []struct {
			ID     string `json:"id"`
			Status struct {
				Reference           string `json:"reference"`
				OrganizationGranted bool   `json:"organization_granted"`
				AgentGranted        bool   `json:"agent_granted"`
				Allowed             bool   `json:"allowed"`
			} `json:"status"`
		} `json:"agents"`
	}
	if err := json.Unmarshal(body, &overview); err != nil {
		return err
	}
	if overview.Status.Reference != s.reference || !overview.Status.OrganizationGranted {
		return fmt.Errorf("Organisationsfreigabe fehlt: %s", body)
	}
	for _, a := range overview.Agents {
		if a.ID == s.agentID && a.Status.Reference == s.reference && a.Status.OrganizationGranted && a.Status.AgentGranted && a.Status.Allowed {
			return nil
		}
	}
	return fmt.Errorf("Agentenfreigabe fehlt: %s", body)
}

func (s *suite) redirectUntouched() error {
	_, second, proxy := s.calls()
	if len(second) != 0 || len(proxy) != 0 {
		return fmt.Errorf("Redirect-Empfänger=%d Sperrproxy=%d Anfragen", len(second), len(proxy))
	}
	return nil
}

func (s *suite) noProviderRequest() error {
	first, second, proxy := s.calls()
	if len(first)+len(second)+len(proxy) != 0 {
		return fmt.Errorf("Anbieteranfrage vor Start: %d/%d/%d", len(first), len(second), len(proxy))
	}
	return nil
}

func (s *suite) logNoSecret() error {
	if bytes.Contains(s.logs(), []byte(syntheticKey)) {
		return fmt.Errorf("synthetischer Schlüssel in Prozessausgabe")
	}
	return nil
}

func (s *suite) noSecret() error {
	for _, data := range s.responseBodies {
		if bytes.Contains(data, []byte(syntheticKey)) {
			return fmt.Errorf("synthetischer Schlüssel in öffentlicher Antwort")
		}
	}
	return s.logNoSecret()
}
