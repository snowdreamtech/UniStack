// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"testing"
)

func TestRunPreflightChecks(t *testing.T) {
	tempDir := t.TempDir()

	err := RunPreflightChecks(tempDir)
	if err != nil {
		t.Fatalf("Expected RunPreflightChecks to return nil, got: %v", err)
	}
}
