package codexcli

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (c Config) verify() error {
	if !c.validScenario() || c.DatabasePath == "" {
		return errors.New("probe configuration missing")
	}

	for _, binary := range [][2]string{{c.CLIPath, c.CLISHA256}, {c.BwrapPath, c.BwrapSHA256}, {c.ActionServerPath, c.ActionServerSHA256}} {
		if err := c.verifyBinary(binary[0], binary[1]); err != nil {
			return err
		}
	}

	return nil
}

func (c Config) validScenario() bool {
	for _, name := range []string{"action_allowed", "write_denied", "shell_denied", "http_denied", "foreign_org", "foreign_agent", "traversal", "symlink"} {
		if c.Scenario == name {
			return true
		}
	}

	return false
}

func (c Config) verifyBinary(path, expected string) error {
	if !filepath.IsAbs(path) || len(expected) != 64 {
		return errors.New("binary pin missing")
	}

	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("binary is not a regular file")
	}

	return c.verifyDigest(path, expected)
}

func (c Config) verifyDigest(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expected) {
		return errors.New("binary digest mismatch")
	}

	return nil
}

func (c Config) toml() string {
	return fmt.Sprint("model = \"synthetic-model\"\nmodel_provider = \"synthetic\"\n",
		"approval_policy = \"never\"\nsandbox_mode = \"read-only\"\nweb_search = \"disabled\"\n",
		"[model_providers.synthetic]\nname = \"Local synthetic proof\"\n",
		"base_url = \"http://127.0.0.1:18761/v1\"\nwire_api = \"responses\"\n",
		"requires_openai_auth = false\nrequest_max_retries = 0\nstream_max_retries = 0\n",
		"[features]\nshell_tool = false\ngoals = false\napps = false\nplugins = false\n",
		"browser_use = false\ncomputer_use = false\nhooks = false\nmulti_agent = false\n",
		"view_image = false\nimage_generation = false\nskill_mcp_dependency_install = false\n",
		"[mcp_servers.fach]\ncommand = \"/opt/codex-action-server\"\n",
		"enabled_tools = [\"artifact.markdown.save\"]\n",
		"[mcp_servers.fach.env]\nACP_CODEX_ACTION_SOCKET = \"/run/acp/action.sock\"\n",
		"[mcp_servers.fach.tools.\"artifact.markdown.save\"]\napproval_mode = \"approve\"\n")
}
