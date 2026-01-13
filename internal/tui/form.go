package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otori-lab/otori-cli/internal/config"
	"github.com/otori-lab/otori-cli/internal/models"
	"github.com/otori-lab/otori-cli/internal/ui"
)

// FieldType represents a field type
type FieldType string

const (
	FieldTypeText   FieldType = "text"
	FieldTypeSelect FieldType = "select"
	FieldTypeList   FieldType = "list"
)

// Model represents the TUI form state
type Model struct {
	config       *models.Config
	currentField int
	fields       []Field
	finished     bool
	cancelled    bool
	err          string

	// For selectors
	selectIndex map[string]int
	// For lists
	listInput string
	listUsers []string
}

// Field represents a form field
type Field struct {
	name        string
	label       string
	value       string
	placeholder string
	required    bool
	fieldType   FieldType
	options     []string // for selectors
}

// NewModel creates a new form model
func NewModel() Model {
	return createModel("", nil)
}

// NewModelWithConfig creates a new form model pre-filled with an existing configuration
func NewModelWithConfig(cfg *models.Config) Model {
	return createModel("edit", cfg)
}

// createModel is an internal function to create the model
func createModel(mode string, cfg *models.Config) Model {
	// Initialize with default values
	typeValue := "classic"
	serverValue := ""
	profileValue := ""
	companyValue := ""
	var usersList []string
	selectTypeIndex := 0

	// If editing an existing config, pre-fill fields
	if cfg != nil {
		typeValue = cfg.Type
		serverValue = cfg.ServerName
		profileValue = cfg.ProfileName
		companyValue = cfg.Company
		usersList = cfg.Users

		// Find the selected type index (normalize to lowercase)
		if strings.ToLower(typeValue) == "ia" {
			selectTypeIndex = 1
		}
	}

	model := Model{
		config: models.NewConfig(),
		fields: []Field{
			{
				name:      "type",
				label:     "Profile type",
				required:  true,
				fieldType: FieldTypeSelect,
				options:   []string{"classic", "ia"},
				value:     typeValue,
			},
			{
				name:        "serverName",
				label:       "Server name",
				placeholder: "e.g. my-server",
				required:    true,
				fieldType:   FieldTypeText,
				value:       serverValue,
			},
			{
				name:        "profileName",
				label:       "Profile name",
				placeholder: "default if empty",
				fieldType:   FieldTypeText,
				value:       profileValue,
			},
			{
				name:        "company",
				label:       "Company",
				placeholder: "optional",
				fieldType:   FieldTypeText,
				value:       companyValue,
			},
			{
				name:        "users",
				label:       "Users",
				placeholder: "one per line (Enter to add, Ctrl+D to finish)",
				fieldType:   FieldTypeList,
			},
		},
		selectIndex: map[string]int{"type": selectTypeIndex},
		listInput:   "",
		listUsers:   usersList,
	}

	return model
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles model updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "ctrl+d":
			// For lists: finish input
			if m.fields[m.currentField].fieldType == FieldTypeList {
				if m.listInput != "" {
					cleaned := strings.TrimSpace(m.listInput)
					cleaned = strings.Trim(cleaned, "\x00")
					if cleaned != "" {
						m.listUsers = append(m.listUsers, cleaned)
					}
					m.listInput = ""
				}
				// Move to next field or finish
				if m.currentField < len(m.fields)-1 {
					m.currentField++
					m.listInput = ""
				} else {
					m.finished = true
					return m, tea.Quit
				}
			}

		case "enter":
			field := &m.fields[m.currentField]

			// Handle based on field type
			switch field.fieldType {
			case FieldTypeSelect:
				// Selector: validate and move to next
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
				// List: add user and continue
				if m.listInput != "" {
					cleaned := strings.TrimSpace(m.listInput)
					cleaned = strings.Trim(cleaned, "\x00")
					if cleaned != "" {
						m.listUsers = append(m.listUsers, cleaned)
					}
					m.listInput = ""
				} else {
					// If empty, finish the list and move to next field
					if m.currentField < len(m.fields)-1 {
						m.currentField++
						m.listInput = ""
					} else {
						m.finished = true
						return m, tea.Quit
					}
				}

			case FieldTypeText:
				// Text: validate and move to next
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
				// Navigate in selector
				idx := m.selectIndex[field.name]
				if idx > 0 {
					idx--
					m.selectIndex[field.name] = idx
					field.value = field.options[idx]
				}
			} else if m.currentField > 0 {
				m.currentField--
				m.listInput = ""
				m.err = ""
			}

		case "down":
			field := &m.fields[m.currentField]
			if field.fieldType == FieldTypeSelect {
				// Navigate in selector
				idx := m.selectIndex[field.name]
				if idx < len(field.options)-1 {
					idx++
					m.selectIndex[field.name] = idx
					field.value = field.options[idx]
				}
			} else if m.currentField < len(m.fields)-1 {
				m.currentField++
				m.listInput = ""
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

// View returns the form view
func (m Model) View() string {
	if m.finished {
		return ""
	}

	var sb strings.Builder

	// Display logo
	sb.WriteString(ui.GetLogo())
	sb.WriteString("\n")

	// Display project info
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

	// Form title
	sb.WriteString(titleStyle.Render("Configuration"))
	sb.WriteString("\n\n")

	// Display each field
	for i, field := range m.fields {
		if i == m.currentField {
			// Current field being edited
			label := activeStyle.Render("➜ " + field.label)
			if field.required {
				label += activeStyle.Render(" *")
			}
			sb.WriteString(label)
			sb.WriteString("\n")

			// Display based on field type
			if field.fieldType == FieldTypeSelect {
				// Selector
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
				// User list
				if len(m.listUsers) > 0 {
					for _, user := range m.listUsers {
						sb.WriteString(completeStyle.Render("  ✓ " + user + "\n"))
					}
				}
				// Current input
				if m.listInput == "" {
					sb.WriteString(inputStyle.Render(placeholderStyle.Render("▌ " + field.placeholder)))
				} else {
					sb.WriteString(inputStyle.Render(m.listInput + "▌"))
				}
				sb.WriteString("\n")

			} else {
				// Text field
				if field.value == "" {
					sb.WriteString(inputStyle.Render(placeholderStyle.Render(field.placeholder + " ▌")))
				} else {
					sb.WriteString(inputStyle.Render(field.value + "▌"))
				}
				sb.WriteString("\n")
			}

			// Display errors
			if m.err != "" {
				sb.WriteString(errorStyle.Render("✗ " + m.err))
				sb.WriteString("\n")
			}

		} else {
			// Previous or next fields
			label := "  " + field.label
			if field.required {
				label += " *"
			}
			sb.WriteString(labelStyle.Render(label))
			sb.WriteString("\n")

			// Display field state
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
		sb.WriteString(instructionsStyle.Render("↑↓ Select | Enter Confirm | Ctrl+C Quit"))
	} else if fieldType == FieldTypeList {
		sb.WriteString(instructionsStyle.Render("Enter Add | Ctrl+D Finish | Ctrl+C Quit"))
	} else {
		sb.WriteString(instructionsStyle.Render("↑↓ Navigate | Enter Confirm | Ctrl+C Quit"))
	}
	sb.WriteString("\n")

	return sb.String()
}

// validateCurrentField validates the current field
func (m *Model) validateCurrentField() bool {
	field := &m.fields[m.currentField]
	value := strings.TrimSpace(field.value)
	field.value = value

	if field.required && value == "" {
		m.err = fmt.Sprintf("%s is required", field.label)
		return false
	}

	// Field-specific validations
	switch field.name {
	case "serverName":
		if len(value) < 3 {
			m.err = "Server name must be at least 3 characters"
			return false
		}

	case "profileName":
		if len(value) > 0 && !config.IsValidProfileName(value) {
			m.err = "Use only letters, numbers, hyphens and underscores"
			return false
		}
	}

	return true
}

// cleanUser cleans a user entry
func cleanUser(user string) string {
	cleaned := strings.TrimSpace(user)

	// Remove all null and control characters
	var result strings.Builder
	for _, r := range cleaned {
		if r >= 32 && r != 127 { // Keep only printable characters
			result.WriteRune(r)
		}
	}

	return result.String()
}

// GetConfig returns the filled configuration
func (m Model) GetConfig() *models.Config {
	cfg := models.NewConfig()

	for _, field := range m.fields {
		switch field.name {
		case "type":
			cfg.Type = field.value
		case "serverName":
			cfg.ServerName = field.value
		case "profileName":
			cfg.ProfileName = field.value
			if cfg.ProfileName == "" {
				cfg.ProfileName = "default"
			}
		case "company":
			cfg.Company = field.value
		case "users":
			// Clean and add users (without null or empty characters)
			for _, user := range m.listUsers {
				cleaned := cleanUser(user)
				if cleaned != "" {
					cfg.Users = append(cfg.Users, cleaned)
				}
			}
		}
	}

	return cfg
}

// IsFinished returns true if the form is finished
func (m Model) IsFinished() bool {
	return m.finished
}

// IsCancelled returns true if the user cancelled
func (m Model) IsCancelled() bool {
	return m.cancelled
}
