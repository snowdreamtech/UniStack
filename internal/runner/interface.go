// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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
