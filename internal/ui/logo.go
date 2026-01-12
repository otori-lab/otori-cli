package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// GetLogo retourne le logo OTORI stylisé
func GetLogo() string {
	logoText := `
 ██████╗ ████████╗ ██████╗ ██████╗ ██╗
██╔═══██╗╚══██╔══╝██╔═══██╗██╔══██╗██║
██║   ██║   ██║   ██║   ██║██████╔╝██║
██║   ██║   ██║   ██║   ██║██╔══██╗██║
╚██████╔╝   ██║   ╚██████╔╝██║  ██║██║
 ╚═════╝    ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚═╝
                                      
`

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Bold(true)

	return style.Render(logoText)
}

// GetWelcomeMessage retourne un message de bienvenue stylisé
func GetWelcomeMessage() string {
	msg := "Configuration Assistant"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Italic(true)

	return style.Render(msg)
}
