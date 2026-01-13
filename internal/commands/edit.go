package commands

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/tui"
)

// EditCommand opens an existing profile for editing
func EditCommand(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("please specify the profile name to edit")
	}

	// Load existing configuration
	cfg, err := config.ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	// Create form model pre-filled with existing data
	model := tui.NewModelWithConfig(cfg)

	// Launch TUI program
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Get final model
	m := finalModel.(tui.Model)

	// If user cancelled
	if m.IsCancelled() {
		fmt.Println("Edit cancelled")
		return nil
	}

	// Get final configuration
	finalConfig := m.GetConfig()
	finalConfig.CreatedAt = cfg.CreatedAt // Preserve creation date

	// Preserve profile name if user wants to keep it
	if finalConfig.ProfileName == "" {
		finalConfig.ProfileName = profileName
	}

	// Show preview before confirmation
	previewModel := tui.NewPreviewModel(finalConfig)
	p = tea.NewProgram(previewModel, tea.WithAltScreen())

	finalPreview, err := p.Run()
	if err != nil {
		return fmt.Errorf("preview error: %w", err)
	}

	preview := finalPreview.(tui.PreviewModel)

	// If user declined
	if !preview.IsConfirmed() || preview.IsCancelled() {
		fmt.Println("\nEdit cancelled")
		return nil
	}

	// Validate configuration before saving
	validationErrors := config.ValidateConfig(finalConfig)
	if len(validationErrors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range validationErrors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
		return fmt.Errorf("configuration validation failed")
	}

	// Write modified configuration
	if err := config.WriteConfigWithName(profileName, finalConfig); err != nil {
		return fmt.Errorf("save error: %w", err)
	}

	fmt.Printf("✓ Profile '%s' updated successfully\n", profileName)
	return nil
}
