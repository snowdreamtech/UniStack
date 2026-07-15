// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"crypto/rand"
	"os"
)

// Get returns the value of the environment variable with the given key,
// searching with prefixes in order: UNISTACK_, MISE_, and then the raw key.
// Note: PATH is retrieved directly to avoid pollution from UNISTACK_PATH/MISE_PATH.
func Get(key string) string {
	if key == "PATH" {
		return os.Getenv("PATH")
	}
	// 1. UNISTACK_ prefix
	if v := os.Getenv("UNISTACK_" + key); v != "" {
		return v
	}
	// 2. MISE_ prefix
	if v := os.Getenv("MISE_" + key); v != "" {
		return v
	}
	// 3. Raw key (Native)
	return os.Getenv(key)
}

// GithubProxy returns the configured GitHub proxy URL or a stable public default.
func GithubProxy() string {
	if proxy := Get("GITHUB_PROXY"); proxy != "" {
		return proxy
	}
	return "https://gh-proxy.sn0wdr1am.com/"
}

var (
	//ProjectName Project Name
	ProjectName string = "unistack"

	//Author Author
	Author string = "Snowdream Tech <snowdreamtech@qq.com>"

	//BuildTime Build Time
	BuildTime string = "N/A"

	//GitTag Git Tag
	GitTag string = "N/A"

	//CommitHash Commit Hash
	CommitHash string = "N/A"

	//CommitHashFull Commit Hash
	CommitHashFull string = "N/A"

	//COPYRIGHT COPYRIGHT
	COPYRIGHT string = "Copyright (c) 2023-present SnowdreamTech Inc."

	//LICENSE LICENSE
	LICENSE string = "MIT <https://github.com/snowdreamtech/unistack/blob/main/LICENSE>"

	//Config Config File Path
	Config string = "unistack.toml"

	// Debug indicates whether the application should run in debug mode.
	Debug bool

	// Trace indicates whether the application should run in trace mode.
	Trace bool

	// Quiet indicates whether the application should run in quiet mode.
	Quiet bool

	// Cwd specifies the current working directory for the application.
	Cwd string

	// Silent indicates whether to suppress all output and non-error messages.
	Silent bool

	CryptoRandRead = rand.Read
)

// RandomString returns a random string of the specified length.
func RandomString(n int) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := make([]byte, n)
	if _, err := CryptoRandRead(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
