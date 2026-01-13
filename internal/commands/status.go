package commands

import (
	"fmt"

	"github.com/otori-lab/otori-cli/internal/ui"
	"github.com/spf13/cobra"
)

var statusProfile string
var statusJson bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of honeypots",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.GetLogo())
		fmt.Println("Display status of honeypots")

		if statusProfile != "" {
			fmt.Println("Status of the profile", statusProfile)
		}

		if statusJson {
			fmt.Println("Display status of honeypots as json")
		}

	},
}

func init() {
	statusCmd.Flags().StringVarP(&statusProfile, "profile", "p", "", "specify the profile to use")
	statusCmd.Flags().BoolVarP(&statusJson, "json", "j", false, "Display JSON output")

	RootCmd.AddCommand(statusCmd)
}
