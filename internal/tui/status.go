package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otori-lab/otori-cli/internal/ui"
)

// HoneypotStatus represents the status of a honeypot
type HoneypotStatus string

const (
	StatusActive  HoneypotStatus = "active"
	StatusStopped HoneypotStatus = "stopped"
	StatusError   HoneypotStatus = "error"
)

// Honeypot represents a honeypot instance
type Honeypot struct {
	Name       string         `json:"name"`
	Profile    string         `json:"profile"`
	Type       string         `json:"type"`
	Status     HoneypotStatus `json:"status"`
	Uptime     string         `json:"uptime,omitempty"`
	LastError  string         `json:"last_error,omitempty"`
	ServerName string         `json:"server_name"`
	Port       int            `json:"port"`
}

// StatusModel represents the TUI model for status display
type StatusModel struct {
	honeypots []Honeypot
	blinkOn   bool
	quitting  bool
}

// tickMsg is sent periodically for animations
type tickMsg time.Time

// NewStatusModel creates a new status model
func NewStatusModel(honeypots []Honeypot) StatusModel {
	return StatusModel{
		honeypots: honeypots,
		blinkOn:   true,
	}
}

// Init initializes the model
func (m StatusModel) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd returns a command that sends a tick every 800ms
func tickCmd() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages
func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case tickMsg:
		m.blinkOn = !m.blinkOn
		return m, tickCmd()
	}

	return m, nil
}

// View renders the status display
func (m StatusModel) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder

	// Logo
	sb.WriteString(ui.GetLogo())
	sb.WriteString("\n")

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Bold(true)

	sb.WriteString(titleStyle.Render("Honeypot Status"))
	sb.WriteString("\n\n")

	if len(m.honeypots) == 0 {
		noHoneypotStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		sb.WriteString(noHoneypotStyle.Render("No honeypots running. Deploy one with: otori deploy"))
		sb.WriteString("\n")
	} else {
		// Render each honeypot card
		for _, hp := range m.honeypots {
			sb.WriteString(m.renderCard(hp))
			sb.WriteString("\n")
		}
	}

	// Instructions
	sb.WriteString("\n")
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	sb.WriteString(instructionsStyle.Render("Press 'q' or ESC to quit"))
	sb.WriteString("\n")

	return sb.String()
}

// renderCard renders a single honeypot card
func (m StatusModel) renderCard(hp Honeypot) string {
	// Status indicator
	var statusIndicator string
	var statusColor lipgloss.Color
	var statusText string

	switch hp.Status {
	case StatusActive:
		if m.blinkOn {
			statusIndicator = "●"
		} else {
			statusIndicator = "○"
		}
		statusColor = lipgloss.Color("48") // Green
		statusText = "ACTIVE"
	case StatusStopped:
		statusIndicator = "●"
		statusColor = lipgloss.Color("240") // Gray
		statusText = "STOPPED"
	case StatusError:
		if m.blinkOn {
			statusIndicator = "●"
		} else {
			statusIndicator = "○"
		}
		statusColor = lipgloss.Color("196") // Red
		statusText = "ERROR"
	}

	indicatorStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true)

	statusTextStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true)

	// Card styles
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statusColor).
		Padding(0, 2).
		Width(50)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	// Build card content
	var content strings.Builder

	// Header with status indicator
	header := fmt.Sprintf("%s %s  %s",
		indicatorStyle.Render(statusIndicator),
		valueStyle.Bold(true).Render(hp.Name),
		statusTextStyle.Render(statusText),
	)
	content.WriteString(header)
	content.WriteString("\n\n")

	// Details
	content.WriteString(labelStyle.Render("Profile:     "))
	content.WriteString(valueStyle.Render(hp.Profile))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Type:        "))
	content.WriteString(valueStyle.Render(hp.Type))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Server:      "))
	content.WriteString(valueStyle.Render(hp.ServerName))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Port:        "))
	content.WriteString(valueStyle.Render(fmt.Sprintf("%d", hp.Port)))
	content.WriteString("\n")

	if hp.Status == StatusActive && hp.Uptime != "" {
		content.WriteString(labelStyle.Render("Uptime:      "))
		uptimeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("48"))
		content.WriteString(uptimeStyle.Render(hp.Uptime))
		content.WriteString("\n")
	}

	if hp.Status == StatusError && hp.LastError != "" {
		content.WriteString(labelStyle.Render("Error:       "))
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		content.WriteString(errorStyle.Render(hp.LastError))
		content.WriteString("\n")
	}

	return cardStyle.Render(content.String())
}

// GetMockHoneypots returns mock honeypot data for testing
func GetMockHoneypots() []Honeypot {
	return []Honeypot{
		{
			Name:       "honeypot-prod-1",
			Profile:    "production",
			Type:       "classic",
			Status:     StatusActive,
			Uptime:     "2d 14h 32m",
			ServerName: "ssh-server-01",
			Port:       2222,
		},
		{
			Name:       "honeypot-dev-1",
			Profile:    "development",
			Type:       "ia",
			Status:     StatusStopped,
			ServerName: "test-server",
			Port:       2223,
		},
		{
			Name:       "honeypot-staging",
			Profile:    "staging",
			Type:       "classic",
			Status:     StatusError,
			LastError:  "Connection timeout",
			ServerName: "staging-srv",
			Port:       2224,
		},
	}
}
