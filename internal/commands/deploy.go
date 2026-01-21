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

var deployProfile string
var deployForce bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a honeypot",
	Long:  "Deploy a honeypot using Docker Compose from a profile configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.GetLogo())

		if err := runDeploy(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runDeploy() error {
	// Use default profile if not specified
	profileName := deployProfile
	if profileName == "" {
		profileName = "default"
	}

	// Read profile configuration
	cfg, err := config.ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	// Check if profile is classic type
	if cfg.Type != "classic" {
		return fmt.Errorf("profile '%s' is of type '%s', only 'classic' profiles can be deployed with Docker", profileName, cfg.Type)
	}

	// Get profile directory
	profileDir := filepath.Join(config.GetConfigDir(), profileName)

	// Check if docker-compose.yml exists
	dockerComposePath := filepath.Join(profileDir, "docker-compose.yml")
	if _, err := os.Stat(dockerComposePath); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found in profile '%s'", profileName)
	}

	fmt.Printf("Deploying honeypot from profile '%s'...\n", profileName)
	fmt.Printf("  Server: %s\n", cfg.ServerName)
	fmt.Printf("  Type: %s\n", cfg.Type)
	fmt.Println()

	// Build docker compose command
	var dockerCmd *exec.Cmd
	if deployForce {
		fmt.Println("Force recreating containers...")
		dockerCmd = exec.Command("docker", "compose", "up", "-d", "--force-recreate")
	} else {
		dockerCmd = exec.Command("docker", "compose", "up", "-d")
	}

	// Set working directory to profile directory
	dockerCmd.Dir = profileDir
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	// Run docker compose
	if err := dockerCmd.Run(); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Honeypot '%s' deployed successfully!\n", profileName)
	fmt.Println()
	fmt.Println("Honeypot is listening on:")
	fmt.Println("  SSH:    localhost:2222")
	fmt.Println("  Telnet: localhost:2223")
	fmt.Println()
	fmt.Println("To check status: otori status")
	fmt.Println("To stop:         otori stop -p", profileName)

	return nil
}

func init() {
	deployCmd.Flags().StringVarP(
		&deployProfile,
		"profile",
		"p",
		"",
		"Profile to deploy (default: 'default')",
	)

	deployCmd.Flags().BoolVarP(
		&deployForce,
		"force",
		"f",
		false,
		"Force recreate containers even if already running",
	)

	RootCmd.AddCommand(deployCmd)
}
