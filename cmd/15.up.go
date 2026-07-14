package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/snowdreamtech/unistack/internal/orchestrator"
	"github.com/snowdreamtech/unistack/internal/runner"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up [scenario_file]",
	Short: "Deploy an entire stack scenario",
	Long:  "Read a scenario YAML file, resolve dependencies, and orchestrate the deployment of apps.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scenarioPath := args[0]

		fmt.Printf("📦 Parsing scenario: %s\n", scenarioPath)
		scenario, err := orchestrator.ParseScenario(scenarioPath)
		if err != nil {
			return fmt.Errorf("failed to parse scenario: %w", err)
		}

		fmt.Printf("🔄 Resolving dependencies for scenario '%s'...\n", scenario.Name)
		sortedApps, err := orchestrator.SortAppsByDependency(scenario.Apps)
		if err != nil {
			return fmt.Errorf("dependency resolution failed: %w", err)
		}

		// Since UniStack utilizes the unified 'app' engine, we just need a minimal playbook
		// to include that role. In real usage, you'd generate or point to a valid playbook.
		playbookPath := "ansible/roles/app/tasks/main.yml" 
		if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
			playbookPath = "run_app.yml"
		}

		engRunner := runner.NewAnsibleRunner(playbookPath)
		ctx := context.Background()

		fmt.Printf("🚀 Starting deployment of %d apps...\n\n", len(sortedApps))

		for i, app := range sortedApps {
			fmt.Printf("[%d/%d] 🟢 Deploying app: %s\n", i+1, len(sortedApps), app.Name)
			
			res, err := engRunner.RunApp(ctx, app.Name, app.Vars)
			if err != nil {
				fmt.Printf("❌ Failed to deploy %s: %v\n", app.Name, err)
				if res != nil && res.Error != nil {
					fmt.Printf("Detailed Error: %v\n", res.Error)
				}
				return fmt.Errorf("deployment halted due to failure in %s", app.Name)
			}

			fmt.Printf("✅ Successfully deployed %s\n", app.Name)
		}

		fmt.Println("\n🎉 All apps deployed successfully! Scenario complete.")
		return nil
	},
}
