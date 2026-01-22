package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

	// Update fs.pickle with custom honeyfs entries
	containerName := "otori-" + profileName
	fmt.Println()
	fmt.Println("Updating filesystem structure...")

	// Wait for container to be fully ready
	time.Sleep(3 * time.Second)

	// Get custom paths from honeyfs that need to be added to fs.pickle
	honeyfsDir := filepath.Join(profileDir, "honeyfs")
	fsctlCommands := generateFsctlCommands(honeyfsDir)

	if len(fsctlCommands) > 0 {
		// Build fsctl command string (one command per line, ending with exit)
		fsctlInput := strings.Join(fsctlCommands, "\n") + "\nexit\n"

		// Run fsctl inside the container using Python with correct PYTHONPATH
		// Command: docker exec -i -e PYTHONPATH=/cowrie/cowrie-git/src <container>
		//          /cowrie/cowrie-env/bin/python3 -m cowrie.scripts.fsctl
		//          /cowrie/cowrie-git/src/cowrie/data/fs.pickle
		fsctlCmd := exec.Command("docker", "exec", "-i",
			"-e", "PYTHONPATH=/cowrie/cowrie-git/src",
			containerName,
			"/cowrie/cowrie-env/bin/python3", "-m", "cowrie.scripts.fsctl",
			"/cowrie/cowrie-git/src/cowrie/data/fs.pickle")

		// Pipe the commands to fsctl stdin
		fsctlCmd.Stdin = strings.NewReader(fsctlInput)
		fsctlCmd.Stdout = os.Stdout
		fsctlCmd.Stderr = os.Stderr

		if err := fsctlCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to update fs.pickle: %v\n", err)
		} else {
			fmt.Printf("  Added %d custom entries to filesystem\n", len(fsctlCommands))

			// Restart container to reload fs.pickle
			fmt.Println("Restarting honeypot to apply changes...")
			restartCmd := exec.Command("docker", "compose", "restart")
			restartCmd.Dir = profileDir
			restartCmd.Stdout = os.Stdout
			restartCmd.Stderr = os.Stderr

			if err := restartCmd.Run(); err != nil {
				fmt.Printf("Warning: failed to restart container: %v\n", err)
			}
		}
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

// generateFsctlCommands scans the honeyfs directory and generates fsctl commands
// for custom paths that don't exist in the base Cowrie fs.pickle
func generateFsctlCommands(honeyfsDir string) []string {
	var commands []string

	// Base paths that already exist in Cowrie's fs.pickle (no need to add)
	existingPaths := map[string]bool{
		"/etc":           true,
		"/etc/passwd":    true,
		"/etc/shadow":    true,
		"/etc/group":     true,
		"/etc/hostname":  true,
		"/etc/hosts":     true,
		"/etc/host.conf": true,
		"/etc/inittab":   true,
		"/etc/issue":     true,
		"/etc/issue.net": true,
		"/etc/motd":      true,
		"/etc/resolv.conf": true,
		"/proc":          true,
		"/proc/cpuinfo":  true,
		"/proc/meminfo":  true,
		"/proc/mounts":   true,
		"/proc/version":  true,
		"/proc/modules":  true,
		"/proc/net":      true,
		"/proc/net/arp":  true,
	}

	// Track directories we've already added
	addedDirs := make(map[string]bool)

	// Walk through honeyfs to find custom paths
	filepath.Walk(honeyfsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Get relative path from honeyfs
		relPath, err := filepath.Rel(honeyfsDir, path)
		if err != nil || relPath == "." {
			return nil
		}

		// Convert to absolute path as it appears in the honeypot
		absPath := "/" + relPath

		// Skip if this path already exists in base Cowrie
		if existingPaths[absPath] {
			return nil
		}

		if info.IsDir() {
			// Add mkdir command for new directories
			if !addedDirs[absPath] {
				commands = append(commands, fmt.Sprintf("mkdir %s", absPath))
				addedDirs[absPath] = true
			}
		} else {
			// Ensure parent directory is created first
			parentDir := filepath.Dir(absPath)
			if !existingPaths[parentDir] && !addedDirs[parentDir] {
				commands = append(commands, fmt.Sprintf("mkdir %s", parentDir))
				addedDirs[parentDir] = true
			}
			// Add touch command for new files
			commands = append(commands, fmt.Sprintf("touch %s", absPath))
		}

		return nil
	})

	return commands
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
