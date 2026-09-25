package modelllebenszyklus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"

	modelopenrouter "agentcontrolplane/app/internal/adapter/model/openrouter"
	"agentcontrolplane/app/internal/adapter/sqlite"
	webagent "agentcontrolplane/app/internal/adapter/web/agent"
	webgrant "agentcontrolplane/app/internal/adapter/web/modellfreigabe"
	webopenrouter "agentcontrolplane/app/internal/adapter/web/openrouter"
	weborg "agentcontrolplane/app/internal/adapter/web/organisation"
	appagent "agentcontrolplane/app/internal/app/agent"
	appgrant "agentcontrolplane/app/internal/app/modellfreigabe"
	"agentcontrolplane/app/internal/app/openrouterverbindung"
	apporg "agentcontrolplane/app/internal/app/organisation"
	domainfreigabe "agentcontrolplane/app/internal/domain/modellfreigabe"
	grantport "agentcontrolplane/app/internal/port/modellfreigabe"
	"agentcontrolplane/ui/bridge"
	"github.com/cucumber/godog"
)

func (s *Suite) registerInvoke(sc *godog.ScenarioContext) {
	sc.Step(`^eine geprüfte OpenRouter-Verbindung ist für eine Organisation und ihren Eino-Agenten freigegeben$`, s.invokeSetup)
	sc.Step(`^dieser Agent einen neuen Modellzugangsaufruf über die öffentliche Anwendungsgrenze startet$`, s.invokeModel)
	sc.Step(`^erhält der kontrollierte Modellanbieter genau einen neuen Aufruf$`, s.oneModelCall)
	sc.Step(`^ich die Verbindung über die öffentliche Settings-Grenze trenne$`, s.disconnectInvoke)
	sc.Step(`^derselbe Agent erneut einen Modellzugangsaufruf startet$`, s.invokeModel)
	sc.Step(`^wird der neue Aufruf vor dem Anbieter wegen fehlender Verbindung abgelehnt$`, s.invokeDenied)
	sc.Step(`^der kontrollierte Modellanbieter erhält keinen weiteren Aufruf$`, s.noAdditionalModelCall)
	sc.Step(`^ich dieselbe zentrale Verbindung mit einem neuen Schlüssel ohne neue Freigaben einrichte$`, s.reconnectInvoke)
	sc.Step(`^bleiben die alten Freigaben unwirksam und der Anbieter erhält keinen weiteren Aufruf$`, s.noOldGrant)
}

func (s *Suite) invokeSetup() error {
	if err := s.setup(); err != nil {
		return err
	}
	if err := s.launchInvoke(); err != nil {
		return err
	}
	if err := s.routerAction(http.MethodPost, routerPath, boundKey); err != nil {
		return err
	}
	if err := s.checkRouter(); err != nil {
		return err
	}
	if err := s.createInvokeEntities(); err != nil {
		return err
	}
	return s.grantInvoke()
}

func (s *Suite) launchInvoke() error {
	s.app.Close()
	db, err := sqlite.OpenApplication(context.Background(), filepath.Join(s.dir, "invoke.sqlite"))
	if err != nil {
		return err
	}
	s.db = db
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	handler, err := s.invokeHandler(listener.Addr().String())
	if err != nil {
		listener.Close()
		return err
	}
	s.app = httptest.NewUnstartedServer(handler)
	s.app.Listener.Close()
	s.app.Listener = listener
	s.app.Start()
	return nil
}

func (s *Suite) invokeHandler(address string) (http.Handler, error) {
	ui, err := bridge.New()
	if err != nil {
		return nil, err
	}
	identity := apporg.NewLocalIdentity()
	organizations := sqlite.NewOrganizationStore(s.db)
	orgHandler := weborg.NewHandler(apporg.NewService(organizations, identity), ui, address)
	agentHandler := webagent.NewHandler(appagent.NewService(s.db, identity, organizations), ui, address)
	probe := modelopenrouter.NewProbeAt(nil, s.provider.URL+"/api/v1/key")
	s.routerService = openrouterverbindung.NewService(s.store, probe)
	settingsHandler := webopenrouter.NewHandler(s.routerService, ui, address)
	s.grantService = appgrant.NewService(sqlite.NewModellfreigabeStore(s.db), organizations, s.db, identity, s.routerService)
	grantHandler := webgrant.NewHandler(s.grantService, ui, address)
	return s.invokeRoutes(orgHandler, agentHandler, settingsHandler, grantHandler), nil
}

