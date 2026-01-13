package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otori-lab/otori-cli/internal/models"
)

// WriteConfig writes the configuration to a JSON file
func WriteConfig(config *models.Config) error {
	// Add timestamp
	config.CreatedAt = time.Now().Format(time.RFC3339)

	// Clean users (remove null and empty characters)
	var cleanedUsers []string
	for _, user := range config.Users {
		cleaned := strings.TrimSpace(user)
		// Remove all null and control characters
		cleaned = removeNullChars(cleaned)
		if cleaned != "" {
			cleanedUsers = append(cleanedUsers, cleaned)
		}
	}
	config.Users = cleanedUsers

	// Create config directory if it doesn't exist
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Build filename
	filename := filepath.Join(configDir, config.ProfileName+".json")

	// Encode configuration to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	// Write file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// WriteConfigWithName writes a configuration with a specific profile name (for editing)
func WriteConfigWithName(profileName string, config *models.Config) error {
	// Clean users (remove null and empty characters)
	var cleanedUsers []string
	for _, user := range config.Users {
		cleaned := strings.TrimSpace(user)
		cleaned = removeNullChars(cleaned)
		if cleaned != "" {
			cleanedUsers = append(cleanedUsers, cleaned)
		}
	}
	config.Users = cleanedUsers

	// Create config directory if it doesn't exist
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Build filename with specified name
	filename := filepath.Join(configDir, profileName+".json")

	// Encode configuration to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	// Write file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// ReadConfig reads a configuration from a JSON file
func ReadConfig(profileName string) (*models.Config, error) {
	if profileName == "" {
		profileName = "default"
	}

	filename := filepath.Join(getConfigDir(), profileName+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	return &config, nil
}

// ListConfigs lists all available profiles
func ListConfigs() ([]string, error) {
	configDir := getConfigDir()
	entries, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			profiles = append(profiles, entry.Name()[:len(entry.Name())-5])
		}
	}

	return profiles, nil
}

// getConfigDir returns the config directory path
func getConfigDir() string {
	// Create 'profiles' folder in current directory
	return "profiles"
}

// removeNullChars removes null and control characters
func removeNullChars(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Keep only printable characters (>= 32, != 127)
		if r >= 32 && r != 127 {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}
