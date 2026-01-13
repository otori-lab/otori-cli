package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otori-lab/otori-cli/internal/models"
	"gopkg.in/yaml.v3"
)

// ExportFormat represents the export format
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatYAML ExportFormat = "yaml"
	FormatCSV  ExportFormat = "csv"
)

// ExportConfig exports a configuration in the specified format
func ExportConfig(profileName string, format ExportFormat, outputPath string) error {
	cfg, err := ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	if outputPath == "" {
		outputPath = filepath.Join("exports", profileName+"."+string(format))
	}

	// Create export directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	switch format {
	case FormatYAML:
		return exportYAML(cfg, outputPath)
	case FormatCSV:
		return exportCSV(cfg, outputPath)
	case FormatJSON:
		// JSON already supported natively
		return fmt.Errorf("use WriteConfig for JSON")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportYAML exports to YAML format
func exportYAML(config *models.Config, outputPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error encoding YAML: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// exportCSV exports to CSV format
func exportCSV(config *models.Config, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Headers
	headers := []string{"Type", "ServerName", "ProfileName", "Company", "Users"}
	writer.Write(headers)

	// Data - join users with "; " separator
	usersStr := strings.Join(config.Users, "; ")

	row := []string{
		config.Type,
		config.ServerName,
		config.ProfileName,
		config.Company,
		usersStr,
	}
	writer.Write(row)

	return nil
}

// ImportYAML imports a configuration from YAML
func ImportYAML(filePath string, profileName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error decoding YAML: %w", err)
	}

	// Use provided profile name or the one from file
	if profileName == "" {
		profileName = config.ProfileName
		if profileName == "" {
			profileName = filepath.Base(filePath[:len(filePath)-len(filepath.Ext(filePath))])
		}
	}

	config.ProfileName = profileName
	config.CreatedAt = time.Now().Format(time.RFC3339)

	// Validate before importing
	validationErrors := ValidateConfig(&config)
	if len(validationErrors) > 0 {
		return fmt.Errorf("invalid configuration: %v", validationErrors[0].Message)
	}

	return WriteConfig(&config)
}

// ImportCSV imports a configuration from CSV
func ImportCSV(filePath string, profileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("empty CSV")
	}

	// Skip headers and read first data row
	row := records[1]
	if len(row) < 5 {
		return fmt.Errorf("invalid CSV: missing columns")
	}

	config := models.NewConfig()
	config.Type = strings.ToLower(row[0]) // Normalize type to lowercase
	config.ServerName = row[1]
	config.ProfileName = row[2]
	if profileName != "" {
		config.ProfileName = profileName
	}
	config.Company = row[3]

	// Parse users (separated by "; ")
	if row[4] != "" {
		usersStr := row[4]
		// Split by "; " separator and clean each user
		users := strings.Split(usersStr, "; ")
		for _, user := range users {
			cleaned := strings.TrimSpace(user)
			if cleaned != "" {
				config.Users = append(config.Users, cleaned)
			}
		}
	}

	config.CreatedAt = time.Now().Format(time.RFC3339)

	// Validate before importing
	validationErrors := ValidateConfig(config)
	if len(validationErrors) > 0 {
		return fmt.Errorf("invalid configuration: %v", validationErrors[0].Message)
	}

	return WriteConfig(config)
}
