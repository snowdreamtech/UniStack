// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/snowdreamtech/unigo/internal/pkg/version"
)

const (
	githubAPIURL = "https://api.github.com/repos/snowdreamtech/UniGo/releases/latest"
	cacheFile    = "update-cache.json"
	checkPeriod  = 24 * time.Hour
	promptPeriod = 24 * time.Hour
)

// UpdateCache stores the cache for latest version check.
type UpdateCache struct {
	LatestVersion string    `json:"latest_version"`
	LastChecked   time.Time `json:"last_checked"`
	LastPrompted  time.Time `json:"last_prompted"`
}

var (
	// commandBlacklist contains commands that should never show an update prompt.
	commandBlacklist = map[string]bool{
		"env":         true,
		"completion":  true,
		"version":     true,
		"self-update": true,
		"__complete":  true, // Cobra completion
	}
	cacheMutex sync.Mutex
)

// getCachePath returns the path to the update cache file.
func getCachePath() string {
	return filepath.Join(env.GetDataDir(), cacheFile)
}

// readCache reads the update cache from disk.
func readCache() (*UpdateCache, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	data, err := os.ReadFile(getCachePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &UpdateCache{}, nil
		}
		return nil, err
	}

	var cache UpdateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return &UpdateCache{}, nil // Ignore corrupt cache
	}
	return &cache, nil
}

// writeCache writes the update cache to disk.
func writeCache(cache *UpdateCache) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	dir := filepath.Dir(getCachePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(getCachePath(), data, 0644)
}

// ClearCache clears the update cache from disk.
func ClearCache() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	return os.Remove(getCachePath())
}

// ReleaseAsset represents an asset in a GitHub release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseInfo represents the GitHub release information.
type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

// FetchLatestReleaseInfo fetches the latest release information from GitHub API.
func FetchLatestReleaseInfo(ctx context.Context) (*ReleaseInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Relies on default transport which respects HTTP_PROXY/HTTPS_PROXY env vars.
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release ReleaseInfo
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// CheckUpdateAsync asynchronously checks for an update if 24 hours have passed since the last check.
func CheckUpdateAsync(currentVersion string) {
	if env.Silent || env.Quiet {
		return
	}

	if currentVersion == "N/A" || currentVersion == "dev" || currentVersion == "" {
		return
	}

	go func() {
		cache, err := readCache()
		if err != nil {
			return // Ignore errors in async routine
		}

		// Check if we need to fetch a new version
		if time.Since(cache.LastChecked) < checkPeriod {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		releaseInfo, err := FetchLatestReleaseInfo(ctx)
		if err != nil {
			// On error, just update the check time to avoid spamming the API
			cache.LastChecked = time.Now()
			_ = writeCache(cache)
			return
		}

		cache.LatestVersion = strings.TrimPrefix(releaseInfo.TagName, "v")
		cache.LastChecked = time.Now()
		_ = writeCache(cache)
	}()
}

// PromptIfAvailable prints an update prompt if a newer version is available and constraints are met.
func PromptIfAvailable(currentVersion string, cmdName string) {
	if env.Silent || env.Quiet {
		return
	}

	if commandBlacklist[cmdName] {
		return
	}

	if currentVersion == "N/A" || currentVersion == "dev" || currentVersion == "" {
		return
	}

	// Only prompt if we are in an interactive terminal to prevent breaking scripts/pipes.
	if !isatty.IsTerminal(os.Stderr.Fd()) {
		return
	}

	cache, err := readCache()
	if err != nil || cache.LatestVersion == "" {
		return
	}

	// Clean up versions for comparison
	curVer := strings.TrimPrefix(currentVersion, "v")
	latestVer := strings.TrimPrefix(cache.LatestVersion, "v")

	// Avoid prompting if it's not a valid released version or current >= latest
	if version.CompareVersions(latestVer, curVer) <= 0 {
		return
	}

	// Avoid prompting more than once per day
	if time.Since(cache.LastPrompted) < promptPeriod {
		return
	}

	// Print prompt to Stderr
	pterm.Warning.Printf("unigo version %s available\n", cache.LatestVersion)
	pterm.Warning.Printf("To update, run `unigo self-update`\n")

	// Update prompt time
	cache.LastPrompted = time.Now()
	_ = writeCache(cache)
}
