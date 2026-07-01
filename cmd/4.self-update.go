// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unigo/internal/pkg/archive"
	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/snowdreamtech/unigo/internal/updater"
	"github.com/spf13/cobra"
)

var skipChecksum bool

func init() {
	selfUpdateCmd.Flags().BoolVar(&skipChecksum, "skip-checksum", false, "Skip checksum verification")
	rootCmd.AddCommand(selfUpdateCmd)
}

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update to the latest version",
	Long:  "Update the application to the latest available version from GitHub Releases.",
	RunE:  runSelfUpdate,
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	fmt.Printf("Checking for updates to %s...\n", env.ProjectName)

	releaseInfo, err := updater.FetchLatestReleaseInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVer := strings.TrimPrefix(releaseInfo.TagName, "v")
	currentVer := strings.TrimPrefix(env.GitTag, "v")

	fmt.Printf("Current version: %s\n", currentVer)
	fmt.Printf("Latest version:  %s\n", latestVer)

	if latestVer == currentVer || latestVer == "" {
		fmt.Println("Already up to date.")
		return nil
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var downloadURL string
	var assetName string
	var checksumsURL string

	for _, asset := range releaseInfo.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, "checksums.txt") {
			checksumsURL = asset.BrowserDownloadURL
			continue
		}
		if strings.Contains(name, goos) && strings.Contains(name, goarch) {
			if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") ||
				strings.HasSuffix(name, ".tar.xz") || strings.HasSuffix(name, ".tar.zst") ||
				strings.HasSuffix(name, ".tar.bz2") || strings.HasSuffix(name, ".tar.lz4") {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
			}
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s", goos, goarch)
	}

	var expectedHash string
	if !skipChecksum {
		if checksumsURL == "" {
			return fmt.Errorf("checksums.txt not found in release assets")
		}
		fmt.Printf("Downloading checksums...\n")
		chkReq, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumsURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create checksum request: %w", err)
		}
		chkResp, err := http.DefaultClient.Do(chkReq)
		if err != nil {
			return fmt.Errorf("failed to download checksums: %w", err)
		}
		defer chkResp.Body.Close()

		chkBody, err := io.ReadAll(chkResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read checksums: %w", err)
		}

		// Find the hash for our assetName
		lines := strings.Split(string(chkBody), "\n")
		for _, line := range lines {
			if strings.Contains(line, assetName) {
				parts := strings.Fields(line)
				if len(parts) >= 2 && parts[1] == assetName {
					expectedHash = parts[0]
					break
				}
			}
		}
		if expectedHash == "" {
			return fmt.Errorf("hash for %s not found in checksums.txt", assetName)
		}
	} else {
		fmt.Printf("Skipping checksum verification due to --skip-checksum flag.\n")
	}

	fmt.Printf("Downloading %s...\n", downloadURL)
	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	dlResp, err := http.DefaultClient.Do(dlReq)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer dlResp.Body.Close()

	archiveData, err := io.ReadAll(dlResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read downloaded archive: %w", err)
	}

	if !skipChecksum {
		fmt.Printf("Verifying checksum...\n")
		hasher := sha256.New()
		hasher.Write(archiveData)
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if actualHash != expectedHash {
			return fmt.Errorf("checksum verification failed: expected %s, got %s", expectedHash, actualHash)
		}
		fmt.Printf("Checksum verified successfully.\n")
	}

	fmt.Printf("Extracting binary...\n")
	binaryName := "unigo"
	if goos == "windows" {
		binaryName = "unigo.exe"
	}

	binaryData, err := archive.ExtractBinary(archiveData, binaryName)
	if err != nil {
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlink: %w", err)
	}

	tmpPath := execPath + ".new"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := tmpFile.Write(binaryData); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write binary: %w", err)
	}
	tmpFile.Close()

	// Atomically replace the current binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Clear update cache after successful update
	_ = updater.ClearCache()

	fmt.Printf("Successfully updated to version %s!\n", latestVer)
	return nil
}
