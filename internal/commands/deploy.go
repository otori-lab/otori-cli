package commands

import (
	"fmt"

	"github.com/otori-lab/otori-cli/internal/ui"
	"github.com/spf13/cobra"
)

var deployProfile string
var deployForce bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a honeypot",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.GetLogo())
		fmt.Println("deploy called")

		if deployProfile == "" {
			fmt.Println("using default profile")
		} else {
			fmt.Println("using profile ->", deployProfile)
		}

		if deployForce {
			fmt.Println("forced deployment")
		}

	},
}

func init() {
	deployCmd.Flags().StringVarP(
		&deployProfile,
		"profile",
		"p",
		"",
		"Profile to deploy to",
	)

	deployCmd.Flags().BoolVarP(
		&deployForce,
		"force",
		"f",
		false,
		"Force deployment even if a honeypot is already running",
	)

	RootCmd.AddCommand(deployCmd)
}
