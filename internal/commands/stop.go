package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopForce bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping honeypot...")

		if stopForce {
			fmt.Println("force stop")
		}
	},
}

func init() {
	stopCmd.Flags().BoolVarP(
		&stopForce,
		"force",
		"f",
		false,
		"force stop even if a honeypot is already running	",
	)

	RootCmd.AddCommand(stopCmd)
}
