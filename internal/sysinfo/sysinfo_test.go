// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sysinfo

import (
	"runtime"
	"testing"
)

func TestIsMusl(t *testing.T) {
	// The exact result depends on the environment running the test,
	// but we can at least ensure it doesn't panic and returns false
	// on non-linux systems.
	result := IsMusl()
	if runtime.GOOS != "linux" && result {
		t.Errorf("IsMusl() should be false on %s", runtime.GOOS)
	}
}
