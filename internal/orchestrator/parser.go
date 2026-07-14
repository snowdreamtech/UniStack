package orchestrator

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unistack/internal/types"
	"gopkg.in/yaml.v3"
)

// ParseScenario reads a scenario YAML file and unmarshals it into a Scenario struct.
func ParseScenario(filePath string) (*types.Scenario, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file %s: %w", filePath, err)
	}

	var scenario types.Scenario
	err = yaml.Unmarshal(data, &scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scenario YAML: %w", err)
	}

	return &scenario, nil
}

// SortAppsByDependency performs a topological sort on the apps in the scenario
// to ensure they are deployed in the correct order based on depends_on.
func SortAppsByDependency(apps []types.App) ([]types.App, error) {
	appMap := make(map[string]types.App)
	for _, app := range apps {
		appMap[app.Name] = app
	}

	var result []types.App
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("circular dependency detected: %s", name)
		}
		if visited[name] {
			return nil
		}

		visiting[name] = true

		app, exists := appMap[name]
		if !exists {
			return fmt.Errorf("dependency %s not found in scenario apps", name)
		}

		for _, dep := range app.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		result = append(result, app)
		return nil
	}

	for _, app := range apps {
		if !visited[app.Name] {
			if err := visit(app.Name); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}
