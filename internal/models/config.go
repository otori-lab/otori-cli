package models

// Config représente la configuration du profil Otori
type Config struct {
	Type        string   `json:"type"`        // classique ou IA
	ServerName  string   `json:"serverName"`  // obligatoire
	ProfileName string   `json:"profileName"` // default si non spécifié
	Company     string   `json:"company"`     // optionnel
	Users       []string `json:"users"`       // optionnel
	CreatedAt   string   `json:"createdAt"`   // timestamp de création
}

// NewConfig crée une nouvelle configuration
func NewConfig() *Config {
	return &Config{
		ProfileName: "default",
		Users:       []string{},
	}
}
