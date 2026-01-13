package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/otori-lab/otori-cli/internal/config"
)

// ListCommand liste tous les profils disponibles
func ListCommand() error {
	profiles, err := config.ListConfigs()
	if err != nil {
		return fmt.Errorf("erreur lecture des profils: %w", err)
	}

	if len(profiles) == 0 {
		fmt.Println("Aucun profil trouvé. Créez-en un avec: otori init")
		return nil
	}

	fmt.Println("\n📋 Profils disponibles:\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROFIL\tTYPE\tSERVEUR\tENTREPRISE\tCRÉATION")

	for _, name := range profiles {
		cfg, err := config.ReadConfig(name)
		if err != nil {
			fmt.Fprintf(w, "%s\t[erreur]\t-\t-\t-\n", name)
			continue
		}

		createdAt := cfg.CreatedAt
		if len(createdAt) > 16 {
			createdAt = createdAt[:16]
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			name, cfg.Type, cfg.ServerName, cfg.Company, createdAt)
	}

	w.Flush()
	fmt.Println()
	return nil
}

// ShowCommand affiche les détails d'un profil
func ShowCommand(profileName string) error {
	if profileName == "" {
		profileName = "default"
	}

	cfg, err := config.ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profil '%s' non trouvé: %w", profileName, err)
	}

	fmt.Printf("\n📄 Profil: %s\n\n", profileName)
	fmt.Printf("  Type:        %s\n", cfg.Type)
	fmt.Printf("  Serveur:     %s\n", cfg.ServerName)
	fmt.Printf("  Entreprise:  %s\n", cfg.Company)
	fmt.Printf("  Créé:        %s\n\n", cfg.CreatedAt)

	if len(cfg.Users) > 0 {
		fmt.Println("  Utilisateurs:")
		for _, user := range cfg.Users {
			fmt.Printf("    • %s\n", user)
		}
	} else {
		fmt.Println("  Utilisateurs: (aucun)")
	}
	fmt.Println()

	return nil
}

// DeleteCommand supprime un profil
func DeleteCommand(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("veuillez spécifier le nom du profil à supprimer")
	}

	// Confirmation
	fmt.Printf("⚠️  Êtes-vous sûr de vouloir supprimer le profil '%s'? (oui/non): ", profileName)
	var response string
	fmt.Scanln(&response)

	if response != "oui" && response != "yes" && response != "y" {
		fmt.Println("❌ Suppression annulée")
		return nil
	}

	// Trouver le fichier
	configDir := "profiles"
	filename := filepath.Join(configDir, profileName+".json")

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("profil '%s' non trouvé", profileName)
	}

	// Supprimer
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("erreur suppression du profil: %w", err)
	}

	fmt.Printf("✓ Profil '%s' supprimé avec succès\n", profileName)
	return nil
}
