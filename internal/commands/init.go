package commands

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/models"
	"github.com/otori-lab/otori-cli/internal/tui"
	"github.com/spf13/cobra"
)

var initType string
var initProfileName string
var initServerName string
var initCompanyName string
var initUsers []string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a honeypot profile",
	Run: func(cmd *cobra.Command, args []string) {

		if cmd.Flags().NFlag() == 0 {
			// Mode interactif avec TUI Bubble Tea
			runInteractiveInit()
			return
		}

		// Mode non-interactif: validation des champs requis
		if initType == "" {
			fmt.Println("Error: --type is required in non-interactive mode")
			os.Exit(1)
		}
		if initServerName == "" {
			fmt.Println("Error: --server-name is required in non-interactive mode")
			os.Exit(1)
		}

		// Normaliser et valider le type
		normalizedType := strings.ToLower(initType)
		if normalizedType != "classic" && normalizedType != "ia" {
			fmt.Println("Error: --type must be 'classic' or 'ia'")
			os.Exit(1)
		}

		// Créer la configuration
		cfg := models.NewConfig()
		cfg.Type = normalizedType
		cfg.ServerName = initServerName
		cfg.Company = initCompanyName
		cfg.Users = initUsers

		// Définir le nom du profil (default si vide)
		if initProfileName != "" {
			cfg.ProfileName = initProfileName
		} else {
			cfg.ProfileName = "default"
		}

		// Sauvegarder la configuration
		if err := config.WriteConfig(cfg); err != nil {
			fmt.Printf("Erreur lors de la sauvegarde: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Profil '%s' créé avec succès!\n", cfg.ProfileName)

	},
}

// runInteractiveInit lance le TUI interactif pour créer un profil
func runInteractiveInit() {
	// Étape 1: Lancer le formulaire
	formModel := tui.NewModel()
	formProgram := tea.NewProgram(formModel)

	finalFormModel, err := formProgram.Run()
	if err != nil {
		fmt.Printf("Erreur lors du formulaire: %v\n", err)
		os.Exit(1)
	}

	// Vérifier si l'utilisateur a annulé
	form, ok := finalFormModel.(tui.Model)
	if !ok {
		fmt.Println("Erreur: impossible de récupérer le formulaire")
		os.Exit(1)
	}

	if form.IsCancelled() {
		fmt.Println("Configuration annulée.")
		return
	}

	// Récupérer la configuration du formulaire
	cfg := form.GetConfig()

	// Étape 2: Lancer le preview pour confirmation
	previewModel := tui.NewPreviewModel(cfg)
	previewProgram := tea.NewProgram(previewModel)

	finalPreviewModel, err := previewProgram.Run()
	if err != nil {
		fmt.Printf("Erreur lors du preview: %v\n", err)
		os.Exit(1)
	}

	preview, ok := finalPreviewModel.(tui.PreviewModel)
	if !ok {
		fmt.Println("Erreur: impossible de récupérer le preview")
		os.Exit(1)
	}

	if preview.IsCancelled() || !preview.IsConfirmed() {
		fmt.Println("Configuration annulée.")
		return
	}

	// Étape 3: Sauvegarder la configuration
	if err := config.WriteConfig(cfg); err != nil {
		fmt.Printf("Erreur lors de la sauvegarde: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Profil '%s' créé avec succès!\n", cfg.ProfileName)
}

func init() {
	initCmd.Flags().StringVarP(
		&initType,
		"type",
		"t",
		"",
		"Type of honeypot: 'classic' or 'ia'",
	)

	initCmd.Flags().StringVarP(
		&initProfileName,
		"profile-name",
		"p",
		"",
		"Name of the profile to create",
	)

	initCmd.Flags().StringVarP(
		&initServerName,
		"server-name",
		"s",
		"",
		"Name of the server simulated by the honeypot",
	)

	initCmd.Flags().StringVarP(
		&initCompanyName,
		"company",
		"c",
		"",
		"Name of the company that own the honeypot",
	)

	initCmd.Flags().StringSliceVarP(
		&initUsers,
		"users",
		"u",
		[]string{},
		"Comma-separated list of fake users (e.g. root,admin,test)",
	)

	RootCmd.AddCommand(initCmd)
}
