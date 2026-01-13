package config

import (
	"fmt"
	"strings"

	"github.com/otori-lab/otori-cli/internal/models"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidateConfig validates a configuration
func ValidateConfig(config *models.Config) []ValidationError {
	var errors []ValidationError

	// Check type
	if config.Type == "" {
		errors = append(errors, ValidationError{
			Field:   "Type",
			Message: "Type is required",
		})
	} else if config.Type != "classic" && config.Type != "ia" {
		errors = append(errors, ValidationError{
			Field:   "Type",
			Message: "Type must be 'classic' or 'ia'",
		})
	}

	// Check server name
	if config.ServerName == "" {
		errors = append(errors, ValidationError{
			Field:   "ServerName",
			Message: "Server name is required",
		})
	} else if len(config.ServerName) < 3 {
		errors = append(errors, ValidationError{
			Field:   "ServerName",
			Message: "Server name must be at least 3 characters",
		})
	}

	// Check profile name
	if config.ProfileName == "" {
		errors = append(errors, ValidationError{
			Field:   "ProfileName",
			Message: "Profile name is required",
		})
	} else if !IsValidProfileName(config.ProfileName) {
		errors = append(errors, ValidationError{
			Field:   "ProfileName",
			Message: "Profile name must contain only alphanumeric characters, hyphens and underscores",
		})
	}

	// Check for duplicate users
	uniqueUsers := make(map[string]bool)
	for _, user := range config.Users {
		if user != "" {
			lowerUser := strings.ToLower(user)
			if uniqueUsers[lowerUser] {
				errors = append(errors, ValidationError{
					Field:   "Users",
					Message: fmt.Sprintf("Duplicate user '%s'", user),
				})
			}
			uniqueUsers[lowerUser] = true
		}
	}

	return errors
}

// IsValidProfileName checks if a profile name is valid (exported for reuse)
func IsValidProfileName(name string) bool {
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

// ProfileExists checks if a profile already exists
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
