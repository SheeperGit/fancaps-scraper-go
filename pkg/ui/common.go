package ui

import (
	"github.com/charmbracelet/lipgloss"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

var (
	SuccessStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Style for successful operations.
	HighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")) // Style for ongoing operations.
	ErrStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // Style for errors.
)

/* Returns the longest title name from titles `titles` and its length. */
func GetLongestTitle(titles []*types.Title) string {
	name := ""
	for _, title := range titles {
		if len(title.Name) > len(name) {
			name = title.Name
		}
	}

	return name
}
