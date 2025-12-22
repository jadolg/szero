package pkg

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var numStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
var nsStyle = lipgloss.NewStyle().Bold(true)

// N formats a number with highlighting for better visibility in logs
func N(n any) string {
	return numStyle.Render(fmt.Sprintf("%v", n))
}

// NS formats a namespace with bold text for better visibility in logs
func NS(ns string) string {
	return nsStyle.Render(ns)
}
