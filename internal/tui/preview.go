package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otori-lab/otori-cli/internal/models"
	"github.com/otori-lab/otori-cli/internal/ui"
)

// PreviewModel affiche un aperçu de la configuration avant validation
type PreviewModel struct {
	config      *models.Config
	confirmed   bool
	cancelled   bool
	selectedYes bool
}

// NewPreviewModel crée un nouveau modèle de preview
func NewPreviewModel(config *models.Config) PreviewModel {
	return PreviewModel{
		config:      config,
		selectedYes: true, // Par défaut on confirme
	}
}

// Init initialise le modèle
func (m PreviewModel) Init() tea.Cmd {
	return nil
}

// Update gère les actions utilisateur
func (m PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "left", "h":
			m.selectedYes = true

		case "right", "l":
			m.selectedYes = false

		case "enter":
			if m.selectedYes {
				m.confirmed = true
			} else {
				m.cancelled = true
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

// View affiche l'aperçu
func (m PreviewModel) View() string {
	var sb strings.Builder

	// Logo
	sb.WriteString(ui.GetLogo())
	sb.WriteString("\n\n")

	// Titre
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Bold(true)

	sb.WriteString(titleStyle.Render("Aperçu de la configuration"))
	sb.WriteString("\n\n")

	// Afficher les données
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("48"))

	sb.WriteString(labelStyle.Render("Type:"))
	sb.WriteString("           ")
	sb.WriteString(valueStyle.Render(m.config.Type))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Serveur:"))
	sb.WriteString("        ")
	sb.WriteString(valueStyle.Render(m.config.ServerName))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Profil:"))
	sb.WriteString("         ")
	sb.WriteString(valueStyle.Render(m.config.ProfileName))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Entreprise:"))
	sb.WriteString("      ")
	sb.WriteString(valueStyle.Render(m.config.Company))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Utilisateurs:"))
	sb.WriteString("\n")
	for _, user := range m.config.Users {
		if user != "" {
			sb.WriteString("  • ")
			sb.WriteString(valueStyle.Render(user))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")

	// Boutons
	yesStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)
	noStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)

	if m.selectedYes {
		yesStyle = yesStyle.
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("48"))
		noStyle = noStyle.
			Foreground(lipgloss.Color("240"))
	} else {
		yesStyle = yesStyle.
			Foreground(lipgloss.Color("240"))
		noStyle = noStyle.
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("196"))
	}

	sb.WriteString(yesStyle.Render(" OUI "))
	sb.WriteString("  ")
	sb.WriteString(noStyle.Render(" NON "))
	sb.WriteString("\n\n")

	// Instructions
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	sb.WriteString(instructionsStyle.Render("← → Choisir | Enter Confirmer | Ctrl+C Annuler"))
	sb.WriteString("\n")

	return sb.String()
}

// IsConfirmed retourne true si l'utilisateur a confirmé
func (m PreviewModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled retourne true si l'utilisateur a annulé
func (m PreviewModel) IsCancelled() bool {
	return m.cancelled
}