func (s *Suite) invokeRoutes(org, agent, settings, grant http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/modellfreigabe/openrouter") {
			grant.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/settings/modellanbieter/openrouter") {
			settings.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/agenten") {
			agent.ServeHTTP(w, r)
			return
		}
		org.ServeHTTP(w, r)
	})
}

func (s *Suite) requestJSON(method, path string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(method, s.app.URL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header.Set("Origin", s.app.URL)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.app.Client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	s.status = response.StatusCode
	s.response, err = io.ReadAll(io.LimitReader(response.Body, 1<<20))
	return err
}

func (s *Suite) createInvokeEntities() error {
	if err := s.requestJSON(http.MethodPost, "/api/organisationen", map[string]string{"name": "Testorganisation"}); err != nil {
		return err
	}
	if err := s.createdID(&s.organizationID); err != nil {
		return err
	}
	path := "/api/organisationen/" + s.organizationID + "/agenten"
	if err := s.requestJSON(http.MethodPost, path, map[string]string{"name": "Testagent", "template_id": "recherche", "execution_kind": "eino"}); err != nil {
		return err
	}
	return s.createdID(&s.agentID)
}

func (s *Suite) createdID(target *string) error {
	if s.status != http.StatusCreated {
		return fmt.Errorf("Anlage HTTP %d: %s", s.status, s.response)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(s.response, &result); err != nil {
		return err
	}
	if result.ID == "" {
		return fmt.Errorf("öffentliche Anlage lieferte keine ID")
	}
	*target = result.ID
	return nil
}

func (s *Suite) grantInvoke() error {
	base := "/api/organisationen/" + s.organizationID + "/modellfreigabe/openrouter"
	if err := s.requestJSON(http.MethodPost, base+"/organisation", nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Organisationsfreigabe HTTP %d: %s", s.status, s.response)
	}
	if err := s.requestJSON(http.MethodPost, base+"/agenten/"+s.agentID, nil); err != nil {
		return err
	}
	if s.status != http.StatusOK {
		return fmt.Errorf("Agentenfreigabe HTTP %d: %s", s.status, s.response)
	}
	return nil
}

func (s *Suite) invokeModel() error {
	s.invokeErr = s.grantService.Invoke(context.Background(), s.organizationID, s.agentID, openrouterverbindung.Reference, ControlledCaller{s})
	return nil
}

func (c ControlledCaller) Call(ctx context.Context, _, _ string, access grantport.Access) error {
	key, err := access.Resolve(ctx)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.suite.provider.URL+"/api/v1/chat/completions", strings.NewReader(`{"model":"synthetic","messages":[]}`))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+key)
	response, err := c.suite.provider.Client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("kontrollierter Anbieter HTTP %d", response.StatusCode)
	}
	return nil
}

func (s *Suite) oneModelCall() error {
	if s.invokeErr != nil || s.providerState.posts.Load() != 1 {
		return fmt.Errorf("Basisaufruf: %v, POSTs=%d", s.invokeErr, s.providerState.posts.Load())
	}
	return nil
}

func (s *Suite) disconnectInvoke() error { return s.routerAction(http.MethodDelete, routerPath, "") }

func (s *Suite) invokeDenied() error {
	if s.invokeErr == nil || s.invokeErr.Error() != domainfreigabe.ReasonConnection {
		return fmt.Errorf("getrennter Zugang: %v", s.invokeErr)
	}
	return s.noAdditionalModelCall()
}

func (s *Suite) noAdditionalModelCall() error {
	if count := s.providerState.posts.Load(); count != 1 {
		return fmt.Errorf("%d statt genau einem Anbieter-POST", count)
	}
	return nil
}

func (s *Suite) reconnectInvoke() error {
	if err := s.routerAction(http.MethodPost, routerPath, newKey); err != nil {
		return err
	}
	return s.checkRouter()
}

func (s *Suite) noOldGrant() error {
	if s.invokeErr == nil || s.invokeErr.Error() != domainfreigabe.ReasonOrganization {
		return fmt.Errorf("alte Freigaben nach Neuverbindung: %v", s.invokeErr)
	}
	return s.noAdditionalModelCall()
}
