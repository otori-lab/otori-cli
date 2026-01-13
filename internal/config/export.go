package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/otori-lab/otori-cli/internal/models"
	"gopkg.in/yaml.v3"
)

// ExportFormat représente le format d'export
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatYAML ExportFormat = "yaml"
	FormatCSV  ExportFormat = "csv"
)

// ExportConfig exporte une configuration dans le format spécifié
func ExportConfig(profileName string, format ExportFormat, outputPath string) error {
	cfg, err := ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profil '%s' non trouvé: %w", profileName, err)
	}

	if outputPath == "" {
		outputPath = filepath.Join("exports", profileName+"."+string(format))
	}

	// Créer le répertoire d'export s'il n'existe pas
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erreur création répertoire: %w", err)
	}

	switch format {
	case FormatYAML:
		return exportYAML(cfg, outputPath)
	case FormatCSV:
		return exportCSV(cfg, outputPath)
	case FormatJSON:
		// JSON déjà supporté nativement
		return fmt.Errorf("utiliser WriteConfig pour JSON")
	default:
		return fmt.Errorf("format non supporté: %s", format)
	}
}

// exportYAML exporte au format YAML
func exportYAML(config *models.Config, outputPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("erreur encoding YAML: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("erreur écriture fichier: %w", err)
	}

	return nil
}

// exportCSV exporte au format CSV
func exportCSV(config *models.Config, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// En-têtes
	headers := []string{"Type", "ServerName", "ProfileName", "Company", "Users"}
	writer.Write(headers)

	// Données
	usersStr := ""
	if len(config.Users) > 0 {
		for i, user := range config.Users {
			if user != "" {
				if i > 0 {
					usersStr += "; "
				}
				usersStr += user
			}
		}
	}

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

// ImportYAML importe une configuration depuis YAML
func ImportYAML(filePath string, profileName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erreur lecture fichier: %w", err)
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("erreur decode YAML: %w", err)
	}

	// Utiliser le nom du profil fourni ou celui du fichier
	if profileName == "" {
		profileName = config.ProfileName
		if profileName == "" {
			profileName = filepath.Base(filePath[:len(filePath)-len(filepath.Ext(filePath))])
		}
	}

	config.ProfileName = profileName
	config.CreatedAt = time.Now().Format(time.RFC3339)

	// Valider avant d'importer
	validationErrors := ValidateConfig(&config)
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration invalide: %v", validationErrors[0].Message)
	}

	return WriteConfig(&config)
}

// ImportCSV importe une configuration depuis CSV
func ImportCSV(filePath string, profileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erreur ouverture fichier: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("erreur lecture CSV: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV vide")
	}

	// Sauter les en-têtes et lire la première ligne
	row := records[1]
	if len(row) < 5 {
		return fmt.Errorf("CSV invalide: colonnes manquantes")
	}

	config := models.NewConfig()
	config.Type = row[0]
	config.ServerName = row[1]
	config.ProfileName = row[2]
	if profileName != "" {
		config.ProfileName = profileName
	}
	config.Company = row[3]

	// Parser les utilisateurs (séparés par "; ")
	if row[4] != "" {
		usersStr := row[4]
		// Traiter différents séparateurs
		if len(usersStr) > 0 {
			config.Users = []string{usersStr}
		}
	}

	config.CreatedAt = time.Now().Format(time.RFC3339)

	// Valider avant d'importer
	validationErrors := ValidateConfig(config)
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration invalide: %v", validationErrors[0].Message)
	}

	return WriteConfig(config)
}
