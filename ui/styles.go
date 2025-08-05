package ui

import "github.com/charmbracelet/lipgloss"

var (
	Prompt = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	Input  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)
