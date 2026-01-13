package commands

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otori-lab/otori-cli/internal/tui"
	"github.com/otori-lab/otori-cli/internal/ui"
	"github.com/spf13/cobra"
)

var statusProfile string
var statusJson bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of honeypots",
	Run: func(cmd *cobra.Command, args []string) {
		// Get honeypots (using mock data for now)
		honeypots := tui.GetMockHoneypots()

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

	RootCmd.AddCommand(statusCmd)
}
