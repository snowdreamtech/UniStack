package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/snowdreamtech/unistack/internal/types"
)

// AnsibleRunner implements EngineRunner using the local ansible-playbook binary.
type AnsibleRunner struct {
	PlaybookPath string
	Inventory    string // e.g., "localhost,"
}

// NewAnsibleRunner creates a new AnsibleRunner instance.
func NewAnsibleRunner(playbookPath string) *AnsibleRunner {
	return &AnsibleRunner{
		PlaybookPath: playbookPath,
		Inventory:    "localhost,",
	}
}

// RunApp executes the Ansible playbook for the specific app, passing variables via extra-vars.
func (a *AnsibleRunner) RunApp(ctx context.Context, appName string, vars map[string]interface{}) (*types.ExecResult, error) {
	// Ensure the appName is injected into the vars
	if vars == nil {
		vars = make(map[string]interface{})
	}
	vars["app_name"] = appName

	// Serialize vars to JSON for Ansible extra-vars
	extraVarsJSON, err := json.Marshal(vars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extra vars to JSON: %w", err)
	}

	cmd := exec.CommandContext(ctx, "ansible-playbook",
		"-i", a.Inventory,
		"-c", "local",
		"-e", string(extraVarsJSON),
		a.PlaybookPath,
	)

	// Inject ANSIBLE_STDOUT_CALLBACK=json to force structured output
	// This makes the command execution robust against textual changes in Ansible output.
	cmd.Env = append(cmd.Environ(), "ANSIBLE_STDOUT_CALLBACK=json")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	success := err == nil

	result := &types.ExecResult{
		Success: success,
		Output:  stdoutBuf.String(),
	}

	if !success {
		result.Error = fmt.Errorf("ansible execution failed: %v. Stderr: %s", err, stderrBuf.String())
	}

	return result, nil
}
