// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package registry

// Package represents the structure of a package.yml file
type Package struct {
	APIVersion    string          `yaml:"apiVersion"`
	Kind          string          `yaml:"kind"`
	Metadata      Metadata        `yaml:"metadata"`
	Compatibility []Compatibility `yaml:"compatibility"`
	Delivery      Delivery        `yaml:"delivery"`
	Dependencies  Dependencies    `yaml:"dependencies"`
}

// Metadata holds core information about the package
type Metadata struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	AppVersion  []string `yaml:"appVersion"`
	Description string   `yaml:"description"`
	Authors     []string `yaml:"authors"`
	Homepage    string   `yaml:"homepage"`
	License     string   `yaml:"license"`
	Tags        []string `yaml:"tags"`
}

// Compatibility specifies supported OS and architectures
type Compatibility struct {
	OS   string   `yaml:"os"`
	Arch []string `yaml:"arch"`
}

// Delivery dictates how the package is delivered/installed
type Delivery struct {
	Type     string   `yaml:"type"`
	Binaries []string `yaml:"binaries"`
}

// Dependencies contains required and recommended package dependencies
type Dependencies struct {
	Required    map[string]string `yaml:"required"`
	Recommended map[string]string `yaml:"recommended"`
}
