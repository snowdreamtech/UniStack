// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/spf13/cobra"
	"github.com/ulikunitz/xz"
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

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	fmt.Printf("Checking for updates to %s...\n", env.ProjectName)

	apiURL := "https://api.github.com/repos/snowdreamtech/UniGo/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var release struct {
		TagName string         `json:"tag_name"`
		Assets  []releaseAsset `json:"assets"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")
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

	for _, asset := range release.Assets {
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

	var binaryData []byte
	var decompressed io.Reader

	// Intelligent extraction based on magic bytes
	if len(archiveData) > 2 && bytes.HasPrefix(archiveData, []byte{0x1f, 0x8b}) {
		// Gzip
		gzr, err := gzip.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzr.Close()
		decompressed = gzr
	} else if len(archiveData) > 3 && bytes.HasPrefix(archiveData, []byte("BZh")) {
		// Bzip2
		decompressed = bzip2.NewReader(bytes.NewReader(archiveData))
	} else if len(archiveData) > 5 && bytes.HasPrefix(archiveData, []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}) {
		// XZ
		xzr, err := xz.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return fmt.Errorf("failed to create xz reader: %w", err)
		}
		decompressed = xzr
	} else if len(archiveData) > 3 && bytes.HasPrefix(archiveData, []byte{0x28, 0xb5, 0x2f, 0xfd}) {
		// Zstd
		zstdr, err := zstd.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return fmt.Errorf("failed to create zstd reader: %w", err)
		}
		defer zstdr.Close()
		decompressed = zstdr
	} else if len(archiveData) > 3 && bytes.HasPrefix(archiveData, []byte{0x04, 0x22, 0x4d, 0x18}) {
		// LZ4
		decompressed = lz4.NewReader(bytes.NewReader(archiveData))
	}

	if decompressed != nil {
		tr := tar.NewReader(decompressed)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read tar archive: %w", err)
			}
			if filepath.Base(hdr.Name) == binaryName && !hdr.FileInfo().IsDir() {
				binaryData, err = io.ReadAll(tr)
				if err != nil {
					return fmt.Errorf("failed to read binary from tar: %w", err)
				}
				break
			}
		}
	} else if len(archiveData) > 4 && bytes.HasPrefix(archiveData, []byte{0x50, 0x4b, 0x03, 0x04}) {
		// Zip
		zr, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
		if err != nil {
			return fmt.Errorf("failed to create zip reader: %w", err)
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == binaryName && !f.FileInfo().IsDir() {
				rc, err := f.Open()
				if err != nil {
					return fmt.Errorf("failed to open file in zip: %w", err)
				}
				binaryData, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					return fmt.Errorf("failed to read binary from zip: %w", err)
				}
				break
			}
		}
	} else {
		// Try treating it as raw binary data just in case
		binaryData = archiveData
	}

	if len(binaryData) == 0 {
		return fmt.Errorf("failed to find %s inside the downloaded archive", binaryName)
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

	fmt.Printf("Successfully updated to version %s!\n", latestVer)
	return nil
}
