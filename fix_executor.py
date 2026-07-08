import re
import os

with open('internal/pkg/orchestrator/executor.go', 'r') as f:
    content = f.read()

# Add imports
if '"runtime"' not in content:
    content = re.sub(
        r'"path/filepath"',
        '"path/filepath"\n\t"runtime"\n\t"strings"',
        content
    )

# Fix venv paths
venv_block_old = """	venvDir := filepath.Join(env.GetDataDir(), ".ansible", "venv")
	venvBin := filepath.Join(venvDir, "bin", "ansible-playbook")"""
venv_block_new = """	venvDir := filepath.Join(env.GetDataDir(), ".ansible", "venv")
	
	venvBinDir := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		venvBinDir = filepath.Join(venvDir, "Scripts")
	}
	venvBin := filepath.Join(venvBinDir, "ansible-playbook")"""
if venv_block_old in content:
    content = content.replace(venv_block_old, venv_block_new)

# Fix pythonCmd
python_block_old = """	// 3. Create the virtual environment
	pythonCmd := "python3"
	fmt.Println("🚀 Bootstrapping Python Virtual Environment for Ansible...")"""
python_block_new = """	// 3. Create the virtual environment
	pythonCmd := "python3"
	if _, err := exec.LookPath(pythonCmd); err != nil {
		if _, err := exec.LookPath("python"); err == nil {
			pythonCmd = "python"
		}
	}
	fmt.Println("🚀 Bootstrapping Python Virtual Environment for Ansible...")"""
if python_block_old in content:
    content = content.replace(python_block_old, python_block_new)

# Fix bin paths
content = content.replace('pipBin := filepath.Join(venvDir, "bin", "pip")', 'pipBin := filepath.Join(venvBinDir, "pip")')
content = content.replace('galaxyBin := filepath.Join(venvDir, "bin", "ansible-galaxy")', 'galaxyBin := filepath.Join(venvBinDir, "ansible-galaxy")')

# Fix buildVenvEnv
build_venv_env_old = """func buildVenvEnv(venvDir string) []string {
	env := os.Environ()
	pathIdx := -1
	for i, e := range env {
		if len(e) > 5 && e[:5] == "PATH=" {
			pathIdx = i
			break
		}
	}

	venvBinDir := filepath.Join(venvDir, "bin")
	if pathIdx != -1 {
		env[pathIdx] = fmt.Sprintf("PATH=%s:%s", venvBinDir, env[pathIdx][5:])
	} else {
		env = append(env, fmt.Sprintf("PATH=%s", venvBinDir))
	}

	env = append(env, fmt.Sprintf("VIRTUAL_ENV=%s", venvDir))
	return env
}"""
build_venv_env_new = """func buildVenvEnv(venvDir string) []string {
	venvBinDir := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		venvBinDir = filepath.Join(venvDir, "Scripts")
	}

	env := os.Environ()
	pathUpdated := false
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			env[i] = fmt.Sprintf("PATH=%s%c%s", venvBinDir, os.PathListSeparator, e[5:])
			pathUpdated = true
			break
		}
	}

	if !pathUpdated {
		env = append(env, fmt.Sprintf("PATH=%s", venvBinDir))
	}

	env = append(env, fmt.Sprintf("VIRTUAL_ENV=%s", venvDir))
	return env
}"""
if build_venv_env_old in content:
    content = content.replace(build_venv_env_old, build_venv_env_new)

with open('internal/pkg/orchestrator/executor.go', 'w') as f:
    f.write(content)

