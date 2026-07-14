// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/common-nighthawk/go-figure"
	"github.com/snowdreamtech/unigo/internal/env"
	"github.com/snowdreamtech/unigo/internal/sysinfo"
	"github.com/spf13/cobra"
)

// init registers the version command to the root command.
func init() {
	rootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command which displays the application version information.
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Print the version number of " + env.ProjectName,
	Long:    "Display version information including build details, commit hash, and build time.",
	Aliases: []string{"v"},
	Run:     runVersion,
}

// welcome prints a stylized title in green color.
func welcome() {
	title := figure.NewColorFigure(strings.ToUpper(env.ProjectName), "larry3d", "green", true)
	title.Print()
}

func runVersion(cmd *cobra.Command, args []string) {
	welcome()

	osArch := runtime.GOOS + "/" + runtime.GOARCH
	if sysinfo.IsMusl() {
		osArch += " (musl)"
	}
	buildVersion := fmt.Sprintf("%s version %s-%s %s\n", env.ProjectName, env.GitTag, env.CommitHash, osArch)
	copyrightDetail := fmt.Sprintf("%s\n", env.COPYRIGHT)
	licenseDetail := fmt.Sprintf("License: %s\n", env.LICENSE)
	authorDetail := fmt.Sprintf("Written by %s", env.Author)
	buildDetail := fmt.Sprintf("Built at %s", env.BuildTime)

	var builder strings.Builder
	builder.WriteString("\n")
	builder.WriteString(buildVersion)
	builder.WriteString(copyrightDetail)
	builder.WriteString(licenseDetail)
	builder.WriteString("\n")
	builder.WriteString(authorDetail)
	builder.WriteString("\n")
	builder.WriteString(buildDetail)

	fmt.Println(builder.String())
}
