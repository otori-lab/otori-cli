package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otori-lab/otori-cli/internal/models"
	"github.com/otori-lab/otori-cli/internal/ui"
)

// FieldType représente le type de champ
type FieldType string

const (
	FieldTypeText   FieldType = "text"
	FieldTypeSelect FieldType = "select"
	FieldTypeList   FieldType = "list"
)

// Model représente l'état du formulaire TUI
type Model struct {
	config       *models.Config
	currentField int
	fields       []Field
	finished     bool
	cancelled    bool
	err          string

	// Pour les sélecteurs
	selectIndex map[string]int
	// Pour les listes
	listInput string
	listUsers []string
}

// Field représente un champ du formulaire
type Field struct {
	name        string
	label       string
	value       string
	placeholder string
	required    bool
	fieldType   FieldType
	options     []string // pour les sélecteurs
}

// NewModel crée un nouveau modèle de formulaire
func NewModel() Model {
	return Model{
		config: models.NewConfig(),
		fields: []Field{
			{
				name:      "type",
				label:     "Type de profil",
				required:  true,
				fieldType: FieldTypeSelect,
				options:   []string{"classique", "IA"},
				value:     "classique",
			},
			{
				name:        "serverName",
				label:       "Nom du serveur",
				placeholder: "ex: mon-serveur",
				required:    true,
				fieldType:   FieldTypeText,
			},
			{
				name:        "profileName",
				label:       "Nom du profil",
				placeholder: "default si vide",
				fieldType:   FieldTypeText,
			},
			{
				name:        "company",
				label:       "Entreprise",
				placeholder: "optionnel",
				fieldType:   FieldTypeText,
			},
			{
				name:        "users",
				label:       "Utilisateurs",
				placeholder: "un nom par ligne (Enter pour ajouter, Ctrl+D pour terminer)",
				fieldType:   FieldTypeList,
			},
		},
		selectIndex: make(map[string]int),
		listInput:   "",
		listUsers:   []string{},
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

		case "ctrl+d":
			// Pour les listes: terminer la saisie
			if m.fields[m.currentField].fieldType == FieldTypeList {
				if m.listInput != "" {
					cleaned := strings.TrimSpace(m.listInput)
					cleaned = strings.Trim(cleaned, "\x00")
					if cleaned != "" {
						m.listUsers = append(m.listUsers, cleaned)
					}
					m.listInput = ""
				}
				// Passer au champ suivant
				if m.currentField < len(m.fields)-1 {
					m.currentField++
					m.listInput = ""
					m.listUsers = []string{}
				} else {
					m.finished = true
					return m, tea.Quit
				}
			}

		case "enter":
			field := &m.fields[m.currentField]

			// Gestion selon le type de champ
			switch field.fieldType {
			case FieldTypeSelect:
				// Sélecteur: valider et passer au suivant
				if m.validateCurrentField() {
					m.err = ""
					if m.currentField < len(m.fields)-1 {
						m.currentField++
					} else {
						m.finished = true
						return m, tea.Quit
					}
				}

			case FieldTypeList:
				// Liste: ajouter l'utilisateur et continuer
				if m.listInput != "" {
					cleaned := strings.TrimSpace(m.listInput)
					cleaned = strings.Trim(cleaned, "\x00")
					if cleaned != "" {
						m.listUsers = append(m.listUsers, cleaned)
					}
					m.listInput = ""
				} else {
					// Si vide et on a des utilisateurs, passer au suivant
					if len(m.listUsers) > 0 {
						if m.currentField < len(m.fields)-1 {
							m.currentField++
							m.listInput = ""
							m.listUsers = []string{}
						} else {
							m.finished = true
							return m, tea.Quit
						}
					}
				}

			case FieldTypeText:
				// Texte: valider et passer au suivant
				if m.validateCurrentField() {
					m.err = ""
					if m.currentField < len(m.fields)-1 {
						m.currentField++
					} else {
						m.finished = true
						return m, tea.Quit
					}
				}
			}

		case "up":
			field := &m.fields[m.currentField]
			if field.fieldType == FieldTypeSelect {
				// Naviguer dans le sélecteur
				idx := m.selectIndex[field.name]
				if idx > 0 {
					idx--
					m.selectIndex[field.name] = idx
					field.value = field.options[idx]
				}
			} else if m.currentField > 0 {
				m.currentField--
				m.listInput = ""
				m.listUsers = []string{}
				m.err = ""
			}

		case "down":
			field := &m.fields[m.currentField]
			if field.fieldType == FieldTypeSelect {
				// Naviguer dans le sélecteur
				idx := m.selectIndex[field.name]
				if idx < len(field.options)-1 {
					idx++
					m.selectIndex[field.name] = idx
					field.value = field.options[idx]
				}
			} else if m.currentField < len(m.fields)-1 {
				m.currentField++
				m.listInput = ""
				m.listUsers = []string{}
				m.err = ""
			}

		case "backspace":
			field := &m.fields[m.currentField]
			if field.fieldType == FieldTypeList {
				if len(m.listInput) > 0 {
					m.listInput = m.listInput[:len(m.listInput)-1]
					m.err = ""
				}
			} else {
				if len(field.value) > 0 {
					field.value = field.value[:len(field.value)-1]
					m.err = ""
				}
			}

		default:
			field := &m.fields[m.currentField]
			if field.fieldType == FieldTypeList {
				m.listInput += msg.String()
				m.err = ""
			} else if field.fieldType != FieldTypeSelect {
				field.value += msg.String()
				m.err = ""
			}
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

	// Afficher les infos du projet
	sb.WriteString(ui.GetProjectInfo())
	sb.WriteString("\n")
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

	selectActiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("48")).
		Background(lipgloss.Color("206")).
		Bold(true)

	selectInactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

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

			// Afficher selon le type de champ
			if field.fieldType == FieldTypeSelect {
				// Sélecteur
				for j, option := range field.options {
					if j == m.selectIndex[field.name] {
						sb.WriteString(selectActiveStyle.Render(" ● " + option + " "))
					} else {
						sb.WriteString(selectInactiveStyle.Render(" ○ " + option))
					}
					sb.WriteString("  ")
				}
				sb.WriteString("\n")

			} else if field.fieldType == FieldTypeList {
				// Liste d'utilisateurs
				if len(m.listUsers) > 0 {
					for _, user := range m.listUsers {
						sb.WriteString(completeStyle.Render("  ✓ " + user + "\n"))
					}
				}
				// Input actuel
				if m.listInput == "" {
					sb.WriteString(inputStyle.Render(placeholderStyle.Render("▌ " + field.placeholder)))
				} else {
					sb.WriteString(inputStyle.Render(m.listInput + "▌"))
				}
				sb.WriteString("\n")

			} else {
				// Champ texte
				if field.value == "" {
					sb.WriteString(inputStyle.Render(placeholderStyle.Render(field.placeholder + " ▌")))
				} else {
					sb.WriteString(inputStyle.Render(field.value + "▌"))
				}
				sb.WriteString("\n")
			}

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

			// Afficher l'état du champ
			if field.fieldType == FieldTypeSelect {
				sb.WriteString(completeStyle.Render("    ✓ " + field.value + "\n"))

			} else if field.fieldType == FieldTypeList {
				if len(m.listUsers) == 0 {
					sb.WriteString(placeholderStyle.Render("    " + field.placeholder + "\n"))
				} else {
					for _, user := range m.listUsers {
						sb.WriteString(completeStyle.Render("    ✓ " + user + "\n"))
					}
				}

			} else {
				if field.value == "" {
					sb.WriteString(placeholderStyle.Render("    " + field.placeholder))
				} else {
					sb.WriteString(completeStyle.Render("    ✓ " + field.value))
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	// Instructions
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	fieldType := m.fields[m.currentField].fieldType
	if fieldType == FieldTypeSelect {
		sb.WriteString(instructionsStyle.Render("↑↓ Choisir | Enter Valider | Ctrl+C Quitter"))
	} else if fieldType == FieldTypeList {
		sb.WriteString(instructionsStyle.Render("Enter Ajouter | Ctrl+D Terminer | Ctrl+C Quitter"))
	} else {
		sb.WriteString(instructionsStyle.Render("↑↓ Naviguer | Enter Valider | Ctrl+C Quitter"))
	}
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

	// Validations spécifiques selon le champ
	switch field.name {
	case "serverName":
		if len(value) < 3 {
			m.err = "Le serveur doit avoir au moins 3 caractères"
			return false
		}

	case "profileName":
		if len(value) > 0 && !isValidProfileName(value) {
			m.err = "Utilisez seulement des lettres, chiffres, tirets et underscores"
			return false
		}
	}

	return true
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
			// Nettoyer et ajouter les utilisateurs (sans caractères nuls ou vides)
			for _, user := range m.listUsers {
				cleaned := strings.TrimSpace(user)
				// Enlever les caractères nuls
				cleaned = strings.Trim(cleaned, "\x00")
				if cleaned != "" {
					config.Users = append(config.Users, cleaned)
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
