// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package types

// Scenario represents a complete stack of applications to be deployed.
type Scenario struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Apps        []App  `yaml:"apps"`
}

// App represents a single application component within a Scenario.
type App struct {
	Name      string                 `yaml:"name"`
	DependsOn []string               `yaml:"depends_on,omitempty"`
	Vars      map[string]interface{} `yaml:"vars,omitempty"`
}

// ExecResult represents the result of an engine execution.
type ExecResult struct {
	Success bool
	Output  string
	Error   error
}
