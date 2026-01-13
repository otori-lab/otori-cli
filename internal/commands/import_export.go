package commands

import (
	"fmt"
	"strings"

	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/ui"
)

// ExportCommand exports a profile
func ExportCommand(profileName, outputPath, format string) error {
	fmt.Println(ui.GetLogo())

	if profileName == "" {
		return fmt.Errorf("please specify the profile to export")
	}

	// Normalize format
	format = strings.ToLower(format)
	if format == "" {
		format = "yaml"
	}

	var exportFormat config.ExportFormat
	switch format {
	case "yaml", "yml":
		exportFormat = config.FormatYAML
	case "csv":
		exportFormat = config.FormatCSV
	case "json":
		exportFormat = config.FormatJSON
	default:
		return fmt.Errorf("unsupported format: %s (use: yaml, csv)", format)
	}

	if err := config.ExportConfig(profileName, exportFormat, outputPath); err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = "exports/" + profileName + "." + format
	}

	fmt.Printf("✓ Profile '%s' exported to %s\n", profileName, outputPath)
	return nil
}

// ImportCommand imports a configuration
func ImportCommand(filePath, profileName string) error {
	fmt.Println(ui.GetLogo())

	if filePath == "" {
		return fmt.Errorf("please specify the file path to import")
	}

	// Determine format from extension
	var ext string
	if len(filePath) > 4 {
		ext = strings.ToLower(filePath[len(filePath)-4:])
		if ext == ".yml" || ext == ".yaml" {
			if err := config.ImportYAML(filePath, profileName); err != nil {
				return err
			}
		} else if ext == ".json" {
			// JSON: read and import as is
			cfg, err := config.ReadConfig(filePath[:len(filePath)-5])
			if err != nil {
				return fmt.Errorf("error reading JSON: %w", err)
			}
			if profileName != "" {
				cfg.ProfileName = profileName
			}
			if err := config.WriteConfig(cfg); err != nil {
				return err
			}
		} else if ext == ".csv" {
			if err := config.ImportCSV(filePath, profileName); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported format: %s", ext)
		}
	} else {
		return fmt.Errorf("invalid file: %s", filePath)
	}

	if profileName == "" {
		profileName = "imported"
	}

	fmt.Printf("✓ Configuration imported as '%s'\n", profileName)
	return nil
}
