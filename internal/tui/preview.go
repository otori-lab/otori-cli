package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otori-lab/otori-cli/internal/models"
	"github.com/otori-lab/otori-cli/internal/ui"
)

// PreviewModel displays a configuration preview before saving
type PreviewModel struct {
	config      *models.Config
	confirmed   bool
	cancelled   bool
	selectedYes bool
}

// NewPreviewModel creates a new preview model
func NewPreviewModel(config *models.Config) PreviewModel {
	return PreviewModel{
		config:      config,
		selectedYes: true, // Default to confirm
	}
}

// Init initializes the model
func (m PreviewModel) Init() tea.Cmd {
	return nil
}

// Update handles user actions
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

// View displays the preview
func (m PreviewModel) View() string {
	var sb strings.Builder

	// Logo
	sb.WriteString(ui.GetLogo())
	sb.WriteString("\n\n")

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Bold(true)

	sb.WriteString(titleStyle.Render("Configuration Preview"))
	sb.WriteString("\n\n")

	// Display data
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("48"))

	sb.WriteString(labelStyle.Render("Type:"))
	sb.WriteString("         ")
	sb.WriteString(valueStyle.Render(m.config.Type))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Server:"))
	sb.WriteString("       ")
	sb.WriteString(valueStyle.Render(m.config.ServerName))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Profile:"))
	sb.WriteString("      ")
	sb.WriteString(valueStyle.Render(m.config.ProfileName))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Company:"))
	sb.WriteString("      ")
	sb.WriteString(valueStyle.Render(m.config.Company))
	sb.WriteString("\n")

	sb.WriteString(labelStyle.Render("Users:"))
	sb.WriteString("\n")
	for _, user := range m.config.Users {
		if user != "" {
			sb.WriteString("  • ")
			sb.WriteString(valueStyle.Render(user))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")

	// Buttons
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

	sb.WriteString(yesStyle.Render(" YES "))
	sb.WriteString("  ")
	sb.WriteString(noStyle.Render(" NO "))
	sb.WriteString("\n\n")

	// Instructions
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	sb.WriteString(instructionsStyle.Render("← → Select | Enter Confirm | Ctrl+C Cancel"))
	sb.WriteString("\n")

	return sb.String()
}

// IsConfirmed returns true if the user confirmed
func (m PreviewModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user cancelled
func (m PreviewModel) IsCancelled() bool {
	return m.cancelled
}
