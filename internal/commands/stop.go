package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/ui"
	"github.com/spf13/cobra"
)

var stopProfile string
var stopForce bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running honeypot",
	Long:  "Stop a running honeypot container using Docker Compose",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.GetLogo())

		if err := runStop(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runStop() error {
	// Use default profile if not specified
	profileName := stopProfile
	if profileName == "" {
		profileName = "default"
	}

	// Check if profile exists
	_, err := config.ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	// Get profile directory
	profileDir := filepath.Join(config.GetConfigDir(), profileName)

	// Check if docker-compose.yml exists
	dockerComposePath := filepath.Join(profileDir, "docker-compose.yml")
	if _, err := os.Stat(dockerComposePath); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found in profile '%s'", profileName)
	}

	fmt.Printf("Stopping honeypot '%s'...\n", profileName)

	// Build docker compose command
	var dockerCmd *exec.Cmd
	if stopForce {
		// Force stop with timeout 0
		dockerCmd = exec.Command("docker", "compose", "down", "-t", "0")
	} else {
		dockerCmd = exec.Command("docker", "compose", "down")
	}

	// Set working directory to profile directory
	dockerCmd.Dir = profileDir
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	// Run docker compose down
	if err := dockerCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop containers: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Honeypot '%s' stopped successfully!\n", profileName)

	return nil
}

func init() {
	stopCmd.Flags().StringVarP(
		&stopProfile,
		"profile",
		"p",
		"",
		"Profile to stop (default: 'default')",
	)

	stopCmd.Flags().BoolVarP(
		&stopForce,
		"force",
		"f",
		false,
		"Force stop (immediate shutdown)",
	)

	RootCmd.AddCommand(stopCmd)
}
