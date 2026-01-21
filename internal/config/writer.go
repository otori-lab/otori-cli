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

// WriteConfig writes the configuration to a profile directory
// For "classic" type: creates profile folder with JSON + cowrie.cfg + userdb.txt
// For "ia" type: creates profile folder with JSON only
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

	// Create profile directory (profiles/{profileName}/)
	profileDir := getProfileDir(config.ProfileName)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("error creating profile directory: %w", err)
	}

	// Build JSON filename
	filename := filepath.Join(profileDir, config.ProfileName+".json")

	// Encode configuration to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	// Write JSON file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// For classic type, also generate Cowrie config files
	if config.Type == "classic" {
		if err := WriteCowrieConfig(profileDir, config); err != nil {
			return fmt.Errorf("error writing cowrie.cfg: %w", err)
		}
		if err := WriteUserDB(profileDir, config); err != nil {
			return fmt.Errorf("error writing userdb.txt: %w", err)
		}
		if err := WriteHoneyFS(profileDir, config); err != nil {
			return fmt.Errorf("error writing honeyfs: %w", err)
		}
		if err := WriteDockerCompose(profileDir, config); err != nil {
			return fmt.Errorf("error writing docker-compose.yml: %w", err)
		}
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

	// Create profile directory (profiles/{profileName}/)
	profileDir := getProfileDir(profileName)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("error creating profile directory: %w", err)
	}

	// Build filename with specified name
	filename := filepath.Join(profileDir, profileName+".json")

	// Encode configuration to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	// Write JSON file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// For classic type, also generate Cowrie config files
	if config.Type == "classic" {
		if err := WriteCowrieConfig(profileDir, config); err != nil {
			return fmt.Errorf("error writing cowrie.cfg: %w", err)
		}
		if err := WriteUserDB(profileDir, config); err != nil {
			return fmt.Errorf("error writing userdb.txt: %w", err)
		}
		if err := WriteHoneyFS(profileDir, config); err != nil {
			return fmt.Errorf("error writing honeyfs: %w", err)
		}
		if err := WriteDockerCompose(profileDir, config); err != nil {
			return fmt.Errorf("error writing docker-compose.yml: %w", err)
		}
	}

	return nil
}

// ReadConfig reads a configuration from a profile directory
func ReadConfig(profileName string) (*models.Config, error) {
	if profileName == "" {
		profileName = "default"
	}

	// Try new structure first: profiles/{profileName}/{profileName}.json
	profileDir := getProfileDir(profileName)
	filename := filepath.Join(profileDir, profileName+".json")
	data, err := os.ReadFile(filename)

	// Fallback to old structure: profiles/{profileName}.json
	if err != nil {
		oldFilename := filepath.Join(getConfigDir(), profileName+".json")
		data, err = os.ReadFile(oldFilename)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
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
		// New structure: directories are profiles
		if entry.IsDir() {
			// Check if the profile JSON exists inside
			jsonFile := filepath.Join(configDir, entry.Name(), entry.Name()+".json")
			if _, err := os.Stat(jsonFile); err == nil {
				profiles = append(profiles, entry.Name())
			}
		}
		// Fallback: old structure with direct JSON files
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			profileName := entry.Name()[:len(entry.Name())-5]
			// Avoid duplicates if both structures exist
			found := false
			for _, p := range profiles {
				if p == profileName {
					found = true
					break
				}
			}
			if !found {
				profiles = append(profiles, profileName)
			}
		}
	}

	return profiles, nil
}

// getConfigDir returns the config directory path (~/.otori/profiles)
func getConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home not found
		return "profiles"
	}
	return filepath.Join(homeDir, ".otori", "profiles")
}

// GetConfigDir exports the config directory path for use by other packages
func GetConfigDir() string {
	return getConfigDir()
}

// getProfileDir returns the profile directory path for a specific profile
func getProfileDir(profileName string) string {
	return filepath.Join(getConfigDir(), profileName)
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
