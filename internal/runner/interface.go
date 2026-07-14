package runner

import (
	"context"

	"github.com/snowdreamtech/unistack/internal/types"
)

// EngineRunner defines the interface for underlying execution engines.
// This allows future engines (e.g., direct Kubernetes API, pure Shell) to be swapped in
// without modifying the orchestrator logic.
type EngineRunner interface {
	// RunApp triggers the deployment of a single application component.
	RunApp(ctx context.Context, appName string, vars map[string]interface{}) (*types.ExecResult, error)
}
