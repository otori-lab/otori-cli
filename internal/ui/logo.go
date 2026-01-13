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

// GetProjectInfo retourne les informations du projet et les créateurs
func GetProjectInfo() string {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	credits := `ECE PARIS | PFE 2025

Créateurs:
  • Axelle ROUZIER
  • Wadih BEN ABDESSELEM
  • Jean-Nicolas NEYRET
  • Fabio OLIVEIRA
  • Paul-Alexandre FORTUNA
  • Mathis FOUCADE
`

	return infoStyle.Render(credits)
}
