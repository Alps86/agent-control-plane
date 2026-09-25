package aktivitaet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cucumber/godog"
)

func (s *Suite) registerBrowser(sc *godog.ScenarioContext) {
	sc.Step(`^ich die Organisation "([^"]+)" im Browser öffne$`, s.openOrganization)
	sc.Step(`^ich von dort den Aktivitätsverlauf öffne$`, s.followActivity)
	sc.Step(`^sehe ich getrennt die Anlage und die Zuweisung von "([^"]+)"$`, s.twoBrowserEvents)
	sc.Step(`^ich sehe bei beiden Ereignissen Akteur, Quelle und Zeitpunkt$`, s.browserMetadata)
	sc.Step(`^ich den Deep Link eines Ereignisses öffne$`, s.followTask)
	sc.Step(`^sehe ich das Aufgabendetail von "([^"]+)"$`, s.taskInBrowser)
	sc.Step(`^ich den Aktivitätsverlauf von "([^"]+)" im Browser öffne$`, s.openActivity)
	sc.Step(`^sehe ich einen verständlichen Leerzustand ohne Aufgabenereignis$`, s.emptyBrowserActivity)
}

func (s *Suite) startBrowser() error {
	if s.browser != nil {
		return nil
	}

	script, err := filepath.Abs("browser.mjs")
	if err != nil {
		return err
	}

	s.browser = exec.Command("node", script)
	s.browser.Stderr = os.Stderr
	s.browser.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return s.browserPipes()
}

func (s *Suite) browserPipes() error {
	input, err := s.browser.StdinPipe()
	if err != nil {
		return err
	}

	output, err := s.browser.StdoutPipe()
	if err != nil {
		return err
	}

	s.input, s.output = bufio.NewWriter(input), bufio.NewScanner(output)
	return s.browser.Start()
}

func (s *Suite) stopBrowser() {
	if s.browser == nil || s.browser.Process == nil {
		return
	}

	_ = syscall.Kill(-s.browser.Process.Pid, syscall.SIGKILL)
	_ = s.browser.Wait()
	s.browser = nil
}

func (s *Suite) browserCommand(command map[string]string) error {
	data, _ := json.Marshal(command)
	if _, err := s.input.Write(append(data, '\n')); err != nil {
		return err
	}

	if err := s.input.Flush(); err != nil {
		return err
	}

	return s.browserReply()
}

func (s *Suite) browserReply() error {
	if !s.output.Scan() {
		return fmt.Errorf("Chrome endete: %v", s.output.Err())
	}

	var reply BrowserReply
	if err := json.Unmarshal(s.output.Bytes(), &reply); err != nil {
		return err
	}

	if !reply.OK {
		return fmt.Errorf("Chrome: %s", reply.Error)
	}

	s.page = reply.Page
	return nil
}

func (s *Suite) browserNavigate(path, mode string) error {
	return s.browserCommand(map[string]string{"url": s.url(path), "mode": mode})
}

func (s *Suite) openOrganization(org string) error {
	if err := s.startBrowser(); err != nil {
		return err
	}

	return s.browserNavigate("/organisationen/"+s.organizations[org], "organization")
}

func (s *Suite) followActivity() error {
	path := "/organisationen/" + s.organizations["Nordstern"] + "/aktivitaet"
	return s.browserCommand(map[string]string{"click": `main a[href="` + path + `"]`, "mode": "activity"})
}

func (s *Suite) openActivity(org string) error {
	if err := s.startBrowser(); err != nil {
		return err
	}

	return s.browserNavigate("/organisationen/"+s.organizations[org]+"/aktivitaet", "activity")
}

func (s *Suite) twoBrowserEvents(title string) error {
	if strings.Count(s.page.Text, title) < 2 || !strings.Contains(s.page.Text, "angelegt") || !strings.Contains(s.page.Text, "zugewiesen") {
		return fmt.Errorf("Anlage/Zuweisung fehlen: %+v", s.page)
	}

	return nil
}

func (s *Suite) browserMetadata() error {
	for _, value := range []string{"Betreiber", "API"} {
		if !strings.Contains(s.page.Text, value) {
			return fmt.Errorf("Metadatum %q fehlt: %+v", value, s.page)
		}
	}

	if len(s.page.Times) != 2 {
		return fmt.Errorf("Zwei Zeitpunkte fehlen: %+v", s.page)
	}

	for _, value := range s.page.Times {
		if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
			return fmt.Errorf("Zeitpunkt %q ungültig: %w", value, err)
		}
	}

	return nil
}

func (s *Suite) followTask() error {
	path := "/organisationen/" + s.organizations["Nordstern"] + "/projekte/" + s.projects[s.key("Nordstern", "Website")] + "/aufgaben/" + s.taskID
	return s.browserCommand(map[string]string{"click": `main a[href="` + path + `"]`, "mode": "task"})
}

func (s *Suite) taskInBrowser(title string) error {
	if s.page.Heading != title || !strings.Contains(s.page.URL, s.taskID) {
		return fmt.Errorf("Aufgabendetail fehlt: %+v", s.page)
	}

	return nil
}

func (s *Suite) emptyBrowserActivity() error {
	if !strings.Contains(s.page.Text, "keine Aktivität") {
		return fmt.Errorf("Leerzustand fehlt: %+v", s.page)
	}

	return nil
}
