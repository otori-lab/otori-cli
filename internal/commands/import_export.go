package commands

import (
	"fmt"
	"strings"

	"github.com/otori-lab/otori-cli/internal/config"
)

// ExportCommand exporte un profil
func ExportCommand(profileName, outputPath, format string) error {
	if profileName == "" {
		return fmt.Errorf("veuillez spécifier le profil à exporter")
	}

	// Normaliser le format
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
		return fmt.Errorf("format non supporté: %s (utiliser: yaml, csv)", format)
	}

	if err := config.ExportConfig(profileName, exportFormat, outputPath); err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = "exports/" + profileName + "." + format
	}

	fmt.Printf("✓ Profil '%s' exporté vers %s\n", profileName, outputPath)
	return nil
}

// ImportCommand importe une configuration
func ImportCommand(filePath, profileName string) error {
	if filePath == "" {
		return fmt.Errorf("veuillez spécifier le chemin du fichier à importer")
	}

	// Déterminer le format à partir de l'extension
	var ext string
	if len(filePath) > 4 {
		ext = strings.ToLower(filePath[len(filePath)-4:])
		if ext == ".yml" || ext == ".yaml" {
			if err := config.ImportYAML(filePath, profileName); err != nil {
				return err
			}
		} else if ext == ".json" {
			// JSON: lire et importer comme tel
			cfg, err := config.ReadConfig(filePath[:len(filePath)-5])
			if err != nil {
				return fmt.Errorf("erreur lecture JSON: %w", err)
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
			return fmt.Errorf("format non supporté: %s", ext)
		}
	} else {
		return fmt.Errorf("fichier invalide: %s", filePath)
	}

	if profileName == "" {
		profileName = "imported"
	}

	fmt.Printf("✓ Configuration importée sous le nom '%s'\n", profileName)
	return nil
}
