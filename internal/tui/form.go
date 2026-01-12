package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otori-lab/otori-cli/internal/models"
	"github.com/otori-lab/otori-cli/internal/ui"
)

// Model représente l'état du formulaire TUI
type Model struct {
	config       *models.Config
	currentField int
	fields       []Field
	finished     bool
	cancelled    bool
	err          string
}

// Field représente un champ du formulaire
type Field struct {
	name        string
	label       string
	value       string
	placeholder string
	required    bool
}

// NewModel crée un nouveau modèle de formulaire
func NewModel() Model {
	return Model{
		config: models.NewConfig(),
		fields: []Field{
			{
				name:        "type",
				label:       "Type de profil",
				placeholder: "classique ou IA",
				required:    true,
			},
			{
				name:        "serverName",
				label:       "Nom du serveur",
				placeholder: "ex: mon-serveur",
				required:    true,
			},
			{
				name:        "profileName",
				label:       "Nom du profil",
				placeholder: "default si vide",
				required:    false,
			},
			{
				name:        "company",
				label:       "Entreprise",
				placeholder: "optionnel",
				required:    false,
			},
			{
				name:        "users",
				label:       "Utilisateurs",
				placeholder: "séparés par des virgules",
				required:    false,
			},
		},
		currentField: 0,
	}
}

// Init initialise le modèle
func (m Model) Init() tea.Cmd {
	return nil
}

// Update gère les mises à jour du modèle
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			// Valider le champ
			if m.validateCurrentField() {
				m.err = ""
				if m.currentField < len(m.fields)-1 {
					m.currentField++
				} else {
					m.finished = true
					return m, tea.Quit
				}
			}

		case "up":
			if m.currentField > 0 {
				m.currentField--
				m.err = ""
			}

		case "down":
			if m.currentField < len(m.fields)-1 {
				m.currentField++
				m.err = ""
			}

		case "backspace":
			field := &m.fields[m.currentField]
			if len(field.value) > 0 {
				field.value = field.value[:len(field.value)-1]
				m.err = ""
			}

		default:
			// Ajouter le caractère au champ actuel
			field := &m.fields[m.currentField]
			field.value += msg.String()
			m.err = ""
		}
	}
	return m, nil
}

// View renvoie la vue du formulaire
func (m Model) View() string {
	if m.finished {
		return ""
	}

	var sb strings.Builder

	// Afficher le logo au-dessus
	sb.WriteString(ui.GetLogo())
	sb.WriteString("\n")

	// Styles
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Bold(true)

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	completeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("48"))

	placeholderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("206")).
		Padding(0, 1)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	// Titre du questionnaire
	sb.WriteString(titleStyle.Render("Configuration"))
	sb.WriteString("\n\n")

	// Afficher chaque champ
	for i, field := range m.fields {
		if i == m.currentField {
			// Champ actuel en édition
			label := activeStyle.Render("➜ " + field.label)
			if field.required {
				label += activeStyle.Render(" *")
			}
			sb.WriteString(label)
			sb.WriteString("\n")

			// Afficher l'input
			if field.value == "" {
				sb.WriteString(inputStyle.Render(placeholderStyle.Render(field.placeholder + " ▌")))
			} else {
				sb.WriteString(inputStyle.Render(field.value + "▌"))
			}
			sb.WriteString("\n")

			// Afficher les erreurs
			if m.err != "" {
				sb.WriteString(errorStyle.Render("✗ " + m.err))
				sb.WriteString("\n")
			}
		} else {
			// Champs précédents ou suivants
			label := "  " + field.label
			if field.required {
				label += " *"
			}
			sb.WriteString(labelStyle.Render(label))
			sb.WriteString("\n")

			if field.value == "" {
				sb.WriteString(placeholderStyle.Render("    " + field.placeholder))
			} else {
				sb.WriteString(completeStyle.Render("    ✓ " + field.value))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Instructions
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	sb.WriteString(instructionsStyle.Render("↑↓ Naviguer | Enter Valider | Ctrl+C Quitter"))
	sb.WriteString("\n")

	return sb.String()
}

// validateCurrentField valide le champ actuel
func (m *Model) validateCurrentField() bool {
	field := &m.fields[m.currentField]
	value := strings.TrimSpace(field.value)
	field.value = value

	if field.required && value == "" {
		m.err = fmt.Sprintf("%s est obligatoire", field.label)
		return false
	}

	if field.name == "type" && value != "" {
		lower := strings.ToLower(value)
		if lower != "classique" && lower != "ia" {
			m.err = "Le type doit être 'classique' ou 'IA'"
			return false
		}
		field.value = lower
	}

	return true
}

// GetConfig retourne la configuration remplie
func (m Model) GetConfig() *models.Config {
	config := models.NewConfig()

	for _, field := range m.fields {
		switch field.name {
		case "type":
			config.Type = field.value
		case "serverName":
			config.ServerName = field.value
		case "profileName":
			config.ProfileName = field.value
			if config.ProfileName == "" {
				config.ProfileName = "default"
			}
		case "company":
			config.Company = field.value
		case "users":
			if field.value != "" {
				parts := strings.Split(field.value, ",")
				for _, part := range parts {
					if trimmed := strings.TrimSpace(part); trimmed != "" {
						config.Users = append(config.Users, trimmed)
					}
				}
			}
		}
	}

	return config
}

// IsFinished retourne true si le formulaire est terminé
func (m Model) IsFinished() bool {
	return m.finished
}

// IsCancelled retourne true si l'utilisateur a annulé
func (m Model) IsCancelled() bool {
	return m.cancelled
}
