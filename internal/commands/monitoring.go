package commands

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/otori-lab/otori-cli/internal/ui"
	"github.com/spf13/cobra"
)

// Constants for monitoring containers
const (
	monitoringContainerName = "otori-monitoring"
	dbContainerName         = "otori-db"
	monitoringNetwork       = "otori-network"
	monitoringImage         = "ghcr.io/otori-lab/otori-monitoring:latest"
	postgresImage           = "postgres:16-alpine"
	defaultMonitoringPort   = 8000
)

// Flags
var (
	monitoringPort  int
	logsFollow      bool
	logsLines       int
)

// monitoringCmd is the parent command for all monitoring subcommands
var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Manage the Otori monitoring server",
	Long:  "Manage the Otori monitoring server for collecting and visualizing honeypot data",
}

// monitoringStartCmd starts the monitoring stack
var monitoringStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the monitoring server",
	Long:  "Start the monitoring server (PostgreSQL + API) using Docker",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.GetLogo())

		if err := runMonitoringStart(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// monitoringStopCmd stops the monitoring stack
var monitoringStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the monitoring server",
	Long:  "Stop the monitoring server and database containers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.GetLogo())

		if err := runMonitoringStop(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// monitoringStatusCmd shows the status of monitoring containers
var monitoringStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display monitoring server status",
	Long:  "Display the status of the monitoring server and database containers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.GetLogo())

		if err := runMonitoringStatus(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// monitoringLogsCmd shows logs from the monitoring container
var monitoringLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View monitoring server logs",
	Long:  "View logs from the monitoring server container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMonitoringLogs(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// runMonitoringStart starts the monitoring stack
func runMonitoringStart() error {
	// Check if already running
	if isContainerRunning(monitoringContainerName) {
		port := getMonitoringPort()
		fmt.Printf("Monitoring is already running at http://localhost:%d\n", port)
		return nil
	}

	fmt.Println("Starting Otori monitoring server...")
	fmt.Println()

	// Step 1: Create network if it doesn't exist
	fmt.Print("Creating network... ")
	if err := createNetworkIfNotExists(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("failed to create network: %w", err)
	}
	fmt.Println("OK")

	// Step 2: Start PostgreSQL
	fmt.Print("Starting database... ")
	if err := startPostgres(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("failed to start database: %w", err)
	}
	fmt.Println("OK")

	// Step 3: Wait for PostgreSQL to be ready
	fmt.Print("Waiting for database to be ready... ")
	if err := waitForContainerHealthy(dbContainerName, 60*time.Second); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("database failed to become healthy: %w", err)
	}
	fmt.Println("OK")

	// Step 4: Start monitoring server
	fmt.Print("Starting monitoring server... ")
	if err := startMonitoringServer(); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("failed to start monitoring server: %w", err)
	}
	fmt.Println("OK")

	// Step 5: Wait for API to be ready
	fmt.Print("Waiting for API to be ready... ")
	if err := waitForAPIReady(30 * time.Second); err != nil {
		fmt.Println("FAILED")
		fmt.Println()
		fmt.Println("Warning: API did not respond in time, but containers are running.")
		fmt.Println("Check logs with: otori monitoring logs")
	} else {
		fmt.Println("OK")
	}

	fmt.Println()
	fmt.Printf("Monitoring started successfully!\n")
	fmt.Println()
	fmt.Printf("  Dashboard: http://localhost:%d\n", monitoringPort)
	fmt.Printf("  API:       http://localhost:%d/api\n", monitoringPort)
	fmt.Printf("  Health:    http://localhost:%d/health\n", monitoringPort)
	fmt.Println()
	fmt.Println("For honeypots running in Docker, use:")
	fmt.Printf("  MONITORING_URL=http://host.docker.internal:%d\n", monitoringPort)

	return nil
}

// runMonitoringStop stops the monitoring stack
func runMonitoringStop() error {
	fmt.Println("Stopping Otori monitoring server...")
	fmt.Println()

	// Stop monitoring container first
	if isContainerRunning(monitoringContainerName) || containerExists(monitoringContainerName) {
		fmt.Print("Stopping monitoring server... ")
		if err := stopAndRemoveContainer(monitoringContainerName); err != nil {
			fmt.Println("FAILED")
			fmt.Printf("  Warning: %v\n", err)
		} else {
			fmt.Println("OK")
		}
	} else {
		fmt.Println("Monitoring server is not running")
	}

	// Stop database container
	if isContainerRunning(dbContainerName) || containerExists(dbContainerName) {
		fmt.Print("Stopping database... ")
		if err := stopAndRemoveContainer(dbContainerName); err != nil {
			fmt.Println("FAILED")
			fmt.Printf("  Warning: %v\n", err)
		} else {
			fmt.Println("OK")
		}
	} else {
		fmt.Println("Database is not running")
	}

	fmt.Println()
	fmt.Println("Monitoring stopped")
	fmt.Println("Note: Data volume 'otori-pgdata' is preserved for persistence")

	return nil
}

// runMonitoringStatus shows the status of monitoring containers
func runMonitoringStatus() error {
	fmt.Println("Monitoring Status")
	fmt.Println("=================")
	fmt.Println()

	// Check database status
	dbStatus := "Stopped"
	dbHealth := ""
	if isContainerRunning(dbContainerName) {
		dbStatus = "Running"
		if isContainerHealthy(dbContainerName) {
			dbHealth = " (healthy)"
		} else {
			dbHealth = " (unhealthy)"
		}
	}
	fmt.Printf("  Database:   %s%s\n", dbStatus, dbHealth)

	// Check monitoring server status
	apiStatus := "Stopped"
	apiURL := ""
	if isContainerRunning(monitoringContainerName) {
		apiStatus = "Running"
		port := getMonitoringPort()
		apiURL = fmt.Sprintf(" (http://localhost:%d)", port)

		// Test API health endpoint
		if isAPIHealthy(port) {
			apiStatus = "Running"
		} else {
			apiStatus = "Running (API not responding)"
		}
	}
	fmt.Printf("  API Server: %s%s\n", apiStatus, apiURL)

	// Check network
	networkStatus := "Not created"
	if networkExists(monitoringNetwork) {
		networkStatus = "Created"
	}
	fmt.Printf("  Network:    %s\n", networkStatus)

	// Check volume
	volumeStatus := "Not created"
	if volumeExists("otori-pgdata") {
		volumeStatus = "Created"
	}
	fmt.Printf("  Data Volume: %s\n", volumeStatus)

	return nil
}

// runMonitoringLogs shows logs from the monitoring container
func runMonitoringLogs() error {
	if !containerExists(monitoringContainerName) {
		return fmt.Errorf("monitoring server is not running. Start it with: otori monitoring start")
	}

	args := []string{"logs"}
	if logsFollow {
		args = append(args, "-f")
	}
	args = append(args, "-n", strconv.Itoa(logsLines), monitoringContainerName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Helper functions

// createNetworkIfNotExists creates the Docker network if it doesn't exist
func createNetworkIfNotExists() error {
	if networkExists(monitoringNetwork) {
		return nil
	}

	cmd := exec.Command("docker", "network", "create", monitoringNetwork)
	return cmd.Run()
}

// networkExists checks if a Docker network exists
func networkExists(name string) bool {
	cmd := exec.Command("docker", "network", "inspect", name)
	return cmd.Run() == nil
}

// volumeExists checks if a Docker volume exists
func volumeExists(name string) bool {
	cmd := exec.Command("docker", "volume", "inspect", name)
	return cmd.Run() == nil
}

// startPostgres starts the PostgreSQL container
func startPostgres() error {
	// Check if already running
	if isContainerRunning(dbContainerName) {
		return nil
	}

	// Remove existing stopped container if any
	if containerExists(dbContainerName) {
		exec.Command("docker", "rm", dbContainerName).Run()
	}

	cmd := exec.Command("docker", "run", "-d",
		"--name", dbContainerName,
		"--network", monitoringNetwork,
		"-e", "POSTGRES_USER=otori",
		"-e", "POSTGRES_PASSWORD=otori_secret",
		"-e", "POSTGRES_DB=otori",
		"-v", "otori-pgdata:/var/lib/postgresql/data",
		"--health-cmd", "pg_isready -U otori",
		"--health-interval", "5s",
		"--health-timeout", "5s",
		"--health-retries", "5",
		"--restart", "unless-stopped",
		postgresImage,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}

	return nil
}

// startMonitoringServer starts the monitoring server container
func startMonitoringServer() error {
	// Check if already running
	if isContainerRunning(monitoringContainerName) {
		return nil
	}

	// Remove existing stopped container if any
	if containerExists(monitoringContainerName) {
		exec.Command("docker", "rm", monitoringContainerName).Run()
	}

	// Pull latest image
	pullCmd := exec.Command("docker", "pull", monitoringImage)
	pullCmd.Run() // Ignore pull errors, image might be cached

	cmd := exec.Command("docker", "run", "-d",
		"--name", monitoringContainerName,
		"--network", monitoringNetwork,
		"-p", fmt.Sprintf("%d:8000", monitoringPort),
		"-e", "DATABASE_URL=postgresql://otori:otori_secret@otori-db:5432/otori",
		"--restart", "unless-stopped",
		monitoringImage,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}

	return nil
}

// isContainerRunning checks if a container is running
func isContainerRunning(name string) bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// containerExists checks if a container exists (running or stopped)
func containerExists(name string) bool {
	cmd := exec.Command("docker", "inspect", name)
	return cmd.Run() == nil
}

// isContainerHealthy checks if a container's health check is passing
func isContainerHealthy(name string) bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Health.Status}}", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "healthy"
}

