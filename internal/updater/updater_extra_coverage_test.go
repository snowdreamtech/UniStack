// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/snowdreamtech/unistack/internal/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type customTransport struct {
	rt     http.RoundTripper
	server *httptest.Server
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect all requests to our test server
	req.URL.Scheme = "http"
	req.URL.Host = t.server.Listener.Addr().String()
	return t.rt.RoundTrip(req)
}

func TestFetchLatestReleaseInfo_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/snowdreamtech/UniStack/releases/latest", r.URL.Path)
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))

		info := ReleaseInfo{
			TagName: "v1.2.3",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}))
	defer ts.Close()

	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = &customTransport{
		rt:     http.DefaultTransport,
		server: ts,
	}

	ctx := context.Background()
	info, err := FetchLatestReleaseInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", info.TagName)
}

func TestFetchLatestReleaseInfo_HttpError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = &customTransport{
		rt:     http.DefaultTransport,
		server: ts,
	}

	ctx := context.Background()
	_, err := FetchLatestReleaseInfo(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func TestFetchLatestReleaseInfo_JSONError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = &customTransport{
		rt:     http.DefaultTransport,
		server: ts,
	}

	ctx := context.Background()
	_, err := FetchLatestReleaseInfo(ctx)
	assert.Error(t, err)
}

func TestFetchLatestReleaseInfo_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Sleep long enough to trigger timeout
	}))
	defer ts.Close()

	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = &customTransport{
		rt:     http.DefaultTransport,
		server: ts,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := FetchLatestReleaseInfo(ctx)
	assert.Error(t, err)
}

func TestClearCache(t *testing.T) {
	// write something to cache first
	cache := UpdateCache{LatestVersion: "v1.2.3"}
	writeCache(&cache)

	err := ClearCache()
	assert.NoError(t, err)

	cache2, err := readCache()
	assert.NoError(t, err) // readCache returns empty struct, not error, when missing
	assert.Equal(t, "", cache2.LatestVersion)
}

func TestCheckUpdateAsync(t *testing.T) {
	// Make silent to return early
	env.Silent = true
	CheckUpdateAsync("1.0.0")
	env.Silent = false

	// Invalid version
	CheckUpdateAsync("dev")
	CheckUpdateAsync("N/A")
	CheckUpdateAsync("")

	// Clear cache to simulate first check
	ClearCache()

	// Fast track CheckUpdateAsync by mocking the server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := ReleaseInfo{TagName: "v2.0.0"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}))
	defer ts.Close()

	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = &customTransport{
		rt:     http.DefaultTransport,
		server: ts,
	}

	CheckUpdateAsync("1.0.0")
	// wait for async to finish
	time.Sleep(50 * time.Millisecond)

	cache, err := readCache()
	assert.NoError(t, err)
	assert.Equal(t, "2.0.0", cache.LatestVersion)
}

func TestPromptIfAvailable(t *testing.T) {
	// Test silent return
	env.Silent = true
	PromptIfAvailable("1.0.0", "some-cmd")
	env.Silent = false

	// Test blacklist
	PromptIfAvailable("1.0.0", "version")

	// Test invalid versions
	PromptIfAvailable("N/A", "test")
	PromptIfAvailable("dev", "test")

	// We can't easily test the rest of PromptIfAvailable without mocking isatty.IsTerminal
	// But we can cover the earlier branches.
}
