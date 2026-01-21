package commands

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/tui"
	"github.com/otori-lab/otori-cli/internal/ui"
	"github.com/spf13/cobra"
)

var statusProfile string
var statusJson bool
var statusAll bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of honeypots",
	Long:  "Display status of running honeypots. Use -a to show all profiles (including stopped).",
	Run: func(cmd *cobra.Command, args []string) {
		// Get running honeypots from Docker
		honeypots := tui.GetRunningHoneypots()

		// If --all flag, also include stopped profiles
		if statusAll {
			honeypots = addStoppedProfiles(honeypots)
		}

		// Filter by profile if specified
		if statusProfile != "" {
			var filtered []tui.Honeypot
			for _, hp := range honeypots {
				if hp.Profile == statusProfile {
					filtered = append(filtered, hp)
				}
			}
			honeypots = filtered
		}

		// JSON output mode
		if statusJson {
			fmt.Println(ui.GetLogo())
			outputJSON(honeypots)
			return
		}

		// Interactive TUI mode
		model := tui.NewStatusModel(honeypots)
		p := tea.NewProgram(model)

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running status: %v\n", err)
		}
	},
}

// addStoppedProfiles adds profiles that exist but are not running
func addStoppedProfiles(running []tui.Honeypot) []tui.Honeypot {
	// Get all profiles
	profiles, err := config.ListConfigs()
	if err != nil {
		return running
	}

	// Create a map of running profiles
	runningMap := make(map[string]bool)
	for _, hp := range running {
		runningMap[hp.Profile] = true
	}

	// Add stopped profiles
	for _, profileName := range profiles {
		if !runningMap[profileName] {
			// Read profile config to get details
			cfg, err := config.ReadConfig(profileName)
			if err != nil {
				continue
			}

			honeypot := tui.Honeypot{
				Name:       "otori-" + profileName,
				Profile:    profileName,
				Type:       cfg.Type,
				Status:     tui.StatusStopped,
				ServerName: cfg.ServerName,
				Port:       2222,
			}
			running = append(running, honeypot)
		}
	}

	return running
}

// outputJSON outputs honeypots as JSON
func outputJSON(honeypots []tui.Honeypot) {
	data, err := json.MarshalIndent(honeypots, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func init() {
	statusCmd.Flags().StringVarP(&statusProfile, "profile", "p", "", "Filter by profile name")
	statusCmd.Flags().BoolVarP(&statusJson, "json", "j", false, "Output as JSON")
	statusCmd.Flags().BoolVarP(&statusAll, "all", "a", false, "Show all profiles (including stopped)")

	RootCmd.AddCommand(statusCmd)
}
