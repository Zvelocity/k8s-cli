package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Common styles used throughout the application
var (
	TitleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Bold(true).
			Foreground(lipgloss.Color("39"))

	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("69"))
)

// StylePodStatus returns a styled pod status string based on its status value
func StylePodStatus(status string) string {
	switch status {
	case "Running":
		return SuccessStyle.Render(status)
	case "Pending":
		return WarningStyle.Render(status)
	case "Failed", "Unknown", "Error":
		return ErrorStyle.Render(status)
	default:
		return status
	}
}
