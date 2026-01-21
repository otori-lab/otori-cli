package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/ui"
	"github.com/spf13/cobra"
)

// profilesCmd is the parent command for profile management
var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage honeypot profiles",
	Long:  "List, show, and delete honeypot profiles",
}

// profilesListCmd lists all profiles
var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available profiles",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ListCommand(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

// profilesShowCmd shows a specific profile
var profilesShowCmd = &cobra.Command{
	Use:   "show [profile-name]",
	Short: "Show details of a profile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := "default"
		if len(args) > 0 {
			profileName = args[0]
		}
		if err := ShowCommand(profileName); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

// profilesDeleteCmd deletes a profile
var profilesDeleteCmd = &cobra.Command{
	Use:   "delete [profile-name]",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := DeleteCommand(args[0]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesShowCmd)
	profilesCmd.AddCommand(profilesDeleteCmd)
	RootCmd.AddCommand(profilesCmd)
}

// ListCommand lists all available profiles
func ListCommand() error {
	fmt.Println(ui.GetLogo())

	profiles, err := config.ListConfigs()
	if err != nil {
		return fmt.Errorf("error reading profiles: %w", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles found. Create one with: otori init")
		return nil
	}

	fmt.Println("\nAvailable profiles:\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROFILE\tTYPE\tSERVER\tCOMPANY\tCREATED")

	for _, name := range profiles {
		cfg, err := config.ReadConfig(name)
		if err != nil {
			fmt.Fprintf(w, "%s\t[error]\t-\t-\t-\n", name)
			continue
		}

		createdAt := cfg.CreatedAt
		if len(createdAt) > 16 {
			createdAt = createdAt[:16]
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			name, cfg.Type, cfg.ServerName, cfg.Company, createdAt)
	}

	w.Flush()
	fmt.Println()
	return nil
}

// ShowCommand displays profile details
func ShowCommand(profileName string) error {
	fmt.Println(ui.GetLogo())

	if profileName == "" {
		profileName = "default"
	}

	cfg, err := config.ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	fmt.Printf("\nProfile: %s\n\n", profileName)
	fmt.Printf("  Type:       %s\n", cfg.Type)
	fmt.Printf("  Server:     %s\n", cfg.ServerName)
	fmt.Printf("  Company:    %s\n", cfg.Company)
	fmt.Printf("  Created:    %s\n\n", cfg.CreatedAt)

	if len(cfg.Users) > 0 {
		fmt.Println("  Users:")
		for _, user := range cfg.Users {
			fmt.Printf("    - %s\n", user)
		}
	} else {
		fmt.Println("  Users: (none)")
	}
	fmt.Println()

	return nil
}

// DeleteCommand deletes a profile
func DeleteCommand(profileName string) error {
	fmt.Println(ui.GetLogo())

	if profileName == "" {
		return fmt.Errorf("please specify the profile name to delete")
	}

	// Confirmation
	fmt.Printf("Are you sure you want to delete profile '%s'? (yes/no): ", profileName)
	var response string
	fmt.Scanln(&response)

	if response != "yes" && response != "y" {
		fmt.Println("Deletion cancelled")
		return nil
	}

	// Find profile (new structure: directory, old structure: file)
	configDir := config.GetConfigDir()
	profileDir := filepath.Join(configDir, profileName)
	oldFilename := filepath.Join(configDir, profileName+".json")

	// Check new structure first (directory)
	if info, err := os.Stat(profileDir); err == nil && info.IsDir() {
		// Delete entire directory
		if err := os.RemoveAll(profileDir); err != nil {
			return fmt.Errorf("error deleting profile directory: %w", err)
		}
		fmt.Printf("✓ Profile '%s' deleted successfully\n", profileName)
		return nil
	}

	// Fallback to old structure (single file)
	if _, err := os.Stat(oldFilename); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	// Delete file
	if err := os.Remove(oldFilename); err != nil {
		return fmt.Errorf("error deleting profile: %w", err)
	}

	fmt.Printf("✓ Profile '%s' deleted successfully\n", profileName)
	return nil
}
