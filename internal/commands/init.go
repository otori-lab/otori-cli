package commands

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/models"
	"github.com/otori-lab/otori-cli/internal/tui"
	"github.com/otori-lab/otori-cli/internal/ui"
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
			// Interactive mode with TUI Bubble Tea
			runInteractiveInit()
			return
		}

		// Display logo
		fmt.Println(ui.GetLogo())

		// Non-interactive mode: validate required fields
		if initType == "" {
			fmt.Println("Error: --type is required in non-interactive mode")
			os.Exit(1)
		}
		if initServerName == "" {
			fmt.Println("Error: --server-name is required in non-interactive mode")
			os.Exit(1)
		}

		// Normalize type to lowercase
		normalizedType := strings.ToLower(initType)

		// Create configuration
		cfg := models.NewConfig()
		cfg.Type = normalizedType
		cfg.ServerName = initServerName
		cfg.Company = initCompanyName
		cfg.Users = initUsers

		// Set profile name (default if empty)
		if initProfileName != "" {
			cfg.ProfileName = initProfileName
		} else {
			cfg.ProfileName = "default"
		}

		// Validate configuration
		validationErrors := config.ValidateConfig(cfg)
		if len(validationErrors) > 0 {
			fmt.Println("Validation errors:")
			for _, err := range validationErrors {
				fmt.Printf("  - %s: %s\n", err.Field, err.Message)
			}
			os.Exit(1)
		}

		// Save configuration
		if err := config.WriteConfig(cfg); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Profile '%s' created successfully!\n", cfg.ProfileName)

	},
}

// runInteractiveInit runs the interactive TUI to create a profile
func runInteractiveInit() {
	// Step 1: Launch the form
	formModel := tui.NewModel()
	formProgram := tea.NewProgram(formModel)

	finalFormModel, err := formProgram.Run()
	if err != nil {
		fmt.Printf("Error during form: %v\n", err)
		os.Exit(1)
	}

	// Check if user cancelled
	form, ok := finalFormModel.(tui.Model)
	if !ok {
		fmt.Println("Error: unable to retrieve form")
		os.Exit(1)
	}

	if form.IsCancelled() {
		fmt.Println("Configuration cancelled.")
		return
	}

	// Get configuration from form
	cfg := form.GetConfig()

	// Step 2: Launch preview for confirmation
	previewModel := tui.NewPreviewModel(cfg)
	previewProgram := tea.NewProgram(previewModel)

	finalPreviewModel, err := previewProgram.Run()
	if err != nil {
		fmt.Printf("Error during preview: %v\n", err)
		os.Exit(1)
	}

	preview, ok := finalPreviewModel.(tui.PreviewModel)
	if !ok {
		fmt.Println("Error: unable to retrieve preview")
		os.Exit(1)
	}

	if preview.IsCancelled() || !preview.IsConfirmed() {
		fmt.Println("Configuration cancelled.")
		return
	}

	// Step 3: Validate configuration
	validationErrors := config.ValidateConfig(cfg)
	if len(validationErrors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range validationErrors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
		os.Exit(1)
	}

	// Step 4: Save configuration
	if err := config.WriteConfig(cfg); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Profile '%s' created successfully!\n", cfg.ProfileName)
}

func init() {
	initCmd.Flags().StringVarP(
		&initType,
		"type",
		"t",
		"",
		"Type of honeypot: 'classic' or 'ia'",
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
