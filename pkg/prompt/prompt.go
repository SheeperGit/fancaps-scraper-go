package prompt

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Style for help tips. */
var HelpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

/* Rendered text for search query prompt. */
var SearchHelpPrompt = HelpStyle.Render(
	"Type the name of a movie, TV series, or anime you'd like to search for.\n" +
		"(e.g., \"Predator\", \"Family Guy\", \"Hunter x Hunter\", etc.)\n" +
		"Tip: You can enter just part of a title to search.\n")

/*
Returns the text from the user prompt with prompt text `promptText`
and renders a help description `helpText`.
*/
func PromptUser(promptText, helpText string) string {
	fmt.Print(helpText)
	fmt.Print(promptText)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ""
}

/* Returns a list of titles with episodes selected by the user from titles `titles`. */
func SelectEpisodes(titles []types.Title) []types.Title {
	// TODO: Use err values to continue prompting.
	var newTitles []types.Title

	return newTitles
}
