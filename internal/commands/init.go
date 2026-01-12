package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initType string
var initProfileName string
var initServerName string
var initCompanyName string
var initUsers []string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a honeypot profile",
	Run: func(cmd *cobra.Command, args []string) {

		if cmd.Flags().NFlag() == 0 {
			fmt.Println("init interactive mode")
			return
		}

		//validation
		if initType == "" {
			fmt.Println("Error: --type is required in non-interactive mode")
			os.Exit(1)
		}
		if initServerName == "" {
			fmt.Println("Error: --server-name is required in non-interactive mode")
			os.Exit(1)
		}

		fmt.Println("init non-interactive mode")
		fmt.Println("Type is: ", initType)
		fmt.Println("Server name is: ", initServerName)
		if initProfileName != "" {
			fmt.Println("profile-name =", initProfileName)
		}
		if initCompanyName != "" {
			fmt.Println("company =", initCompanyName)
		}
		if len(initUsers) > 0 {
			fmt.Println("users =", initUsers)
		}

	},
}

func init() {
	initCmd.Flags().StringVarP(
		&initType,
		"type",
		"t",
		"",
		"Type of honeypot (classic or ia)",
	)

	initCmd.Flags().StringVarP(
		&initProfileName,
		"profile-name",
		"p",
		"",
		"Name of the profile to create",
	)

	initCmd.Flags().StringVarP(
		&initServerName,
		"server-name",
		"s",
		"",
		"Name of the server simulated by the honeypot",
	)

	initCmd.Flags().StringVarP(
		&initCompanyName,
		"company",
		"c",
		"",
		"Name of the company that own the honeypot",
	)

	initCmd.Flags().StringSliceVarP(
		&initUsers,
		"users",
		"u",
		[]string{},
		"Comma-separated list of fake users (e.g. root,admin,test)",
	)

	RootCmd.AddCommand(initCmd)
}
