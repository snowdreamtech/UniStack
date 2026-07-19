// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/snowdreamtech/unistack/internal/env"
	"gopkg.in/yaml.v3"
)

// ensureAnsibleInstalled checks for ansible and installs it in a venv if missing
func ensureAnsibleInstalled(workDir string, pipIndexUrl string) (string, []string, error) {
	// First check if ansible-playbook is already in the system PATH
	sysBin, err := exec.LookPath("ansible-playbook")
	hasSystemAnsible := (err == nil)

	// Paths for local venv - placed OUTSIDE workDir so it survives atomic file extractions
	// when the UniStack binary is upgraded but python dependencies remain unchanged.
	venvDir := filepath.Join(env.GetDataDir(), ".ansible", "venv")

	venvBinDir := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		venvBinDir = filepath.Join(venvDir, "Scripts")
	}
	venvBin := filepath.Join(venvBinDir, "ansible-playbook.exe")
	if runtime.GOOS != "windows" {
		venvBin = filepath.Join(venvBinDir, "ansible-playbook")
	}
	markerFile := filepath.Join(env.GetDataDir(), ".ansible", ".bootstrap_complete")

	// Calculate dependency hash to detect version upgrades
	currentHash, _ := calculateDependenciesHash(workDir)

	var activeBin string
	var activeEnv []string

	if hasSystemAnsible {
		activeBin = sysBin
		activeEnv = nil
	} else {
		activeBin = venvBin
		activeEnv = buildVenvEnv(venvDir)
	}

	// If atomic marker exists, check hash and binary
	if markerData, err := os.ReadFile(markerFile); err == nil {
		if string(markerData) == currentHash {
			if hasSystemAnsible {
				slog.Debug(fmt.Sprintf("✅ Found system Ansible at %s\n", sysBin))
				return activeBin, activeEnv, nil
			} else if _, err := os.Stat(venvBin); err == nil {
				return activeBin, activeEnv, nil
			}
		} else {
			slog.Debug("🔄 Dependencies have changed (binary upgrade detected). Rebuilding environment...")
		}
	}

	// The global lock is now held by PrepareEnvironment, so we can proceed directly.

	// Double check marker after acquiring lock (not strictly needed now, but safe)
	if markerData, err := os.ReadFile(markerFile); err == nil {
		if string(markerData) == currentHash {
			if hasSystemAnsible {
				slog.Debug(fmt.Sprintf("✅ Found system Ansible at %s\n", sysBin))
				return activeBin, activeEnv, nil
			} else if _, err := os.Stat(venvBin); err == nil {
				return activeBin, activeEnv, nil
			}
		}
	}

	// Global context for all network operations (30 minute timeout), wrapped in a signal trap for Ctrl+C
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(sigCtx, 30*time.Minute)
	defer cancel()

	// Helper function for command retry with context
	runWithRetry := func(name string, createCmd func(context.Context) *exec.Cmd, maxRetries int, delay time.Duration) error {
		var lastErr error
		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				slog.Debug(fmt.Sprintf("⚠️ %s failed, retrying in %v (attempt %d/%d)...\n", name, delay, i+1, maxRetries))
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return fmt.Errorf("context timeout during %s", name)
				}
			}
			cmd := createCmd(ctx)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			lastErr = cmd.Run()
			if lastErr == nil {
				return nil
			}
		}
		return fmt.Errorf("%s failed after %d attempts: %w", name, maxRetries, lastErr)
	}

	if !hasSystemAnsible {
		// We are going to bootstrap. Remove incomplete venv if exists (Scorched Earth)
		os.RemoveAll(venvDir)
		os.Remove(markerFile)

		// Define robust execution closure with scorched earth on final failure
		bootstrapSuccess := false
		defer func() {
			if !bootstrapSuccess {
				slog.Debug("💥 Bootstrap failed or was interrupted. Executing scorched earth rollback...")
				os.RemoveAll(venvDir)
				os.Remove(markerFile)
			}
		}()

		// Delegate Python discovery and auto-installation to the python module
		pythonCmd, err := EnsurePythonInstalled(ctx)
		if err != nil {
			return "", nil, err
		}

		// Delegate venv creation to the python module
		venvEnv, err := SetupVirtualEnvironment(ctx, pythonCmd, venvDir)
		if err != nil {
			return "", nil, err
		}
		activeEnv = venvEnv // Update activeEnv with the new venv

		pipBin := filepath.Join(venvBinDir, "pip")

		if pipIndexUrl != "" {
			slog.Debug(fmt.Sprintf("📦 Configuring pip mirror: %s\n", pipIndexUrl))
			cmd := exec.CommandContext(ctx, pipBin, "config", "set", "global.index-url", pipIndexUrl)
			if err := cmd.Run(); err != nil {
				slog.Debug(fmt.Sprintf("⚠️ Warning: failed to set pip mirror: %v\n", err))
			}
		}

		// Install requirements via pip
		reqFile := filepath.Join(workDir, "requirements.txt")
		slog.Debug("📦 Installing Ansible dependencies via pip...")
		err = runWithRetry("pip install", func(c context.Context) *exec.Cmd {
			return exec.CommandContext(c, pipBin, "install", "-r", reqFile)
		}, 3, 3*time.Second)
		if err != nil {
			return "", nil, err
		}

		bootstrapSuccess = true // Pip and Venv setup succeeded
	} else {
		slog.Debug(fmt.Sprintf("✅ Found system Ansible at %s", sysBin))
		os.Remove(markerFile) // Invalidate marker while we install galaxy dependencies
	}

	// Install Ansible Galaxy Collections and Roles
	galaxyReqFile := filepath.Join(workDir, "requirements.yml")
	if _, err := os.Stat(galaxyReqFile); err == nil {
		slog.Debug("🌌 Installing Ansible Galaxy Dependencies (Collections & Roles)...")
		var galaxyBin string
		if !hasSystemAnsible {
			galaxyBin = filepath.Join(venvBinDir, "ansible-galaxy")
		} else {
			galaxyBin, err = exec.LookPath("ansible-galaxy")
			if err != nil {
				slog.Debug("⚠️ System ansible-galaxy not found in PATH, skipping galaxy installation")
				goto SKIP_GALAXY
			}
		}

		// 1. Try online installation first (up to 3 times)
		onlineErr := runWithRetry("ansible-galaxy collection install", func(c context.Context) *exec.Cmd {
			cCmd := exec.CommandContext(c, galaxyBin, "collection", "install", "-r", galaxyReqFile)
			cCmd.Dir = workDir
			envVars := activeEnv
			if envVars == nil {
				envVars = os.Environ()
			}
			envVars = append(envVars, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
			cCmd.Env = envVars
			return cCmd
		}, 3, 3*time.Second)

		// 2. Fallback to offline mechanism if online failed
		if onlineErr != nil {
			err := fallbackToOfflineCollections(ctx, galaxyReqFile, galaxyBin, activeEnv, workDir, runWithRetry)
			if err != nil {
				return "", nil, fmt.Errorf("fallback source build failed: %w (online error was: %v)", err, onlineErr)
			}
		}

		// Install Roles (ignore errors if no roles are defined in requirements.yml)
		_ = runWithRetry("ansible-galaxy role install", func(c context.Context) *exec.Cmd {
			cCmd := exec.CommandContext(c, galaxyBin, "role", "install", "-r", galaxyReqFile)
			cCmd.Dir = workDir
			envVars := activeEnv
			if envVars == nil {
				envVars = os.Environ()
			}
			envVars = append(envVars, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
			cCmd.Env = envVars
			return cCmd
		}, 3, 3*time.Second)
	}
SKIP_GALAXY:

	// Successfully finished everything. Write atomic marker with the hash.
	if file, err := os.OpenFile(markerFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); err == nil {
		file.WriteString(currentHash)
		file.Close()
	}
	os.Chmod(markerFile, 0600) // Immunity against umask stripping read permissions

	return activeBin, activeEnv, nil
}

// calculateDependenciesHash computes a SHA-256 hash of the content of requirements.txt and requirements.yml.
// This allows us to detect when the binary is upgraded and dependencies change, triggering a fresh bootstrap.
func calculateDependenciesHash(workDir string) (string, error) {
	hash := sha256.New()

	reqFile := filepath.Join(workDir, "requirements.txt")
	reqData, err := os.ReadFile(reqFile)
	if err == nil {
		hash.Write(reqData)
	}

	galaxyReqFile := filepath.Join(workDir, "requirements.yml")
	galaxyReqData, err := os.ReadFile(galaxyReqFile)
	if err == nil {
		hash.Write(galaxyReqData)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

type GalaxyRequirements struct {
	Collections []struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"collections"`
}

var galaxyRepoMap = map[string]string{
	"community.general": "ansible-collections/community.general",
	"ansible.posix":     "ansible-collections/ansible.posix",
	"containers.podman": "containers/ansible-podman-collections",
	"community.docker":  "ansible-collections/community.docker",
	"kubernetes.core":   "ansible-collections/kubernetes.core",
}

func fallbackToOfflineCollections(
	ctx context.Context,
	galaxyReqFile string,
	galaxyBin string,
	activeEnv []string,
	workDir string,
	runWithRetry func(string, func(context.Context) *exec.Cmd, int, time.Duration) error,
) error {
	slog.Debug("⚠️ Online collection installation failed. Automatically falling back to source build offline mechanism...")

	reqData, err := os.ReadFile(galaxyReqFile)
	if err != nil {
		return fmt.Errorf("failed to read requirements.yml: %w", err)
	}

	var reqs GalaxyRequirements
	if err := yaml.Unmarshal(reqData, &reqs); err != nil {
		return fmt.Errorf("failed to parse requirements.yml: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "unistack_source_*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for source collections: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var downloadTool string
	if _, err := exec.LookPath("curl"); err == nil {
		downloadTool = "curl"
	} else if _, err := exec.LookPath("wget"); err == nil {
		downloadTool = "wget"
	} else {
		return fmt.Errorf("neither curl nor wget found in system path, cannot download source")
	}

	proxy := os.Getenv("GITHUB_PROXY")
	var builtTarballs []string

	for _, col := range reqs.Collections {
		repo, ok := galaxyRepoMap[col.Name]
		if !ok {
			return fmt.Errorf("collection %s is not in the Repo Map, cannot perform source build fallback", col.Name)
		}

		slog.Debug(fmt.Sprintf("📦 Resolving source for %s version %s (repo: %s)...", col.Name, col.Version, repo))

		tagsToTry := []string{
			fmt.Sprintf("v%s", col.Version),
			col.Version,
		}

		var downloadedTarball string
		for _, tag := range tagsToTry {
			baseURL := fmt.Sprintf("https://github.com/%s/archive/refs/tags/%s.tar.gz", repo, tag)
			if proxy != "" {
				baseURL = proxy + baseURL
			}

			tarballPath := filepath.Join(tmpDir, fmt.Sprintf("%s-%s.tar.gz", col.Name, tag))

			// Try 2 times per tag in case of network flakes
			err = runWithRetry(fmt.Sprintf("download source %s (tag %s)", col.Name, tag), func(c context.Context) *exec.Cmd {
				if downloadTool == "curl" {
					return exec.CommandContext(c, "curl", "-L", "-s", "-f", "-o", tarballPath, baseURL)
				}
				return exec.CommandContext(c, "wget", "-q", "-O", tarballPath, baseURL)
			}, 2, 2*time.Second)

			if err == nil {
				downloadedTarball = tarballPath
				break
			}
		}

		if downloadedTarball == "" {
			return fmt.Errorf("failed to download source for %s: tags tried %v", col.Name, tagsToTry)
		}

		// Extract the downloaded source tarball
		extractDir := filepath.Join(tmpDir, fmt.Sprintf("extract-%s", col.Name))
		os.MkdirAll(extractDir, 0755)

		cmd := exec.CommandContext(ctx, "tar", "-xzf", downloadedTarball, "-C", extractDir, "--strip-components=1")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract source for %s: %w", col.Name, err)
		}

		// Build the collection
		slog.Debug(fmt.Sprintf("🔨 Building collection %s locally...", col.Name))
		err = runWithRetry(fmt.Sprintf("build collection %s", col.Name), func(c context.Context) *exec.Cmd {
			cCmd := exec.CommandContext(c, galaxyBin, "collection", "build")
			cCmd.Dir = extractDir
			envVars := activeEnv
			if envVars == nil {
				envVars = os.Environ()
			}
			cCmd.Env = envVars
			return cCmd
		}, 1, 0)

		if err != nil {
			return fmt.Errorf("failed to build collection %s: %w", col.Name, err)
		}

		// Collect the built tarball
		matches, _ := filepath.Glob(filepath.Join(extractDir, "*.tar.gz"))
		if len(matches) == 0 {
			return fmt.Errorf("failed to find built tarball for %s", col.Name)
		}
		builtTarballs = append(builtTarballs, matches[0])
	}

	// Install locally
	if len(builtTarballs) > 0 {
		args := []string{"collection", "install"}
		args = append(args, builtTarballs...)
		err := runWithRetry("ansible-galaxy collection install (offline source build)", func(c context.Context) *exec.Cmd {
			cCmd := exec.CommandContext(c, galaxyBin, args...)
			cCmd.Dir = workDir
			envVars := activeEnv
			if envVars == nil {
				envVars = os.Environ()
			}
			envVars = append(envVars, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
			cCmd.Env = envVars
			return cCmd
		}, 1, 0)
		if err != nil {
			return fmt.Errorf("failed to install built collections: %w", err)
		}
	}

	return nil
}
