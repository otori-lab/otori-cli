package config

import (
	"fmt"
	"strings"

	"github.com/otori-lab/otori-cli/internal/models"
)

// ValidationError représente une erreur de validation
type ValidationError struct {
	Field   string
	Message string
}

// ValidateConfig valide une configuration
func ValidateConfig(config *models.Config) []ValidationError {
	var errors []ValidationError

	// Vérifier le type
	if config.Type == "" {
		errors = append(errors, ValidationError{
			Field:   "Type",
			Message: "Type obligatoire",
		})
	} else if config.Type != "classique" && config.Type != "IA" {
		errors = append(errors, ValidationError{
			Field:   "Type",
			Message: "Type doit être 'classique' ou 'IA'",
		})
	}

	// Vérifier le serveur
	if config.ServerName == "" {
		errors = append(errors, ValidationError{
			Field:   "ServerName",
			Message: "Nom du serveur obligatoire",
		})
	} else if len(config.ServerName) < 3 {
		errors = append(errors, ValidationError{
			Field:   "ServerName",
			Message: "Nom du serveur doit avoir au moins 3 caractères",
		})
	}

	// Vérifier le profil
	if config.ProfileName == "" {
		errors = append(errors, ValidationError{
			Field:   "ProfileName",
			Message: "Nom du profil obligatoire",
		})
	} else if !isValidProfileName(config.ProfileName) {
		errors = append(errors, ValidationError{
			Field:   "ProfileName",
			Message: "Nom du profil doit contenir seulement des caractères alphanumériques, tirets et underscores",
		})
	}

	// Vérifier qu'il n'y a pas de doublons dans les utilisateurs
	uniqueUsers := make(map[string]bool)
	for _, user := range config.Users {
		if user != "" {
			lowerUser := strings.ToLower(user)
			if uniqueUsers[lowerUser] {
				errors = append(errors, ValidationError{
					Field:   "Users",
					Message: fmt.Sprintf("L'utilisateur '%s' est en doublon", user),
				})
			}
			uniqueUsers[lowerUser] = true
		}
	}

	return errors
}

// isValidProfileName vérifie si un nom de profil est valide
func isValidProfileName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}

	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}

// ProfileExists vérifie si un profil existe déjà
func ProfileExists(profileName string) bool {
	profiles, err := ListConfigs()
	if err != nil {
		return false
	}
	for _, p := range profiles {
		if p == profileName {
			return true
		}
	}
	return false
}
