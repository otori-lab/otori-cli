package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/otori-lab/otori-cli/internal/models"
)

// WriteConfig écrit la configuration dans un fichier JSON
func WriteConfig(config *models.Config) error {
	// Ajouter le timestamp
	config.CreatedAt = time.Now().Format(time.RFC3339)

	// Créer le répertoire de configuration s'il n'existe pas
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("erreur création du répertoire config: %w", err)
	}

	// Construire le nom du fichier
	filename := filepath.Join(configDir, config.ProfileName+".json")

	// Encoder la configuration en JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur encoding JSON: %w", err)
	}

	// Écrire le fichier
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("erreur écriture fichier: %w", err)
	}

	return nil
}

// ReadConfig lit une configuration depuis un fichier JSON
func ReadConfig(profileName string) (*models.Config, error) {
	if profileName == "" {
		profileName = "default"
	}

	filename := filepath.Join(getConfigDir(), profileName+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier: %w", err)
	}

	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erreur decode JSON: %w", err)
	}

	return &config, nil
}

// ListConfigs liste tous les profils disponibles
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

// getConfigDir retourne le chemin du répertoire de configuration
func getConfigDir() string {
	// Créer un dossier 'profiles' dans le répertoire courant
	return "profiles"
}
