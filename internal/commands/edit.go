package commands

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/tui"
)

// EditCommand ouvre un profil existant pour édition
func EditCommand(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("veuillez spécifier le nom du profil à éditer")
	}

	// Charger la configuration existante
	cfg, err := config.ReadConfig(profileName)
	if err != nil {
		return fmt.Errorf("profil '%s' non trouvé: %w", profileName, err)
	}

	// Créer le modèle de formulaire
	model := tui.NewModel()

	// Lancer le programme TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("erreur TUI: %w", err)
	}

	// Récupérer le modèle final
	m := finalModel.(tui.Model)

	// Si l'utilisateur a annulé
	if m.IsCancelled() {
		fmt.Println("❌ Édition annulée")
		return nil
	}

	// Obtenir la configuration finale
	finalConfig := m.GetConfig()
	finalConfig.CreatedAt = cfg.CreatedAt // Préserver la date de création
	
	// Préserver le nom du profil si l'utilisateur veut le garder
	if finalConfig.ProfileName == "" {
		finalConfig.ProfileName = profileName
	}

	// Écrire la configuration modifiée
	if err := config.WriteConfigWithName(profileName, finalConfig); err != nil {
		return fmt.Errorf("erreur sauvegarde: %w", err)
	}

	fmt.Printf("✓ Profil '%s' modifié avec succès\n", profileName)
	return nil
}