// waitForContainerHealthy waits for a container to become healthy
func waitForContainerHealthy(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if isContainerHealthy(name) {
			return nil
		}

		// Check if container is still running
		if !isContainerRunning(name) {
			return fmt.Errorf("container %s stopped unexpectedly", name)
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for container %s to become healthy", name)
}

// waitForAPIReady waits for the monitoring API to respond
func waitForAPIReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://localhost:%d/health", monitoringPort)

	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for API to become ready")
}

// isAPIHealthy checks if the monitoring API is responding
func isAPIHealthy(port int) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/health", port)

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// getMonitoringPort gets the port the monitoring container is mapped to
func getMonitoringPort() int {
	cmd := exec.Command("docker", "inspect", "-f",
		"{{range $p, $conf := .NetworkSettings.Ports}}{{if eq $p \"8000/tcp\"}}{{(index $conf 0).HostPort}}{{end}}{{end}}",
		monitoringContainerName)

	output, err := cmd.Output()
	if err != nil {
		return defaultMonitoringPort
	}

	port, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return defaultMonitoringPort
	}

	return port
}

// stopAndRemoveContainer stops and removes a container
func stopAndRemoveContainer(name string) error {
	// Stop the container
	stopCmd := exec.Command("docker", "stop", name)
	if err := stopCmd.Run(); err != nil {
		// Container might already be stopped
	}

	// Remove the container
	rmCmd := exec.Command("docker", "rm", name)
	return rmCmd.Run()
}

func init() {
	// Add flags to start command
	monitoringStartCmd.Flags().IntVarP(&monitoringPort, "port", "p", defaultMonitoringPort,
		"Port for the monitoring server")

	// Add flags to logs command
	monitoringLogsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false,
		"Follow log output")
	monitoringLogsCmd.Flags().IntVarP(&logsLines, "lines", "n", 50,
		"Number of lines to show")

	// Add subcommands to monitoring command
	monitoringCmd.AddCommand(monitoringStartCmd)
	monitoringCmd.AddCommand(monitoringStopCmd)
	monitoringCmd.AddCommand(monitoringStatusCmd)
	monitoringCmd.AddCommand(monitoringLogsCmd)

	// Add monitoring command to root
	RootCmd.AddCommand(monitoringCmd)
}
